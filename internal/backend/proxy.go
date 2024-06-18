package backend

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"cloud-lb/internal/config"

	"github.com/gorilla/websocket"
)

type ReverseProxy struct {
	config         *config.Config
	httpServer     *http.Server
	backendServers []*url.URL
	upgrader       websocket.Upgrader
	wsMux          sync.Mutex
	wsConns        map[*websocket.Conn]bool
	currentBackend int
}

func NewReverseProxy(config *config.Config) *ReverseProxy {
	backendServers := make([]*url.URL, len(config.BackendServers))
	for i, serverStr := range config.BackendServers {
		backendURL, err := url.Parse(serverStr)
		if err != nil {
			log.Fatalf("Invalid backend server URL: %s", serverStr)
		}
		backendServers[i] = backendURL
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	proxy := &ReverseProxy{
		config:         config,
		backendServers: backendServers,
		upgrader:       upgrader,
		wsConns:        make(map[*websocket.Conn]bool),
	}

	httpServer := &http.Server{
		Addr:    ":3000",
		Handler: http.HandlerFunc(proxy.HandleRequest),
	}
	proxy.httpServer = httpServer

	return proxy
}

func (p *ReverseProxy) loadBalance() *url.URL {
	switch p.config.LoadBalanceAlgo {
	case "roundrobin":
		p.currentBackend = (p.currentBackend + 1) % len(p.backendServers)
		return p.backendServers[p.currentBackend]
	case "random":
		return p.backendServers[rand.Intn(len(p.backendServers))]
	default:
		log.Fatalf("Unsupported load balancing algorithm: %s", p.config.LoadBalanceAlgo)
		return nil
	}
}

func (p *ReverseProxy) Serve() {
	log.Println("Starting reverse proxy server on :8080")
	if err := p.httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func (p *ReverseProxy) HandleRequest(w http.ResponseWriter, r *http.Request) {
	if websocket.IsWebSocketUpgrade(r) {
		p.HandleWebSocket(w, r)
		return
	}

	if !p.isAuthorized(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	target := p.loadBalance()
	proxy := httputil.NewSingleHostReverseProxy(target)
	r.Host = target.Host

	proxy.ServeHTTP(w, r)
}

func (p *ReverseProxy) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	if !p.isAuthorized(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	target := p.loadBalance()

	conn, err := p.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket connection: %v", err)
		return
	}

	p.wsMux.Lock()
	p.wsConns[conn] = true
	p.wsMux.Unlock()

	go p.forwardWebSocket(target, conn)
}

func (p *ReverseProxy) forwardWebSocket(target *url.URL, conn *websocket.Conn) {
	backendConn, _, err := websocket.DefaultDialer.Dial(target.String(), nil)
	if err != nil {
		log.Printf("Failed to connect to backend WebSocket server: %v", err)
		log.Printf("Debug: %v", backendConn)
		conn.Close()
		return
	}
	defer backendConn.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		err := p.forwardMessages(conn, backendConn)
		if err != nil {
			log.Printf("Error forwarding messages from client to server: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		err := p.forwardMessages(backendConn, conn)
		if err != nil {
			log.Printf("Error forwarding messages from server to client: %v", err)
		}
	}()

	wg.Wait()

	p.wsMux.Lock()
	delete(p.wsConns, conn)
	p.wsMux.Unlock()
}

func (p *ReverseProxy) forwardMessages(src, dest *websocket.Conn) error {
	for {
		messageType, payload, err := src.ReadMessage()
		if err != nil {
			return err
		}

		err = dest.WriteMessage(messageType, payload)
		if err != nil {
			return err
		}
	}
}

func (p *ReverseProxy) HandleGracefulShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Received signal: %s. Shutting down gracefully...", sig)

	if err := p.httpServer.Shutdown(context.Background()); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	timeout := time.After(4 * time.Second) // TODO: - Adjust the timeout as needed
	for {
		select {
		case <-timeout:
			log.Println("Timeout reached. Forcing shutdown.")
			os.Exit(1)
		default:
			p.wsMux.Lock()
			if len(p.wsConns) == 0 {
				p.wsMux.Unlock()
				log.Println("All WebSocket connections closed. Exiting gracefully.")
				os.Exit(0)
			}
			p.wsMux.Unlock()
			time.Sleep(100 * time.Millisecond)
		}
	}
}

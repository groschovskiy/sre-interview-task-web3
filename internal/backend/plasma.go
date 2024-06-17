package backend

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
)

func (p *ReverseProxy) loadBalance() *url.URL {
	// TODO: - Fix the logic of balancing
	return p.backendServers[0]
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

package backend

import (
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
)

type ReverseProxy struct {
	config         *Config
	httpServer     *http.Server
	backendServers []*url.URL
	upgrader       websocket.Upgrader
	wsMux          sync.Mutex
	wsConns        map[*websocket.Conn]bool
}

func NewReverseProxy(config *Config) *ReverseProxy {
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

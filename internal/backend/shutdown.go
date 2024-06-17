package backend

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func HandleGracefulShutdown(proxy *ReverseProxy) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Received signal: %s. Shutting down gracefully...", sig)

	if err := proxy.httpServer.Shutdown(context.Background()); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	timeout := time.After(1 * time.Second) // TODO: - Adjust the timeout as needed
	for {
		select {
		case <-timeout:
			log.Println("Timeout reached. Forcing shutdown.")
			os.Exit(1)
		default:
			proxy.wsMux.Lock()
			if len(proxy.wsConns) == 0 {
				proxy.wsMux.Unlock()
				log.Println("All WebSocket connections closed. Exiting gracefully.")
				os.Exit(0)
			}
			proxy.wsMux.Unlock()
			time.Sleep(100 * time.Millisecond)
		}
	}
}

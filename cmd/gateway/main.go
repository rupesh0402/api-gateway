package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rupesh0402/api-gateway/config"
	"github.com/rupesh0402/api-gateway/internal/server"
)

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()

	// REGISTER ROUTES (ONCE)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	srv := server.NewServer(
		cfg.Port,
		mux,
		cfg.ReadTimeout,
		cfg.WriteTimeout,
		cfg.IdleTimeout,
	)

	go srv.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}

	log.Println("server exited cleanly")
}

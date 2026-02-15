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
	"github.com/rupesh0402/api-gateway/internal/ratelimit"
	"github.com/rupesh0402/api-gateway/internal/server"
	"github.com/rupesh0402/api-gateway/internal/worker"
)

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()

	// Health route
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	limiter := ratelimit.NewLimiter(2, 5)
	pool := worker.NewPool(2, 1)

	mux.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
		clientIP := r.RemoteAddr

		if !limiter.Allow(clientIP) {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("Rate limit exceeded\n"))
			return
		}

		respChan := make(chan []byte)

		job := worker.Job{
			Response: respChan,
		}

		pool.Submit(job)

		result := <-respChan

		w.WriteHeader(http.StatusOK)
		w.Write(result)
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

	pool.Shutdown()

	log.Println("server exited cleanly")
}

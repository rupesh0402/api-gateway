package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rupesh0402/api-gateway/config"
	"github.com/rupesh0402/api-gateway/internal/auth"
	"github.com/rupesh0402/api-gateway/internal/proxy"
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

	rp, err := proxy.NewReverseProxy("http://localhost:9001")
	if err != nil {
		log.Fatal(err)
	}

	jwtManager := auth.NewJWTManager("mysecretkey")

	limiter := ratelimit.NewLimiter(2, 5)
	pool := worker.NewPool(2, 1)

	mux.Handle("/api/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// JWT auth
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		tokenString, err := auth.ExtractBearerToken(authHeader)
		if err != nil {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		_, err = jwtManager.Validate(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Rate limit
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		if !limiter.Allow(host) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Forward request
		rp.ServeHTTP(w, r)
	}))

	mux.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Missing token\n"))
			return
		}

		tokenString, err := auth.ExtractBearerToken(authHeader)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Invalid token format\n"))
			return
		}

		_, err = jwtManager.Validate(tokenString)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Invalid or expired token\n"))
			return
		}

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

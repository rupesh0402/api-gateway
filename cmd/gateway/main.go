package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/rupesh0402/api-gateway/config"
	"github.com/rupesh0402/api-gateway/internal/auth"
	"github.com/rupesh0402/api-gateway/internal/logging"
	"github.com/rupesh0402/api-gateway/internal/metrics"
	"github.com/rupesh0402/api-gateway/internal/proxy"
	"github.com/rupesh0402/api-gateway/internal/ratelimit"
	"github.com/rupesh0402/api-gateway/internal/server"
	"github.com/rupesh0402/api-gateway/internal/worker"
)

func main() {
	cfg := config.Load()

	logging.Init()
	metrics.Init()

	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Prometheus
	mux.Handle("/metrics", metrics.MetricsHandler())

	// Reverse proxy
	rp, err := proxy.NewReverseProxy("http://localhost:9001")
	if err != nil {
		logging.Logger.Fatal(err)
	}

	// JWT (now from config)
	jwtManager := auth.NewJWTManager(cfg.JWTSecret)

	// Rate limiter
	limiter := ratelimit.NewLimiter(2, 5)

	// Worker pool (used only for /process demo)
	pool := worker.NewPool(2, 1)

	// =========================
	// PROXY ROUTE
	// =========================

	mux.Handle("/api/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		status := "200"

		// Extract client IP properly
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}

		// JWT Auth
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			status = "401"
			http.Error(w, "Missing token", http.StatusUnauthorized)
			trackAndLog(r, status, start)
			return
		}

		tokenString, err := auth.ExtractBearerToken(authHeader)
		if err != nil {
			status = "401"
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			trackAndLog(r, status, start)
			return
		}

		_, err = jwtManager.Validate(tokenString)
		if err != nil {
			status = "401"
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			trackAndLog(r, status, start)
			return
		}

		// Rate limiting
		if !limiter.Allow(host) {
			status = "429"
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			trackAndLog(r, status, start)
			return
		}

		// Forward to backend
		rp.ServeHTTP(w, r)

		trackAndLog(r, status, start)
	}))

	// =========================
	// WORKER DEMO ROUTE
	// =========================

	mux.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		status := "200"

		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			status = "401"
			http.Error(w, "Missing token", http.StatusUnauthorized)
			trackAndLog(r, status, start)
			return
		}

		tokenString, err := auth.ExtractBearerToken(authHeader)
		if err != nil {
			status = "401"
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			trackAndLog(r, status, start)
			return
		}

		_, err = jwtManager.Validate(tokenString)
		if err != nil {
			status = "401"
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			trackAndLog(r, status, start)
			return
		}

		if !limiter.Allow(host) {
			status = "429"
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			trackAndLog(r, status, start)
			return
		}

		respChan := make(chan []byte, 1)

		job := worker.Job{
			Response: respChan,
		}

		pool.Submit(job)

		result := <-respChan

		w.WriteHeader(http.StatusOK)
		w.Write(result)

		trackAndLog(r, status, start)
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
		logging.Logger.Fatal(err)
	}

	pool.Shutdown()

	logging.Logger.Info("server exited cleanly")
}

// ==========================================
// Shared tracking + structured logging
// ==========================================

func trackAndLog(r *http.Request, status string, start time.Time) {
	duration := time.Since(start)

	metrics.TrackRequest(r.Method, r.URL.Path, status, duration)

	logging.Logger.WithFields(logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"status":     status,
		"latency_ms": duration.Milliseconds(),
	}).Info("request completed")
}

# 🚀 High-Throughput API Gateway (Go)

A production-grade API Gateway built in Go, designed to handle high
throughput with low latency while providing strong authentication, rate
limiting, observability, and graceful shutdown support.

This project demonstrates real-world backend engineering practices used
in scalable distributed systems.

------------------------------------------------------------------------

## 🧠 Overview

This API Gateway serves as a single entry point for backend services and
provides:

-   🔐 JWT Authentication\
-   🚦 Leaky Bucket Rate Limiting\
-   🌐 Reverse Proxy to Upstream Services\
-   📊 Prometheus Metrics\
-   🧾 Structured JSON Logging\
-   🛑 Graceful Shutdown\
-   ⏱ Upstream Timeout Protection

------------------------------------------------------------------------

## 🏗 Architecture

Client → JWT Authentication → Rate Limiting → Reverse Proxy → Backend
Services → Metrics + Logs

------------------------------------------------------------------------

## 📂 Project Structure

    api-gateway/
    ├── cmd/
    │   └── gateway/
    │       └── main.go
    ├── internal/
    │   ├── auth/         # JWT validation
    │   ├── config/       # Configuration loader
    │   ├── logging/      # Structured logging setup
    │   ├── metrics/      # Prometheus instrumentation
    │   ├── proxy/        # Reverse proxy logic
    │   ├── ratelimit/    # Leaky bucket implementation
    │   ├── server/       # HTTP server wrapper
    │   └── worker/       # Worker pool (demo route)
    ├── go.mod
    └── README.md

------------------------------------------------------------------------

## ✨ Features

### 🔐 JWT Authentication

-   Validates Bearer tokens
-   Verifies signature using HMAC
-   Enforces expiration
-   Returns 401 Unauthorized for invalid tokens

### 🚦 Rate Limiting (Leaky Bucket)

-   Per-IP traffic control
-   Burst handling
-   Smooth request shaping
-   Returns 429 Too Many Requests when exceeded

### 🌐 Reverse Proxy

-   Forwards HTTP method, headers, body, and query parameters
-   Tuned http.Transport for performance
-   Upstream timeout handling
-   Returns 504 Gateway Timeout on slow upstreams

### 📊 Metrics (Prometheus)

Exposed via `/metrics` endpoint: - gateway_requests_total -
gateway_request_duration_seconds

### 🧾 Structured Logging

JSON formatted logs containing: - Method - Path - Status code - Latency
(ms)

### 🛑 Graceful Shutdown

-   Handles SIGINT and SIGTERM
-   Stops accepting new requests
-   Drains in-flight requests
-   Cleans up worker pool

------------------------------------------------------------------------

## ⚙️ Configuration

Environment Variables:

JWT_SECRET=your_secret_key

Default Port: :8000

------------------------------------------------------------------------

## ▶️ Running the Project

### 1️⃣ Start Example Backend Service

go run backend.go

Runs on: http://localhost:9001

### 2️⃣ Start API Gateway

JWT_SECRET=mysecret go run ./cmd/gateway

Runs on: http://localhost:8000

------------------------------------------------------------------------

## 🔍 Testing

### Health Check

curl http://localhost:8000/health

### Authenticated Request

curl -H "Authorization: Bearer `<TOKEN>`{=html}"
http://localhost:8000/api/hello

### Rate Limit Test

ab -n 20 -c 10 -H "Authorization: Bearer `<TOKEN>`{=html}"
http://localhost:8000/api/hello

------------------------------------------------------------------------

## 📈 Performance Targets

  Metric        Target
  ------------- ------------
  Throughput    ≥ 5000 RPS
  p95 Latency   \< 10ms
  Error Rate    \< 0.1%

------------------------------------------------------------------------

## 🛡 Security Notes

-   JWT secret is loaded via environment variables.
-   No hardcoded credentials.
-   Upstream timeouts prevent cascading failures.
-   Rate limiting mitigates abuse.

------------------------------------------------------------------------
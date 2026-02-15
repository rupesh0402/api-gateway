package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello from backend service\n"))
	})

	log.Println("Backend running on :9001")
	http.ListenAndServe(":9001", nil)
}

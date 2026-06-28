package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <port-file>", os.Args[0])
	}
	portFile := os.Args[1]

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	if err := os.MkdirAll(filepath.Dir(portFile), 0o755); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(portFile, []byte(fmt.Sprintf("%d", addr.Port)), 0o644); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	serve := func(w http.ResponseWriter, r *http.Request, expectedAuth string, response string) {
		if r.URL.RawQuery != "download=1" {
			http.Error(w, "unexpected query", http.StatusBadRequest)
			return
		}
		if r.Header.Get("Authorization") != expectedAuth {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		_, _ = fmt.Fprint(w, response)
	}
	mux.HandleFunc("/data.txt", func(w http.ResponseWriter, r *http.Request) {
		serve(w, r, "Bearer secret-token", "remote-data")
	})
	mux.HandleFunc("/data-none.txt", func(w http.ResponseWriter, r *http.Request) {
		serve(w, r, "", "none-data")
	})
	mux.HandleFunc("/data-bearer-env.txt", func(w http.ResponseWriter, r *http.Request) {
		serve(w, r, "Bearer secret-token", "bearer-env-data")
	})
	mux.HandleFunc("/data-bearer-inline.txt", func(w http.ResponseWriter, r *http.Request) {
		serve(w, r, "Bearer inline-secret-token", "bearer-inline-data")
	})
	mux.HandleFunc("/data-basic-env.txt", func(w http.ResponseWriter, r *http.Request) {
		serve(w, r, "Basic ZW52LXVzZXI6ZW52LXBhc3N3b3Jk", "basic-env-data")
	})
	mux.HandleFunc("/data-basic-inline.txt", func(w http.ResponseWriter, r *http.Request) {
		serve(w, r, "Basic aW5saW5lLXVzZXI6aW5saW5lLXBhc3N3b3Jk", "basic-inline-data")
	})

	server := &http.Server{Handler: mux}
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-done
		_ = server.Close()
	}()

	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"time"

	handler "github.com/lin-snow/ech0/api"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/maintenance", handler.Handler)
	mux.HandleFunc("/", handler.Ech0)
	server := http.Server{Addr: net.JoinHostPort("", port), Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	log.Fatal(server.ListenAndServe())
}

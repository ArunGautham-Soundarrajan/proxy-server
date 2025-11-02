package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/ArunGautham-Soundarrajan/proxy-server/internal/cache"
	"github.com/ArunGautham-Soundarrajan/proxy-server/internal/proxy"
)

// cache is a simple in-memory map to store cached responses, protected by
// a mutex for concurrent access.
// var cacheTTL = time.Minute * 5

// logger is a package-level logger used by handlers. It is initialized in
// main() to write human-readable text to stdout.
var logger *slog.Logger

func main() {
	lruCache := cache.NewLRUCache(100) // capacity 100
	handler := proxy.NewProxyHandler(lruCache)

	http.Handle("/", handler)

	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("Proxy Server is listening at port 8080")

	http.ListenAndServe(":8080", nil)
}

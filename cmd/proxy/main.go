package main

import (
	"log"
	"net/http"

	"github.com/ArunGautham-Soundarrajan/proxy-server/internal/cache"
	"github.com/ArunGautham-Soundarrajan/proxy-server/internal/proxy"
)

// cache is a simple in-memory map to store cached responses, protected by
// a mutex for concurrent access.
// var cacheTTL = time.Minute * 5

func main() {
	lruCache := cache.NewLRUCache(100) // capacity 100
	handler := proxy.NewProxyHandler(lruCache)

	http.Handle("/", proxy.LoggingMiddleware(handler))

	log.Println("Proxy Server is listening at port 8080")
	err := http.ListenAndServe(":8080", nil)
	log.Fatal(err)
}

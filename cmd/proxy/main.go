package main

import (
	"log"
	"net/http"

	"github.com/ArunGautham-Soundarrajan/proxy-server/internal/cache"
	"github.com/ArunGautham-Soundarrajan/proxy-server/internal/config"
	"github.com/ArunGautham-Soundarrajan/proxy-server/internal/proxy"
)

func main() {
	lruCache := cache.NewLRUCache(config.CAPACITY, config.CacheTTL) // capacity 100
	handler := proxy.NewProxyHandler(lruCache)

	http.Handle("/", proxy.LoggingMiddleware(handler))

	log.Println("Proxy Server is listening at port 8080")
	err := http.ListenAndServe(":8080", nil)
	log.Fatal(err)
}

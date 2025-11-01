package main

import (
	"container/list"
	"net/http"
	"testing"
	"time"
)

func newTestCache(capacity int) *LRUcache {
	return &LRUcache{
		cache:      make(map[string]*list.Element),
		linkedlist: list.New(),
		capacity:   capacity,
	}
}

func TestPutAndGet(t *testing.T) {
	cache := newTestCache(2)
	resp := cacheResponse{
		key:        "a",
		statusCode: 200,
		header:     http.Header{"Content-Type": []string{"text/plain"}},
		body:       []byte("hello"),
		cachedAt:   time.Now(),
	}

	cache.put("a", resp)

	got, ok := cache.get("a")
	if !ok {
		t.Fatalf("expected key 'a' to exist")
	}

	if string(got.body) != "hello" {
		t.Errorf("expected body 'hello', got %q", got.body)
	}
}

func TestGetMissingKey(t *testing.T) {
	cache := newTestCache(1)
	_, ok := cache.get("missing")
	if ok {
		t.Fatal("expected missing key to return false")
	}
}

func TestEviction(t *testing.T) {
	cache := newTestCache(2)

	cache.put("a", cacheResponse{key: "a"})
	cache.put("b", cacheResponse{key: "b"})
	cache.put("c", cacheResponse{key: "c"}) // this should evict "a"

	if _, ok := cache.get("a"); ok {
		t.Errorf("expected 'a' to be evicted")
	}
	if _, ok := cache.get("b"); !ok {
		t.Errorf("expected 'b' to exist")
	}
	if _, ok := cache.get("c"); !ok {
		t.Errorf("expected 'c' to exist")
	}
}

func TestRecentlyUsedNotEvicted(t *testing.T) {
	cache := newTestCache(2)

	cache.put("a", cacheResponse{key: "a"})
	cache.put("b", cacheResponse{key: "b"})

	// Access "a" to make it MRU
	cache.get("a")

	// Add new key, should evict "b"
	cache.put("c", cacheResponse{key: "c"})

	if _, ok := cache.get("b"); ok {
		t.Errorf("expected 'b' to be evicted")
	}
	if _, ok := cache.get("a"); !ok {
		t.Errorf("expected 'a' to remain")
	}
	if _, ok := cache.get("c"); !ok {
		t.Errorf("expected 'c' to exist")
	}
}

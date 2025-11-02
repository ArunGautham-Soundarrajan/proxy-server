package cache

import (
	"bufio"
	"bytes"
	"container/list"
	"io"
	"net/http"
	"strings"
	"testing"
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

	// Create a test HTTP response
	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"text/plain"}},
		Body:       io.NopCloser(strings.NewReader("hello")),
	}

	err := cache.Put("a", resp)
	if err != nil {
		t.Fatalf("failed to put: %v", err)
	}

	got, ok := cache.Get("a")
	if !ok {
		t.Fatalf("expected key 'a' to exist")
	}

	// Convert response bytes back to http.Response for checking
	response, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(got.Response)), nil)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Read the body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	if string(body) != "hello" {
		t.Errorf("expected body 'hello', got %q", body)
	}
}

func TestGetMissingKey(t *testing.T) {
	cache := newTestCache(1)
	_, ok := cache.Get("missing")
	if ok {
		t.Fatal("expected missing key to return false")
	}
}

func TestEviction(t *testing.T) {
	cache := newTestCache(2)

	// Create test responses
	respA := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("a")),
	}
	respB := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("b")),
	}
	respC := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("c")),
	}

	cache.Put("a", respA)
	cache.Put("b", respB)
	cache.Put("c", respC) // this should evict "a"

	if _, ok := cache.Get("a"); ok {
		t.Errorf("expected 'a' to be evicted")
	}
	if _, ok := cache.Get("b"); !ok {
		t.Errorf("expected 'b' to exist")
	}
	if _, ok := cache.Get("c"); !ok {
		t.Errorf("expected 'c' to exist")
	}
}

func TestRecentlyUsedNotEvicted(t *testing.T) {
	cache := newTestCache(2)

	// Create test responses
	respA := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("a")),
	}
	respB := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("b")),
	}
	respC := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("c")),
	}

	cache.Put("a", respA)
	cache.Put("b", respB)

	// Access "a" to make it MRU
	cache.Get("a")

	// Add new key, should evict "b"
	cache.Put("c", respC)

	if _, ok := cache.Get("b"); ok {
		t.Errorf("expected 'b' to be evicted")
	}
	if _, ok := cache.Get("a"); !ok {
		t.Errorf("expected 'a' to remain")
	}
	if _, ok := cache.Get("c"); !ok {
		t.Errorf("expected 'c' to exist")
	}
}

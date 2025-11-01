package main

import (
	"container/list"
	"net/http"
	"sync"
	"time"
)

// cacheResponse represents a cached HTTP response including status code,
// headers, body, and the time it was cached.
type cacheResponse struct {
	key        string
	statusCode int
	header     http.Header
	body       []byte
	cachedAt   time.Time
}

type LRUcache struct {
	cache      map[string]*list.Element
	linkedlist *list.List
	capacity   int
	mu         sync.Mutex
}

func (c *LRUcache) get(k string) (*cacheResponse, bool) {

	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.cache[k]; ok {

		c.linkedlist.MoveToFront(elem)
		return elem.Value.(*cacheResponse), true
	}
	return nil, false
}

func (c *LRUcache) put(k string, resp cacheResponse) {

	c.mu.Lock()
	defer c.mu.Unlock()
	elem := c.linkedlist.PushFront(&resp)
	c.cache[k] = elem

	if c.linkedlist.Len() > c.capacity {
		tail := c.linkedlist.Back()
		c.linkedlist.Remove(tail)
		delete(c.cache, tail.Value.(*cacheResponse).key)
	}

}

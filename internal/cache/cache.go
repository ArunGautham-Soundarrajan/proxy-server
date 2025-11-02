package cache

import (
	"container/list"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"
)

// cacheResponse represents a cached HTTP response including status code,
// headers, body, and the time it was cached.
type cacheResponse struct {
	Key      string
	Response []byte
	CachedAt time.Time
}

// LRU cache stores the map of cache [key]:[cacheResponse](pointer to list.Element)
// We use list.List which is go's implementation of Linkedlist to keep track of
// Most used and lead used item.
// Capacity is used to set the maximum capacity of the cache
type LRUcache struct {
	cache      map[string]*list.Element
	linkedlist *list.List
	capacity   int
	ttl        time.Duration
	mu         sync.Mutex
}

// Get functions gets the k which is the `GET:http://example.com/`
// If the key exists, will return the cached response
func (c *LRUcache) Get(k string) (*cacheResponse, bool) {

	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.cache[k]; ok {

		if time.Since(elem.Value.(*cacheResponse).CachedAt) >= c.ttl {
			c.linkedlist.Remove(elem)
			delete(c.cache, elem.Value.(*cacheResponse).Key)
			return nil, false
		}

		// if the key exists, move the element to the front
		c.linkedlist.MoveToFront(elem)
		return elem.Value.(*cacheResponse), true
	}
	return nil, false
}

// Add an item to the cache if it doesn't exist already
// Get the key for the cache and the response to store
func (c *LRUcache) Put(k string, resp *http.Response) error {

	c.mu.Lock()
	defer c.mu.Unlock()

	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return err
	}

	// Push the stored element to most recently used
	elem := c.linkedlist.PushFront(&cacheResponse{
		Key:      k,
		Response: dump,
		CachedAt: time.Now(),
	})
	c.cache[k] = elem

	// Remove the tail if more than capacity after adding
	// and delete the entry from cache map too
	if c.linkedlist.Len() > c.capacity {
		tail := c.linkedlist.Back()
		c.linkedlist.Remove(tail)
		delete(c.cache, tail.Value.(*cacheResponse).Key)
	}

	return nil

}

func NewLRUCache(capacity int, ttl time.Duration) *LRUcache {
	return &LRUcache{
		cache:      make(map[string]*list.Element),
		linkedlist: list.New(),
		capacity:   capacity,
		ttl:        ttl,
	}
}

package main

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"
)

// cacheResponse represents a cached HTTP response including status code,
// headers, body, and the time it was cached.
type cacheResponse struct {
	statusCode int
	header     http.Header
	body       []byte
	cachedAt   time.Time
}

// cache is a simple in-memory map to store cached responses, protected by
// a mutex for concurrent access.
var cache = make(map[string]cacheResponse)
var cacheMutex sync.RWMutex
var cacheTTL = time.Minute * 5

// Check if the key exists in cache and return the caches response
// If the cache exceed the cache Time to Live, delete the cache and return false
func getCachedResponse(key string) (cacheResponse, bool) {
	cacheMutex.RLock()
	resp, found := cache[key]
	cacheMutex.RUnlock()

	if found && time.Since(resp.cachedAt) >= cacheTTL {
		cacheMutex.Lock()
		delete(cache, key)
		cacheMutex.Unlock()
		found = false
	}
	return resp, found
}

// Function to serve the response if it exists in cache
// Construct headers, write status code and body to response
func serveCachedResponse(w http.ResponseWriter, cr cacheResponse) {

	for header, values := range cr.header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}
	w.WriteHeader(cr.statusCode)
	w.Write(cr.body)

}

// Function to set the response to cache
func setCachedResponse(key string, resp http.Response) {

	bodyBytes, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	cacheMutex.Lock()
	cache[key] = cacheResponse{
		statusCode: resp.StatusCode,
		header:     resp.Header.Clone(),
		body:       bodyBytes,
		cachedAt:   time.Now(),
	}
	cacheMutex.Unlock()

	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
}

// constructHeaders copies headers from the source header map `src` to the
// destination request `dst`, skipping hop-by-hop headers that must not be
// forwarded by proxies (see RFC 2616 section 13.5.1 and common practice).
func constructHeaders(dst *http.Request, src http.Header) {

	// hop-by-hop headers that should not be forwarded
	headersToAvoid := map[string]bool{
		"connection":          true,
		"keep-alive":          true,
		"proxy-authenticate":  true,
		"proxy-authorization": true,
		"te":                  true,
		"trailer":             true,
		"transfer-encoding":   true,
		"upgrade":             true,
	}

	// Iterate over all headers and copy the ones that are safe to forward.
	for key, values := range src {
		if !headersToAvoid[key] {
			for _, value := range values {
				dst.Header.Add(key, value)
			}
		}
	}

}

// forwardRequest builds and sends an HTTP request based on the incoming
// request `r` and returns the upstream response. The function creates a new
// `http.Client`, constructs the target URL from the original request, copies
// safe headers, and performs the request. On error it logs and returns a
// synthetic `http.Response` with a 500 status code.
//
// Note: The returned http.Response is copied by value. Callers should treat
// the returned response carefully and should not assume the body is rewindable.
func forwardRequest(r *http.Request) http.Response {
	client := &http.Client{}

	// Build the absolute URL to forward to. This assumes the incoming request
	// already contains a scheme (e.g., from a Reverse Proxy) and a Host.
	url := r.URL.Scheme + "://" + r.Host + r.URL.RequestURI()
	req, err := http.NewRequest(r.Method, url, r.Body)

	if err != nil {
		logger.Error("Error Constructing Req",
			"method", r.Method,
			"url", r.URL.String(),
			"error", err.Error(),
		)
		return http.Response{StatusCode: http.StatusInternalServerError, Request: r}
	}

	// Copy headers from the incoming request to the new outgoing request
	// while skipping hop-by-hop headers.
	constructHeaders(req, r.Header)

	// Perform the request to the upstream server
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Error Making Req",
			"method", r.Method,
			"url", r.URL.String(),
			"error", err.Error(),
		)
		return http.Response{StatusCode: http.StatusInternalServerError, Request: r}
	}
	return *resp
}

// constructResponse writes headers and body from `resp` to the original
// ResponseWriter `w`. It copies all response headers then writes the status
// code and streams the response body to the client.
func constructResponse(w http.ResponseWriter, resp *http.Response) {

	// Copy all upstream response headers to the downstream response.
	for header, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}

	// Write status code and stream the body.
	w.WriteHeader(resp.StatusCode)
	defer resp.Body.Close()
	io.Copy(w, resp.Body)
}

// proxy is the HTTP handler for incoming requests. It logs the request,
// forwards it to the destination using `forwardRequest`, and then writes the
// upstream response back to the client with `constructResponse`.
func proxy(w http.ResponseWriter, r *http.Request) {

	logger.Info("Incoming Request",
		"url", r.URL.String(),
		"method", r.Method,
		"remote_addr", r.RemoteAddr,
		"user_agent", r.UserAgent(),
	)
	startTime := time.Now()

	// Construct key for the cache
	// KEY: `GET:http://example.com/`
	key := r.Method + ":" + r.URL.String()

	// Check if the key exists in cache and can be served
	if cachedResp, found := getCachedResponse(key); found {
		// Serve from cache
		logger.Info("Serving from Cache",
			"url", r.URL.String(),
			"status", cachedResp.statusCode,
		)
		serveCachedResponse(w, cachedResp)

	} else {

		resp := forwardRequest(r)

		// Set the cache if the method is GET and status is ok
		if r.Method == http.MethodGet && resp.StatusCode == http.StatusOK {
			logger.Info("Cache Miss",
				"url", r.URL.String(),
				"method", r.Method,
			)
			setCachedResponse(key, resp)
		}
		constructResponse(w, &resp)
	}
	duration := time.Since(startTime)
	logger.Info("Response Sent",
		"url", r.URL.String(),
		"duration_ms", duration.Milliseconds(),
	)
}

// logger is a package-level logger used by handlers. It is initialized in
// main() to write human-readable text to stdout.
var logger *slog.Logger

func main() {
	http.HandleFunc("/", proxy)
	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("Proxy Server is listening at port 8080")
	http.ListenAndServe(":8080", nil)
}

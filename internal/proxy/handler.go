package proxy

import (
	"bufio"
	"bytes"
	"io"
	"net/http"

	"github.com/ArunGautham-Soundarrajan/proxy-server/internal/cache"
)

type ProxyHandler struct {
	Cache *cache.LRUcache
}

// Constructor for dependency injection
func NewProxyHandler(c *cache.LRUcache) *ProxyHandler {
	return &ProxyHandler{Cache: c}
}

func (p *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Construct key for the cache
	// KEY: `GET:http://example.com/`
	key := r.Method + ":" + r.URL.String()

	// Check if the key exists in cache and can be served
	if cachedResp, found := p.Cache.Get(key); found {
		resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(cachedResp.Response)), r)
		if err == nil {
			constructResponse(w, resp)
			return
		}

	} else {
		resp := forwardRequest(r)

		// Set the cache if the method is GET and status is ok
		if r.Method == http.MethodGet && resp.StatusCode == http.StatusOK {
			p.Cache.Put(key, &resp)
		}
		constructResponse(w, &resp)
	}
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
		return http.Response{StatusCode: http.StatusInternalServerError, Request: r}
	}

	// Copy headers from the incoming request to the new outgoing request
	// while skipping hop-by-hop headers.
	constructHeaders(req, r.Header)

	// Perform the request to the upstream server
	resp, err := client.Do(req)
	if err != nil {
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

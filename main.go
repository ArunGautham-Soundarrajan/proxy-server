package main

import (
	"io"
	"log/slog"
	"net/http"
	"os"
)

func constructHeaders(dst *http.Request, src http.Header) error {

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
	for key, values := range src {
		if !headersToAvoid[key] {
			for _, value := range values {
				dst.Header.Add(key, value)
			}
		}
	}

	return nil
}

func forwardRequest(r *http.Request) http.Response {
	client := &http.Client{}

	url := r.URL.Scheme + "://" + r.Host + r.URL.RequestURI()
	req, err := http.NewRequest(r.Method, url, r.Body)

	if err != nil {
		logger.Error("Error Constructing Req", "error", err.Error())
		return http.Response{StatusCode: http.StatusInternalServerError, Request: r}
	}

	err = constructHeaders(req, r.Header)
	if err != nil {
		logger.Error("Error Constructing Headers", "error", err.Error())
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Error Making Req", "error", err.Error())
		return http.Response{StatusCode: http.StatusInternalServerError, Request: r}
	}
	return *resp
}

func constructResponse(w http.ResponseWriter, resp *http.Response) {

	for header, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func proxy(w http.ResponseWriter, r *http.Request) {

	logger.Info("Incoming Request", "url", r.URL.String())
	resp := forwardRequest(r)
	constructResponse(w, &resp)
	logger.Info("Response Sent", "url", r.URL.String())

}

var logger *slog.Logger

func main() {
	http.HandleFunc("/", proxy)
	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("Proxy Server is listening at port 8080")
	http.ListenAndServe(":8080", nil)
}

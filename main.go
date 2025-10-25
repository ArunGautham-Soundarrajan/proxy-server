package main

import (
	"fmt"
	"net/http"
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
		fmt.Print(err.Error())
		return http.Response{StatusCode: http.StatusInternalServerError, Request: r}
	}

	err = constructHeaders(req, r.Header)
	if err != nil {
		fmt.Print(err.Error())
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err.Error())
		return http.Response{StatusCode: http.StatusInternalServerError, Request: r}
	}
	return *resp
}

func proxy(w http.ResponseWriter, r *http.Request) {

	resp := forwardRequest(r)
	resp.Write(w)

}

func main() {
	http.HandleFunc("/", proxy)
	http.ListenAndServe(":8080", nil)
}

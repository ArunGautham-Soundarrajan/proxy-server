package proxy

import (
	"log"
	"net/http"
	"time"
)

func LoggingMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		url := r.Method + ":" + r.URL.String()

		log.Println(url, "Processing req")
		start := time.Now()

		next.ServeHTTP(w, r)

		dur := time.Since(start)
		log.Printf("%s Time Taken: %.2f ms", url, float64(dur.Microseconds())/1000)
	})
}

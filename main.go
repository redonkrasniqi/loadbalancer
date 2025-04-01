package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

var (
	backends = []string{
		"http://backend1.local",
		"http://backend2.local",
		"http://backend3.local",
	}
	counter uint32
)

func getNextServer(role string) string {
	if role == "admin" {
		return backends[0]
	}
	index := atomic.AddUint32(&counter, 1) % uint32(len(backends))
	return backends[index]
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	role := r.Header.Get("Role")
	target := getNextServer(role)

	backendURL, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(backendURL)

	log.Printf("Proxying request to: %s (Role: %s)\n", target, role)
	proxy.ServeHTTP(w, r)
}

func main() {
	http.Handle("/", JwtMiddleware(http.HandlerFunc(handleRequest)))

	log.Println("Load Balancer running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

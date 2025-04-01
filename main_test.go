// filepath: main_test.go
package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleRequest(t *testing.T) {
	// Save original backends and restore after test
	originalBackends := backends
	defer func() { backends = originalBackends }()

	// Setup test backends
	testServer1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("backend1"))
	}))
	testServer2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("backend2"))
	}))
	defer testServer1.Close()
	defer testServer2.Close()

	backends = []string{
		testServer1.URL,
		testServer2.URL,
	}

	tests := []struct {
		name          string
		role          string
		expectedHost  string
		expectedLogs  string
	}{
		{
			name:         "Admin role routes to first backend",
			role:         "admin",
			expectedHost: testServer1.URL,
			expectedLogs: "Proxying request to: " + testServer1.URL,
		},
		{
			name:         "Non-admin role gets load balanced",
			role:         "user",
			expectedHost: testServer1.URL, // First request with counter will go to first server
			expectedLogs: "Proxying request to: " + testServer1.URL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture log output
			var logBuffer bytes.Buffer
			log.SetOutput(&logBuffer)
			defer log.SetOutput(log.Writer())

			// Create test request
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Role", tt.role)
			
			// Create response recorder
			rr := httptest.NewRecorder()

			// Handle the request
			handleRequest(rr, req)

			// Verify response
			if status := rr.Code; status != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}

			// Verify logs contain expected message
			if !strings.Contains(logBuffer.String(), tt.expectedLogs) {
				t.Errorf("logs don't contain expected message\ngot: %s\nwant to contain: %s",
					logBuffer.String(), tt.expectedLogs)
			}
		})
	}
}

func TestLoadBalancing(t *testing.T) {
	// Reset counter
	counter = 0

	// Setup test servers
	responses := []string{"backend1", "backend2", "backend3"}
	servers := make([]*httptest.Server, len(responses))
	backends = make([]string, len(responses))

	for i, resp := range responses {
		resp := resp
		servers[i] = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(resp))
		}))
		backends[i] = servers[i].URL
		defer servers[i].Close()
	}

	// Test multiple requests are distributed
	role := "user"
	for i := 0; i < 6; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Role", role)
		rr := httptest.NewRecorder()

		handleRequest(rr, req)

		expectedBackend := backends[i%len(backends)]
		if !strings.Contains(rr.Body.String(), responses[i%len(responses)]) {
			t.Errorf("request %d: expected response from %s, got %s",
				i, expectedBackend, rr.Body.String())
		}
	}
}
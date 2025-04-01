package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetNextServer(t *testing.T) {
	adminTarget := GetNextServer("admin")
	if adminTarget != "http://backend1.local" {
		t.Errorf("Expected 'http://backend1.local' for admin, got %s", adminTarget)
	}

	userTarget1 := GetNextServer("user")
	userTarget2 := GetNextServer("user")
	if userTarget1 == userTarget2 {
		t.Errorf("Expected different backends for different user requests, got %s and %s", userTarget1, userTarget2)
	}
}

func TestHandleRequest(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Role", "admin")
	resp := httptest.NewRecorder()

	HandleRequest(resp, req)

	if resp.Code != http.StatusBadGateway {
		t.Errorf("Expected status Bad Gateway (502) due to no real backends, got %v", resp.Code)
	}
}

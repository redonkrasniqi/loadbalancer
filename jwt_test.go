package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func generateTestJWT(role string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"role": role,
		"exp":  time.Now().Add(time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString(secretKey)
	return tokenString
}

func TestParseJWT_ValidToken(t *testing.T) {
	token := generateTestJWT("admin")
	role, err := parseJWT(token)
	if err != nil || role != "admin" {
		t.Errorf("Expected role 'admin', got %v, error: %v", role, err)
	}
}

func TestParseJWT_InvalidToken(t *testing.T) {
	_, err := parseJWT("invalid.token.here")
	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}
}

func TestJwtMiddleware_ValidToken(t *testing.T) {
	validToken := generateTestJWT("user")
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	resp := httptest.NewRecorder()
	handler := JwtMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Code)
	}
}

func TestJwtMiddleware_MissingToken(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	resp := httptest.NewRecorder()
	handler := JwtMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(resp, req)
	if resp.Code != http.StatusUnauthorized {
		t.Errorf("Expected status Unauthorized, got %v", resp.Code)
	}
}

func TestJwtMiddleware_InvalidToken(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid.token")
	resp := httptest.NewRecorder()
	handler := JwtMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(resp, req)
	if resp.Code != http.StatusUnauthorized {
		t.Errorf("Expected status Unauthorized, got %v", resp.Code)
	}
}

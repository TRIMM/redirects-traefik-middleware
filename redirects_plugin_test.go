package redirects_traefik_middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServeHTTP_Redirection(t *testing.T) {
	nextHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	// Create a test configuration for the middleware
	config := &Config{
		RedirectsAppURL: "http://localhost:8081",
	}

	// Initialize the middleware with the test configuration
	handler, err := New(context.Background(), nextHandler, config, "test-middleware")
	if err != nil {
		t.Fatalf("Failed to create middleware: %v", err)
	}

	req, err := http.NewRequest("GET", "http://demo.localhost/product/iphone", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusFound)
	}

	// Check the Location header for the redirect URL
	expectedRedirectURL := "http://demo.localhost/new-product/iphone"
	location, _ := rr.Result().Location()
	if location.String() != expectedRedirectURL {
		t.Errorf("handler returned unexpected redirect URL: got %v want %v",
			location.String(), expectedRedirectURL)
	}
}

func TestRedirectsPlugin_NoMatch(t *testing.T) {
	nextHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	config := &Config{
		RedirectsAppURL: "http://localhost:8081",
	}

	handler, err := New(context.Background(), nextHandler, config, "test-middleware")
	if err != nil {
		t.Fatalf("Failed to create middleware: %v", err)
	}

	req, err := http.NewRequest("GET", "http://demo.localhost/nonexistent", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}
}

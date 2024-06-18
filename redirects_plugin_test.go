package redirects_traefik_middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServeHTTP_Match_Redirect(t *testing.T) {
	testCases := []struct {
		name             string
		requestURL       string
		expectedRedirect string
	}{
		{
			name:             "Exact domain match redirect",
			requestURL:       "https://demo.localhost/product/iphone",
			expectedRedirect: "https://new-demo.localhost/product/iphone",
		},
		{
			name:             "Exact relative path match redirect",
			requestURL:       "http://example.com/dedicated/host",
			expectedRedirect: "http://example.com/host",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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

			req, err := http.NewRequest("GET", tc.requestURL, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusFound {
				t.Errorf("handler returned wrong status code: got %v want %v",
					rr.Code, http.StatusFound)
			}

			location, err := rr.Result().Location()
			if err != nil {
				t.Fatalf("failed to get Location header: %v", err)
			}

			if location.String() != tc.expectedRedirect {
				t.Errorf("handler returned unexpected redirect URL: got %v want %v",
					location.String(), tc.expectedRedirect)
			}
		})
	}
}

func TestServeHTTP_NoMatch_Redirect(t *testing.T) {
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

	req, err := http.NewRequest("GET", "http://example.com/nonexistent", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

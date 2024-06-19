package redirects_traefik_middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const AppUrl = "http://localhost:8081"
const AppName = "test-middleware"

func TestServeHTTP_Match_Redirect(t *testing.T) {
	nextHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	config := &Config{
		RedirectsAppURL: AppUrl,
	}

	handler, err := New(context.Background(), nextHandler, config, AppName)
	if err != nil {
		t.Fatalf("Failed to create middleware: %v", err)
	}

	testCases := []struct {
		name             string
		requestURL       string
		expectedRedirect string
	}{
		{
			name:             "Exact domain match redirect",
			requestURL:       "https://fromsite2422.com",
			expectedRedirect: "https://new-domain/post/laptop/clothing/",
		},
		{
			name:             "Exact relative path match redirect",
			requestURL:       "http://example.com/product/furniture/electronics/",
			expectedRedirect: "http://example.com/category/iphone/books/",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

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

func BenchmarkServeHTTP_Match_Redirect(b *testing.B) {
	nextHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	config := &Config{
		RedirectsAppURL: AppUrl,
	}

	handler, err := New(context.Background(), nextHandler, config, AppName)
	if err != nil {
		b.Fatalf("Failed to create middleware: %v", err)
	}

	req, err := http.NewRequest("GET", "http://example.com/product/furniture/electronics/", nil)
	if err != nil {
		b.Fatalf("Failed to create request: %v", err)
	}

	b.ResetTimer() // Reset the timer to ignore the setup time

	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusFound {
			b.Errorf("handler returned wrong status code: got %v want %v",
				rr.Code, http.StatusFound)
		}

		// Check the Location header for the redirect URL
		expectedRedirectURL := "http://example.com/category/iphone/books/"
		location, _ := rr.Result().Location()
		if location.String() != expectedRedirectURL {
			b.Errorf("handler returned unexpected redirect URL: got %v want %v",
				location.String(), expectedRedirectURL)
		}
	}
}

func TestServeHTTP_NoMatch_Redirect(t *testing.T) {
	nextHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	config := &Config{
		RedirectsAppURL: AppUrl,
	}

	handler, err := New(context.Background(), nextHandler, config, AppName)
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

func BenchmarkServeHTTP_NoMatch_Redirect(b *testing.B) {
	nextHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	config := &Config{
		RedirectsAppURL: AppUrl,
	}

	handler, err := New(context.Background(), nextHandler, config, AppName)
	if err != nil {
		b.Fatalf("Failed to create middleware: %v", err)
	}

	req, err := http.NewRequest("GET", "http://example.com/nonexistent", nil)
	if err != nil {
		b.Fatalf("Failed to create request: %v", err)
	}

	b.ResetTimer() // Reset the timer to ignore the setup time

	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			b.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
	}
}

package v1

import (
	"context"
	"fmt"
	"github.com/TRIMM/redirects-traefik-middleware/internal/app"
	"github.com/TRIMM/redirects-traefik-middleware/internal/plugin"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type TestRedirectStruct struct {
	name             string
	requestURL       string
	expectedRedirect string
}

func getTestCases() *[]TestRedirectStruct {
	return &[]TestRedirectStruct{
		{
			"Exact domain match redirect",
			"https://old-domain.com",
			"https://new-domain/post/laptop/clothing/",
		},
		{
			"Exact relative path match redirect",
			"http://example.com/product/furniture/electronics/",
			"http://example.com/category/iphone/books/",
		},
	}
}

func startMockRedirectsServer(idx *app.IndexedRedirects) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		requestBody, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}

		request := string(requestBody)
		redirectURL, ok := idx.Match(request)
		if !ok {
			redirectURL = "@empty"
		}
		_, err = fmt.Fprintf(w, "%s", redirectURL)
		if err != nil {
			log.Println("Failed to write response:", err)
		}
	})

	return httptest.NewServer(mux)
}

func getMockRedirectsPlugin(serverURL string) http.Handler {
	nextHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	config := &plugin.Config{
		RedirectsAppURL: serverURL,
	}

	handler, err := New(context.Background(), nextHandler, config, "traefik-app-test")
	if err != nil {
		panic(err)
	}

	return handler
}

func TestServeHTTP_Match_Redirect(t *testing.T) {
	idx := app.NewIndexedRedirects()
	idx.IndexRule("", "old-domain.com", "https://new-domain/post/laptop/clothing/")
	idx.IndexRule("/product/furniture/electronics/", "", "/category/iphone/books/")

	mockServer := startMockRedirectsServer(idx)
	defer mockServer.Close()

	rp := getMockRedirectsPlugin(mockServer.URL)
	testCases := getTestCases()

	for _, tc := range *testCases {
		t.Run(tc.name, func(t *testing.T) {

			req, err := http.NewRequest("GET", tc.requestURL, strings.NewReader(tc.requestURL))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			rp.ServeHTTP(rr, req)

			if rr.Code != http.StatusFound {
				t.Errorf("handler returned wrong status code: got %v want %v",
					rr.Code, http.StatusFound)
			}

			location := rr.Header().Get("Location")
			if location != tc.expectedRedirect {
				t.Errorf("handler returned unexpected redirect URL: got %v want %v", location, tc.expectedRedirect)
			}
		})
	}
}

func TestServeHTTP_NoMatch_Redirect(t *testing.T) {
	idx := app.NewIndexedRedirects()
	mockServer := startMockRedirectsServer(idx)
	defer mockServer.Close()

	rp := getMockRedirectsPlugin(mockServer.URL)
	req, err := http.NewRequest("GET", "http://example.com/nonexistent", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	rp.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func BenchmarkMiddleware_Match_Redirect(b *testing.B) {
	idx := app.NewIndexedRedirects()
	idx.IndexRule("/product/furniture/electronics/", "", "/category/iphone/books/")

	mockServer := startMockRedirectsServer(idx)
	defer mockServer.Close()

	rp := getMockRedirectsPlugin(mockServer.URL)

	req, err := http.NewRequest("GET", "http://example.com/product/furniture/electronics/", nil)
	if err != nil {
		b.Fatalf("Failed to create request: %v", err)
	}

	b.ResetTimer() // Reset the timer to ignore the setup time

	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		rp.ServeHTTP(rr, req)

		if rr.Code != http.StatusFound {
			b.Errorf("handler returned wrong status code: got %v want %v",
				rr.Code, http.StatusFound)
		}

		expectedRedirectURL := "http://example.com/category/iphone/books/"
		location := rr.Header().Get("Location")
		if location != expectedRedirectURL {
			b.Errorf("handler returned unexpected redirect URL: got %v want %v",
				location, expectedRedirectURL)
		}
	}
}

func BenchmarkMiddleware_NoMatch_Redirect(b *testing.B) {
	idx := app.NewIndexedRedirects()
	mockServer := startMockRedirectsServer(idx)
	defer mockServer.Close()

	rp := getMockRedirectsPlugin(mockServer.URL)

	req, err := http.NewRequest("GET", "http://example.com/nonexistent", nil)
	if err != nil {
		b.Fatalf("Failed to create request: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		rp.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			b.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
	}
}

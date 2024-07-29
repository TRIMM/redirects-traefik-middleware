package v2

import (
	"fmt"
	v1 "github.com/TRIMM/redirects-traefik-middleware/api/v1"
	plugin "github.com/TRIMM/redirects-traefik-middleware/internal/plugin"
	"github.com/TRIMM/redirects-traefik-middleware/internal/plugin/v2/matcher"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRedirects(t *testing.T) {
	log.SetOutput(io.Discard)

	testCases := []struct {
		name              string
		targetUrl         string
		fromURL           string
		fromDomain        string
		toURL             string
		expectedToURL     string
		expectedStatsCode int
	}{
		{"matching", "https://www.trimm.nl/dedicated/host", "/dedicated/host", "", "/host", "https://www.trimm.nl/host", http.StatusFound},
		{"matching-wildcard", "https://www.trimm.nl/dedicated/host", "/dedicated/*", "", "/host", "https://www.trimm.nl/host", http.StatusFound},
		{"matching-replace", "https://www.trimm.nl/dedicated/bla", "/dedicated/{id:.+}", "", "/host/{id}", "https://www.trimm.nl/host/bla", http.StatusFound},
		{"matching-regex", "https://www.trimm.nl/dedicated/1337", "/dedicated/{id:^[0-9]+}", "", "/host/not-found", "https://www.trimm.nl/host/not-found", http.StatusFound},
	}

	for _, tc := range testCases {
		redirects := []v1.Redirect{
			{
				Id:         "18e3bde3-f087-4c53-a212-7f78f24978f9",
				FromURL:    tc.fromURL,
				FromDomain: tc.fromDomain,
				ToURL:      tc.toURL,
				UpdatedAt:  time.Date(2024, time.July, 26, 13, 10, 0, 0, time.UTC),
			},
		}

		m := matcher.NewRadixRedirectMatcher(&redirects)

		p := getPlugin(m)
		req := httptest.NewRequest("GET", tc.targetUrl, http.NoBody)
		rr := httptest.NewRecorder()

		// WHEN
		p.ServeHTTP(rr, req)

		// THEN
		if rr.Code != tc.expectedStatsCode {
			t.Errorf("[%s] got status code %d, want %d", tc.name, rr.Code, tc.expectedStatsCode)
		}

		if rr.Code == http.StatusFound {
			location := rr.Header().Get("Location")
			if location != tc.expectedToURL {
				t.Errorf("[%s] got location %s, want %s", tc.name, location, tc.expectedToURL)
			}
		}

	}
}

func BenchmarkMatchingRedirectRadixRedirectMatcher(b *testing.B) {
	log.SetOutput(io.Discard)
	numberOfRedirects := []int{10, 100, 1000, 10000, 100000}

	for _, num := range numberOfRedirects {
		// Create redirects
		redirects := getDummyRedirects(num)
		m := matcher.NewRadixRedirectMatcher(&redirects)

		b.Run(fmt.Sprintf("%d", num), func(b *testing.B) {
			p := getPlugin(m)
			req := httptest.NewRequest("GET", "https://www.trimm.nl/dedicated/host", http.NoBody)
			rr := httptest.NewRecorder()

			for i := 0; i < b.N; i++ {
				p.ServeHTTP(rr, req)
			}
		})
	}
}

func getDummyRedirects(num int) []v1.Redirect {
	// Create redirects
	var redirects []v1.Redirect
	for i := range num {
		redirects = append(redirects, v1.Redirect{
			Id:         uuid.NewString(),
			FromURL:    fmt.Sprintf("/%d-no/%d-match/%d", i, i, i),
			FromDomain: "",
			ToURL:      "/host",
			UpdatedAt:  time.Date(2024, time.July, 26, 13, 10, 0, 0, time.UTC),
		})
	}

	// Worst-case scenario where the matching redirect is latest in the slice
	redirects = append(redirects, v1.Redirect{
		Id:         "18e3bde3-f087-4c53-a212-7f78f24978f9",
		FromURL:    "/dedicated/host",
		FromDomain: "",
		ToURL:      "/host",
		UpdatedAt:  time.Date(2024, time.July, 26, 13, 10, 0, 0, time.UTC),
	})
	return redirects
}

func getPlugin(matcher matcher.Matcher) RedirectsPlugin {
	config := plugin.Config{
		V2: plugin.V2Config{
			Enabled:      true,
			ClientName:   "test-client",
			ClientSecret: "test-client-secret",
			ServerURL:    "http://localhost",
			JwtSecret:    "test-jwt-secret",
		},
	}

	nextHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	return RedirectsPlugin{
		next:    nextHandler,
		name:    "v2-test-plugin",
		config:  &config.V2,
		matcher: matcher,
	}
}

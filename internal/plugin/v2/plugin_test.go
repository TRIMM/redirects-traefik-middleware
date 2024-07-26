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

func TestMatchingRedirect(t *testing.T) {
	// GIVEN
	redirects := []v1.Redirect{
		{
			Id:         "18e3bde3-f087-4c53-a212-7f78f24978f9",
			FromURL:    "/dedicated/host",
			FromDomain: "",
			ToURL:      "/host",
			UpdatedAt:  time.Date(2024, time.July, 26, 13, 10, 0, 0, time.UTC),
		},
	}

	matchers := map[string]matcher.Matcher{
		"RegexRedirectMatcher": matcher.NewRegexRedirectMatcher(&redirects),
		"RadixRedirectMatcher": matcher.NewRadixRedirectMatcher(&redirects),
	}

	for k, m := range matchers {
		p := getPlugin(m)
		req := httptest.NewRequest("GET", "https://www.trimm.nl/dedicated/host", http.NoBody)
		rr := httptest.NewRecorder()

		// WHEN
		p.ServeHTTP(rr, req)

		// THEN
		if rr.Code != http.StatusFound {
			t.Errorf("[%s] got status code %d, want %d", k, rr.Code, http.StatusFound)
		}
	}
}

func BenchmarkMatchingRedirect(b *testing.B) {
	log.SetOutput(io.Discard)
	numberOfRedirects := []int{10, 100, 1000, 10000, 100000}

	for _, num := range numberOfRedirects {
		// Create redirects
		var redirects []v1.Redirect
		for i := range num {
			redirects = append(redirects, v1.Redirect{
				Id:         uuid.NewString(),
				FromURL:    fmt.Sprintf("/no/match/%d", i),
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

		matchers := map[string]matcher.Matcher{
			"RegexRedirectMatcher": matcher.NewRegexRedirectMatcher(&redirects),
			"RadixRedirectMatcher": matcher.NewRadixRedirectMatcher(&redirects),
		}

		for k, m := range matchers {
			b.Run(fmt.Sprintf("redirects_%d_%s", num, k), func(b *testing.B) {
				p := getPlugin(m)
				req := httptest.NewRequest("GET", "https://www.trimm.nl/dedicated/host", http.NoBody)
				rr := httptest.NewRecorder()

				for i := 0; i < b.N; i++ {
					p.ServeHTTP(rr, req)
				}
			})
		}
	}
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

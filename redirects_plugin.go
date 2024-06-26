package redirects_traefik_middleware

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/dgraph-io/ristretto"
)

const noMatchMarker = "@no_match"

type Config struct {
	RedirectsAppURL string `json:"redirectsAppURL,omitempty"`
}

func CreateConfig() *Config {
	return &Config{}
}

type RedirectsPlugin struct {
	next            http.Handler
	name            string
	redirectCache   *ristretto.Cache
	redirectsAppURL string
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M)
		MaxCost:     1 << 30, // maximum cost of cache (1GB)
		BufferItems: 64,      // number of keys per Get buffer
	})
	if err != nil {
		return nil, err
	}

	log.Println("Redirects Traefik Middleware v0.1.9")

	if len(config.RedirectsAppURL) == 0 {
		return nil, fmt.Errorf("RedirectsPlugin 'redirectsURL' cannot be empty")
	}

	log.Println("Redirects App Url [" + strings.ToLower(config.RedirectsAppURL) + "]")

	return &RedirectsPlugin{
		next:            next,
		name:            name,
		redirectCache:   cache,
		redirectsAppURL: config.RedirectsAppURL,
	}, nil
}

/*
ServeHTTP intercepts a request and matches it against the existing rules
If a match is found, it redirects accordingly
*/
func (rp *RedirectsPlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	fullURL := getFullURL(req)
	relativeURL := req.URL.Path

	responseURL, found := rp.getCachedRedirect(fullURL)
	if !found {
		responseURL, found = rp.getCachedRedirect(relativeURL)
		// Cache the redirect for full URL if found for relative URL
		if found && responseURL != noMatchMarker {
			rp.redirectCache.Set(fullURL, responseURL, 1)
		}
	}

	// Handle the found redirect or pass to the next handler
	if found && responseURL != noMatchMarker {
		log.Printf("Redirect exists: %s --> %s\n", fullURL, responseURL)
		if !strings.HasPrefix(responseURL, "http") {
			responseURL = getRelativeRedirect(req, responseURL)
		}
		http.Redirect(rw, req, responseURL, http.StatusFound)
		return
	}

	log.Printf("Redirect does not exist: %s\n", fullURL)
	rp.next.ServeHTTP(rw, req)
}

func (rp *RedirectsPlugin) getCachedRedirect(url string) (string, bool) {
	value, found := rp.redirectCache.Get(url)
	if found {
		return value.(string), true
	}

	// Fetch from the redirect service if not found in cache
	responseURL, isMatch, err := sendRedirectMatchRequest(rp.redirectsAppURL, url)
	if err != nil || !isMatch {
		rp.redirectCache.Set(url, noMatchMarker, 1)
		return "", false
	}

	rp.redirectCache.Set(url, responseURL, 1)

	return responseURL, true
}

func sendRedirectMatchRequest(redirectsAppURL, url string) (string, bool, error) {
	response, err := http.Post(redirectsAppURL, "text/plain", strings.NewReader(url))
	if err != nil {
		return "", false, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", false, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", false, err
	}
	redirectURL := string(body)
	if redirectURL == "@empty" {
		return "", false, nil
	}

	return redirectURL, true, nil
}

func getFullURL(req *http.Request) string {
	var proto = "https://"
	if req.TLS == nil {
		proto = "http://"
	}

	var host = req.URL.Host
	if len(host) == 0 {
		host = req.Host
	}

	return strings.ToLower(proto + host + req.URL.Path)
}

func getRelativeRedirect(req *http.Request, relativeURL string) string {
	var proto = "https://"
	if req.TLS == nil {
		proto = "http://"
	}

	var host = req.URL.Host
	if len(host) == 0 {
		host = req.Host
	}

	return strings.ToLower(proto + host + relativeURL)
}

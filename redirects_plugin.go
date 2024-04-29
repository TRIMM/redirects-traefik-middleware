package redirects_traefik_middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type Config struct {
	RedirectsURL string `json:"redirectsURL,omitempty"`
}

func CreateConfig() *Config {
	return &Config{}
}

type RedirectsPlugin struct {
	next         http.Handler
	name         string
	redirectsURL string
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	fmt.Println("Redirects Traefik Middleware v0.1.4")

	if len(config.RedirectsURL) == 0 {
		return nil, fmt.Errorf("RedirectsPlugin 'redirectsURL' cannot be empty")
	}

	fmt.Println("RedirectsPlugin redirectsURL [" + strings.ToLower(config.RedirectsURL) + "]")

	return &RedirectsPlugin{
		next:         next,
		name:         name,
		redirectsURL: config.RedirectsURL,
	}, nil
}

/*
ServeHTTP intercepts a request and matches it against the existing rules presented in the Trie Data structure
If a match is found, it redirects accordingly
*/
func (rp *RedirectsPlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	//var request = getFullURL(req)
	//if err := rp.logger.LogRequest(request); err != nil {
	//	log.Println("Failed to log request to file: ", err)
	//}
	//
	//redirectURL, ok := rp.redirectsManager.Trie.Match(request)
	//if !ok {
	//	log.Println("No matching redirect rule found!")
	//}

	//http.Redirect(rw, req, redirectURL, http.StatusFound)
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

	var answer = proto + host + req.URL.Path
	return strings.ToLower(answer)
}

package redirects_traefik_middleware

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type Config struct {
	RedirectsAppURL string `json:"redirectsAppURL,omitempty"`
}

func CreateConfig() *Config {
	return &Config{}
}

type RedirectsPlugin struct {
	next            http.Handler
	name            string
	redirectsAppURL string
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	log.Println("Redirects Traefik Middleware v0.1.9")

	if len(config.RedirectsAppURL) == 0 {
		return nil, fmt.Errorf("RedirectsPlugin 'redirectsURL' cannot be empty")
	}

	log.Println("Redirects App Url [" + strings.ToLower(config.RedirectsAppURL) + "]")

	return &RedirectsPlugin{
		next:            next,
		name:            name,
		redirectsAppURL: config.RedirectsAppURL,
	}, nil
}

/*
ServeHTTP intercepts a request and matches it against the existing rules
If a match is found, it redirects accordingly
*/
func (rp *RedirectsPlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var fullURL = getFullURL(req)
	var relativeURL = req.URL.Path

	var responseURL string
	var isMatch bool
	var err error

	responseURL, isMatch, err = getRedirectMatch(rp.redirectsAppURL, fullURL)
	if err != nil {
		log.Println("Error sending HTTP request:", err)
		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !isMatch {
		responseURL, isMatch, err = getRedirectMatch(rp.redirectsAppURL, relativeURL)
	}

	if isMatch {
		log.Println("Redirect exists: " + fullURL + "-->" + responseURL)
		if !strings.HasPrefix(responseURL, "http") {
			responseURL = getRelativeRedirect(req, responseURL)
		}
		http.Redirect(rw, req, responseURL, http.StatusFound)
	} else {
		log.Println("Redirect does not exist: " + fullURL + "-->" + responseURL)
		rp.next.ServeHTTP(rw, req)
	}
}

func getRedirectMatch(appURL, request string) (string, bool, error) {
	var client = &http.Client{}
	req, err := http.NewRequest("POST", appURL, strings.NewReader(request))
	if err != nil {
		return "", false, err
	}

	req.Header.Set("Content-Type", "text/plain")

	res, err := client.Do(req)
	if err != nil {
		return "", false, err
	}

	defer func() {
		err := res.Body.Close()
		if err != nil {
			log.Println("Error closing response body: ", err)
		}
	}()

	response, err := io.ReadAll(res.Body)
	if err != nil {
		return "", false, err
	}

	responseStr := string(response)
	return responseStr, responseStr != "@empty", nil
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

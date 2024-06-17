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
	log.Println("Redirects Traefik Middleware v0.1.8")

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
	var relativeURL = req.URL.Path

	var responseURL, err = getRedirectMatch(rp.redirectsAppURL, relativeURL)
	if err != nil {
		log.Println("Error sending HTTP request:", err)
		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if responseURL != "@empty" {
		log.Println("Redirect exists: " + relativeURL + "-->" + responseURL)
		absolutePath := getFullURL(req, responseURL)
		http.Redirect(rw, req, absolutePath, http.StatusFound)
	} else {
		log.Println("Redirect does not exist: " + relativeURL + "-->" + responseURL)
		http.NotFound(rw, req)
	}
}

func getRedirectMatch(appURL, request string) (string, error) {
	var client = &http.Client{}
	req, err := http.NewRequest("GET", appURL, strings.NewReader(request))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "text/plain")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer func() {
		err := res.Body.Close()
		if err != nil {
			log.Println("Error closing response body: ", err)
		}
	}()

	response, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(response), nil
}

func getFullURL(req *http.Request, responseURL string) string {
	var proto = "https://"
	if req.TLS == nil {
		proto = "http://"
	}

	var host = req.URL.Host
	if len(host) == 0 {
		host = req.Host
	}

	var answer = proto + host + responseURL
	return strings.ToLower(answer)
}

package redirects_traefik_middleware

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
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
	fmt.Println("Redirects Traefik Middleware v0.1.4")

	if len(config.RedirectsAppURL) == 0 {
		return nil, fmt.Errorf("RedirectsPlugin 'redirectsURL' cannot be empty")
	}

	fmt.Println("RedirectsPlugin redirectsURL [" + strings.ToLower(config.RedirectsAppURL) + "]")

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
	var request = getFullURL(req)
	fmt.Println("Plugin side: " + request)
	var answerURL = req.Host
	dial, dialErr := net.DialTimeout("tcp", rp.redirectsAppURL, 2*time.Second)
	if dialErr != nil {
		fmt.Println("The redirects middleware is not reachable on " + rp.redirectsAppURL)
	} else {
		defer func() {
			err := dial.Close()
			if err != nil {
				log.Println("Error closing the connection:", err)
			}
		}()

		fmt.Fprintf(dial, request+"\n")
		answer, _ := bufio.NewReader(dial).ReadString('\n')
		answerTrim := strings.TrimSuffix(answer, "\n")
		if answerTrim != "@empty" {
			answerURL = answerTrim
			fmt.Println("Redirect exists: " + request + "-->" + answerURL)
			http.Redirect(rw, req, answerURL, 302)
		} else {
			fmt.Println("Redirect does not exist!")
		}
	}
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

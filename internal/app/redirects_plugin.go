package app

import (
	"log"
	"net/http"
	"strings"
)

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

/*
ServeHTTP intercepts a request and matches it against the existing rules presented in the Trie Data structure
If a match is found, it redirects accordingly
*/
func (rm *RedirectManager) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var request = getFullURL(req)
	if err := rm.logger.LogRequest(request); err != nil {
		log.Println("Failed to log request to file: ", err)
	}

	redirectURL, ok := rm.trie.Match(request)
	if !ok {
		log.Println("No matching redirect rule found!")
	}

	http.Redirect(rw, req, redirectURL, 302)
}

package main

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
func (t *Trie) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var request = getFullURL(req)
	redirectURL, ok := t.Match(request)
	if !ok {
		log.Println("No matching redirect rule found!")
	}

	http.Redirect(rw, req, redirectURL, 302)
}

package main

import (
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

func (rm *RedirectManager) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var request = getFullURL(req)
	// TODO:: Revise this assignment
	var answerUrl = req.URL.Host

	for _, redirect := range rm.redirects {
		if redirect.FromUrl == request {
			answerUrl = redirect.ToUrl
			break
		}
	}

	http.Redirect(rw, req, answerUrl, 302)
}

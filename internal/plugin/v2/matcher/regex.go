package matcher

import (
	api "github.com/TRIMM/redirects-traefik-middleware/api/v1"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type RegexRedirectMatcher struct {
	redirects *[]api.Redirect
}

func NewRegexRedirectMatcher(redirects *[]api.Redirect) *RegexRedirectMatcher {
	return &RegexRedirectMatcher{
		redirects: redirects,
	}
}

func (m RegexRedirectMatcher) Match(req *http.Request, url string) string {
	// Naive implementation
	for _, redirect := range *m.redirects {
		regex, err := regexp.Compile(redirect.FromURL)
		if err != nil {
			log.Printf("Error compiling regex [%s]: %v", redirect.FromURL, err)
			break
		}

		if regex.MatchString(url) {
			log.Printf("Found redirect [%s] => [%s]", redirect.FromURL, redirect.ToURL)
			return getRedirectUrl(req, redirect.ToURL)
		}
	}

	return ""
}

func getRedirectUrl(req *http.Request, relativeURL string) string {
	proto := "https://"
	if req.TLS == nil {
		proto = "http://"
	}

	host := req.URL.Host
	if len(host) == 0 {
		host = req.Host
	}

	return strings.ToLower(proto + host + relativeURL)
}

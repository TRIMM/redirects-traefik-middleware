package matcher

import (
	"net/http"
	"strings"
)

type Matcher interface {
	Match(req *http.Request, url string) string
}

func getRedirectUrl(req *http.Request, relativeURL string) string {
	host := req.URL.Host
	if len(host) == 0 {
		host = req.Host
	}

	var sb strings.Builder
	if req.TLS != nil {
		sb.WriteString("https://")
	} else {
		sb.WriteString("http://")
	}

	sb.WriteString(host)
	sb.WriteString(relativeURL)

	return strings.ToLower(sb.String())
}

package matcher

import (
	api "github.com/TRIMM/redirects-traefik-middleware/api/v1"
	"log"
	"net/http"
	"strings"
)

type RadixRedirectMatcher struct {
	redirects *[]api.Redirect
	tree      *node
}

var noopHandlerFunc = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

func NewRadixRedirectMatcher(redirects *[]api.Redirect) *RadixRedirectMatcher {
	tr := &node{}
	for _, redirect := range *redirects {
		tr.InsertRoute(mGET, redirect.FromURL, noopHandlerFunc, redirect.ToURL)
	}

	return &RadixRedirectMatcher{
		redirects: redirects,
		tree:      tr,
	}
}

func (m RadixRedirectMatcher) Match(req *http.Request, url string) string {
	rctx := NewRouteContext()
	node, _, _, toUrl := m.tree.FindRoute(rctx, mGET, url)

	if toUrl != "" && node.typ != ntRegexp {
		return getRedirectUrl(req, toUrl)
	} else if toUrl != "" {
		if node.typ == ntRegexp && len(rctx.URLParams.Keys) > 0 {
			// Replace all parameters
			for i, urlParam := range rctx.URLParams.Keys {
				ps, pe := getParameterBeginAndEnd(toUrl, urlParam)
				if ps < 0 || pe < 0 {
					continue
				}

				parameter := toUrl[ps : pe+1]
				value := rctx.URLParams.Values[i]
				toUrl = strings.ReplaceAll(toUrl, parameter, value)
			}
		}

		return getRedirectUrl(req, toUrl)
	}

	return ""
}

func getParameterBeginAndEnd(pattern, urlParam string) (int, int) {
	ps := strings.Index(pattern, "{"+urlParam)
	if ps < 0 {
		return -1, -1
	}

	// Read to closing } taking into account opens and closes in curl count (cc)
	cc := 0
	pe := ps
	for i, c := range pattern[ps:] {

		if c == '{' {
			cc++
		} else if c == '}' {
			cc--
			if cc == 0 {
				pe = ps + i
				break
			}
		}
	}
	if pe == ps {
		log.Printf("route param closing delimiter '}' is missing for pattern [%s]", pattern)
		return -1, -1
	}

	return ps, pe
}

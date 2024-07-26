package matcher

import (
	api "github.com/TRIMM/redirects-traefik-middleware/api/v1"
	"net/http"
	"strconv"
)

type RadixRedirectMatcher struct {
	redirects *[]api.Redirect
	tree      *node
}

type dummyResponseWriter struct {
	header     http.Header
	body       []byte
	statusCode int
}

func (d *dummyResponseWriter) Header() http.Header {
	return d.header
}

func (d *dummyResponseWriter) Write(body []byte) (int, error) {
	d.body = body
	return len(body), nil
}

func (d *dummyResponseWriter) WriteHeader(statusCode int) {
	d.statusCode = statusCode
	if d.header == nil {
		d.header = http.Header{}
	}
	d.Header().Set("Status", strconv.Itoa(statusCode))
}

func NewRadixRedirectMatcher(redirects *[]api.Redirect) *RadixRedirectMatcher {
	tr := &node{}
	for _, redirect := range *redirects {
		tr.InsertRoute(mGET, redirect.FromURL, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, getRedirectUrl(r, redirect.ToURL), http.StatusFound)
		}))
	}

	return &RadixRedirectMatcher{
		redirects: redirects,
		tree:      tr,
	}
}

func (m RadixRedirectMatcher) Match(req *http.Request, url string) string {
	rctx := NewRouteContext()
	node, _, handler := m.tree.FindRoute(rctx, mGET, url)

	if node != nil && handler != nil {
		responseWriter := &dummyResponseWriter{
			header: make(http.Header),
		}
		handler.ServeHTTP(responseWriter, req)

		if responseWriter.statusCode == http.StatusFound {
			return responseWriter.Header().Get("Location")
		}
	}

	return ""
}

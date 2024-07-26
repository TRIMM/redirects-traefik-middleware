package matcher

import "net/http"

type Matcher interface {
	Match(req *http.Request, url string) string
}

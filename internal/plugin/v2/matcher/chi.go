// Implementation below is a based on the original work by
// Chi in https://github.com/go-chi/chi/blob/master/chi.go
// (MIT licensed).

package matcher

import "net/http"

// Routes interface adds two methods for router traversal, which is also
// used by the `docgen` subpackage to generation documentation for Routers.
type Routes interface {
	// Routes returns the routing tree in an easily traversable structure.
	Routes() []Route

	// Middlewares returns the list of middlewares in use by the router.
	Middlewares() Middlewares

	// Match searches the routing tree for a handler that matches
	// the method/path - similar to routing a http request, but without
	// executing the handler thereafter.
	Match(rctx *Context, method, path string) bool
}

// Middlewares type is a slice of standard middleware handlers with methods
// to compose middleware chains and http.Handler's.
type Middlewares []func(http.Handler) http.Handler

// ChainHandler is a http.Handler with support for handler composition and
// execution.
type ChainHandler struct {
	Endpoint    http.Handler
	chain       http.Handler
	Middlewares Middlewares
}

func (c *ChainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.chain.ServeHTTP(w, r)
}

package v2

import (
	"context"
	"fmt"
	"github.com/TRIMM/redirects-traefik-middleware/internal/plugin"
	"log"
	"net/http"
	"strings"
)

type RedirectsPlugin struct {
	next            http.Handler
	name            string
	redirectsAppURL string
}

func New(ctx context.Context, next http.Handler, config *plugin.Config, name string) (http.Handler, error) {
	if len(config.RedirectsAppURL) == 0 {
		return nil, fmt.Errorf("RedirectsPlugin 'redirectsURL' cannot be empty")
	}

	log.Println("Redirects App Url [" + strings.ToLower(config.RedirectsAppURL) + "]")

	return &RedirectsPlugin{
		next:            next,
		name:            name,
		redirectsAppURL: config.RedirectsAppURL,
	}, nil
}

func (a *RedirectsPlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	a.next.ServeHTTP(rw, req)
}

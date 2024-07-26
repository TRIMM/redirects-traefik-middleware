package redirects_traefik_middleware

import (
	"context"
	"github.com/TRIMM/redirects-traefik-middleware/internal/plugin"
	v1 "github.com/TRIMM/redirects-traefik-middleware/internal/plugin/v1"
	v2 "github.com/TRIMM/redirects-traefik-middleware/internal/plugin/v2"
	"net/http"
)

func CreateConfig() *plugin.Config {
	return &plugin.Config{
		RedirectsAppURL: "",
		V2:              true,
	}
}

func New(ctx context.Context, next http.Handler, config *plugin.Config, name string) (http.Handler, error) {
	if config.V2 {
		return v2.New(ctx, next, config, name)
	} else {
		return v1.New(ctx, next, config, name)
	}
}

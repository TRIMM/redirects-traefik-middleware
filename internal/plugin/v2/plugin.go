package v2

import (
	"context"
	api "github.com/TRIMM/redirects-traefik-middleware/api/v1"
	"github.com/TRIMM/redirects-traefik-middleware/internal/plugin"
	"github.com/TRIMM/redirects-traefik-middleware/internal/plugin/v2/matcher"
	"log"
	"net/http"
	"strings"
)

type RedirectsPlugin struct {
	next    http.Handler
	name    string
	config  *plugin.V2Config
	matcher matcher.Matcher
}

func New(ctx context.Context, next http.Handler, config *plugin.Config, name string) (http.Handler, error) {
	log.Println("Redirects Traefik Middleware v2.0.0")

	v2Config := &config.V2
	if v2Config.ServerURL == "" {
		log.Fatal("Server url is required")
	}
	if v2Config.ClientName == "" {
		log.Fatal("Client name is required")
	}
	if v2Config.ClientSecret == "" {
		log.Fatal("Client secret is required")
	}
	if v2Config.JwtSecret == "" {
		log.Fatal("Jwt secret is required")
	}

	// Fetch the redirects
	log.Printf("Fetching redirects from [%s]", v2Config.ServerURL)
	authData := api.NewAuthData(v2Config.ClientName, v2Config.ClientSecret, v2Config.ServerURL, v2Config.JwtSecret)
	graphqlClient := api.NewGraphQLClient(authData)

	redirects, err := graphqlClient.ExecuteRedirectsQuery()
	if err != nil {
		return nil, err
	}

	regexMatcher := matcher.NewRegexRedirectMatcher(&redirects)

	return &RedirectsPlugin{
		next:    next,
		name:    name,
		config:  v2Config,
		matcher: regexMatcher,
	}, nil
}

func (a *RedirectsPlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	log.Printf("Handling request [%+v]", req.URL)

	absoluteUrl := getAbsoluteUrl(req)
	relativeURL := req.URL.Path

	log.Printf("Handling absolute url [%s]", absoluteUrl)
	log.Printf("Handling relative url [%s]", relativeURL)

	if redirectUrl := a.matcher.Match(req, relativeURL); redirectUrl != "" {
		http.Redirect(rw, req, redirectUrl, http.StatusFound)
		return
	}

	if redirectUrl := a.matcher.Match(req, absoluteUrl); redirectUrl != "" {
		http.Redirect(rw, req, redirectUrl, http.StatusFound)
		return
	}

	a.next.ServeHTTP(rw, req)
}

func getAbsoluteUrl(req *http.Request) string {
	var proto = "https://"
	if req.TLS == nil {
		proto = "http://"
	}

	var host = req.URL.Host
	if len(host) == 0 {
		host = req.Host
	}

	return strings.ToLower(proto + host + req.URL.Path)
}

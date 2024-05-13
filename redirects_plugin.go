package redirects_traefik_middleware

import (
	"context"
	"fmt"
	"github.com/TRIMM/redirects-traefik-middleware/pkg/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net/http"
	"strings"
	"time"
)

type Config struct {
	RedirectsAppURL string `json:"redirectsAppURL,omitempty"`
}

func CreateConfig() *Config {
	return &Config{}
}

type RedirectsPlugin struct {
	next            http.Handler
	name            string
	redirectsAppURL string
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	log.Println("Redirects Traefik Middleware v0.1.7")

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

/*
ServeHTTP intercepts a request and matches it against the existing rules
If a match is found, it redirects accordingly
*/
func (rp *RedirectsPlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var request = getFullURL(req)

	var responseURL, err = getRedirectResponse(rp.redirectsAppURL, request)
	if err != nil {
		log.Println("Failed to connect to the gRPC server: ", err)
		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if responseURL != "@empty" {
		log.Println("Redirect exists: " + request + "-->" + responseURL)
		http.Redirect(rw, req, responseURL, http.StatusFound)
	} else {
		log.Println("Redirect does not exist: " + request + "-->" + responseURL)
		http.NotFound(rw, req)
	}
}

type RedirectServiceClient interface {
	GetRedirectMatch(ctx context.Context, in *types.Request, opts ...grpc.CallOption) (*types.Response, error)
}

// Create a concrete implementation of RedirectServiceClient
type redirectServiceClient struct {
	cc *grpc.ClientConn
}

func (c *redirectServiceClient) GetRedirectMatch(ctx context.Context, in *types.Request, opts ...grpc.CallOption) (*types.Response, error) {
	out := new(types.Response)
	err := c.cc.Invoke(ctx, "/RedirectService/GetRedirectMatch", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func NewRedirectServiceClient(cc *grpc.ClientConn) RedirectServiceClient {
	return &redirectServiceClient{cc: cc}
}

func getRedirectResponse(appURL, request string) (string, error) {
	// Revise to grpc.WithTransportCredentials(credentials.NewTLS())
	conn, err := grpc.Dial(appURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", err
	}
	defer conn.Close()

	client := NewRedirectServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := client.GetRedirectMatch(ctx, &types.Request{URL: request})
	if err != nil {
		return "", err
	}

	return r.RedirectURL, nil
}

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

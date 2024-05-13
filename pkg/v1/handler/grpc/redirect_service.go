package grpc

import (
	"context"
	"github.com/TRIMM/redirects-traefik-middleware/internal/app"
	"github.com/TRIMM/redirects-traefik-middleware/pkg/types"
	"google.golang.org/grpc"
	"log"
)

type RedirectService interface {
	GetRedirectMatch(ctx context.Context, request *types.Request) (*types.Response, error)
}

type RedirectsServStruct struct {
	logger          *app.Logger
	redirectManager *app.RedirectManager
}

func RegisterRedirectServiceServer(server *grpc.Server, redirectService *RedirectsServStruct) {
	RegisterRedirectServiceServer(server, redirectService)
}

func NewServer(grpcServer *grpc.Server, logger *app.Logger, redirectManager *app.RedirectManager) {
	redirectsGrpc := &RedirectsServStruct{logger: logger, redirectManager: redirectManager}
	RegisterRedirectServiceServer(grpcServer, redirectsGrpc)
}

func (srv *RedirectsServStruct) GetRedirectMatch(ctx context.Context, in *types.Request) (*types.Response, error) {
	// Logging the incoming requests
	request := in.URL
	if err := srv.logger.LogRequest(request); err != nil {
		log.Println("Failed to log request to file: ", err)
	}

	// Matching against the defined redirects
	redirectURL, ok := srv.redirectManager.Trie.Match(request)
	if !ok {
		redirectURL = "@empty"
	}

	return &types.Response{RedirectURL: redirectURL}, nil
}

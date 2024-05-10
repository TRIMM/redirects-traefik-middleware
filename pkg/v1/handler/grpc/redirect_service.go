package grpc

import (
	"context"
	"github.com/TRIMM/redirects-traefik-middleware/internal/app"
	pb "github.com/TRIMM/redirects-traefik-middleware/proto"
	"google.golang.org/grpc"
	"log"
)

type RedirectsServStruct struct {
	pb.UnimplementedRedirectServiceServer
	logger          *app.Logger
	redirectManager *app.RedirectManager
}

func NewServer(grpcServer *grpc.Server, logger *app.Logger, redirectManager *app.RedirectManager) {
	redirectsGrpc := &RedirectsServStruct{logger: logger, redirectManager: redirectManager}
	pb.RegisterRedirectServiceServer(grpcServer, redirectsGrpc)
}

func (srv *RedirectsServStruct) GetRedirectMatch(ctx context.Context, in *pb.Request) (*pb.Response, error) {
	// Logging the incoming requests
	request := in.GetUrl()
	if err := srv.logger.LogRequest(request); err != nil {
		log.Println("Failed to log request to file: ", err)
	}

	// Matching against the defined redirects
	redirectURL, ok := srv.redirectManager.Trie.Match(request)
	if !ok {
		redirectURL = "@empty"
	}

	return &pb.Response{RedirectUrl: redirectURL}, nil
}

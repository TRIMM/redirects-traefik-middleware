package main

import (
	api "github.com/TRIMM/redirects-traefik-middleware/api/v1"
	"github.com/TRIMM/redirects-traefik-middleware/internal/app"
	handler "github.com/TRIMM/redirects-traefik-middleware/pkg/v1/handler/grpc"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	log.Println("Starting redirects-traefik-middleware")
	config := NewAppConfig()

	authData := api.NewAuthData(config.clientName, config.clientSecret, config.serverURL, config.jwtSecret)
	graphqlClient := api.NewGraphQLClient(authData)

	logger := app.NewLogger(config.logFilePath, graphqlClient)
	logger.SendLogsWeekly()

	redirectManager := app.NewRedirectManager(dbConnect(config.dbFilePath), graphqlClient)
	redirectManager.PopulateMapWithDataFromDB()
	redirectManager.PopulateTrieWithRedirects()

	//Create channels for fetching redirects periodically
	redirectsCh := make(chan []api.Redirect)
	errCh := make(chan error)

	go func() {
		defer close(redirectsCh)
		defer close(errCh)

		redirectManager.FetchRedirectsOverChannel(redirectsCh, errCh)
	}()
	go redirectManager.SyncRedirects(redirectsCh, errCh)

	startGrpcServer(logger, redirectManager)
}

func startGrpcServer(logger *app.Logger, redirectManager *app.RedirectManager) {
	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	// Register the server with the gRPC service
	handler.NewServer(s, logger, redirectManager)
	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

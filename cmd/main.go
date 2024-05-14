package main

import (
	api "github.com/TRIMM/redirects-traefik-middleware/api/v1"
	"github.com/TRIMM/redirects-traefik-middleware/internal/app"
	"github.com/TRIMM/redirects-traefik-middleware/pkg/v1/handlers"
	"log"
	"net/http"
)

func main() {
	log.Println("Starting redirects-traefik-middleware")

	// Create needed configuration for the authentication
	config := NewAppConfig()
	authData := api.NewAuthData(config.clientName, config.clientSecret, config.serverURL, config.jwtSecret)
	graphqlClient := api.NewGraphQLClient(authData)

	// Send logs of incoming requests to the Central API once a week
	logger := app.NewLogger(config.logFilePath, graphqlClient)
	logger.SendLogsWeekly()

	// Add the internal redirects data
	redirectManager := app.NewRedirectManager(dbConnect(config.dbFilePath), graphqlClient)
	redirectManager.PopulateMapsWithDataFromDB()

	// Create channels for fetching redirects periodically
	redirectsCh := make(chan []api.Redirect)
	errCh := make(chan error)

	// Start the go-routine for fetching & syncing the redirects periodically
	go func() {
		defer close(redirectsCh)
		defer close(errCh)

		redirectManager.FetchRedirectsOverChannel(redirectsCh, errCh)
	}()
	go redirectManager.SyncRedirects(redirectsCh, errCh)

	// Register the server for handling requests
	NewHTTPServer(logger, redirectManager)
}

func NewHTTPServer(logger *app.Logger, redirectManager *app.RedirectManager) {
	http.HandleFunc("/", handlers.GetRedirectMatch(logger, redirectManager))
	log.Fatal(http.ListenAndServe(":8081", nil))
}

package main

import (
	api "github.com/TRIMM/redirects-traefik-middleware/api/v1"
	"github.com/TRIMM/redirects-traefik-middleware/internal/app"
	"github.com/TRIMM/redirects-traefik-middleware/pkg/handlers"
	"log"
	"net/http"
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

	handleRequests(logger, redirectManager)
}

// Register the handlers
func handleRequests(logger *app.Logger, redirectManager *app.RedirectManager) {
	http.HandleFunc("/", handlers.GetRedirectMatch(logger, redirectManager))
	log.Fatal(http.ListenAndServe(":8081", nil))
}

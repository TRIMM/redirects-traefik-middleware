package main

import (
	"fmt"
	api "github.com/TRIMM/redirects-traefik-middleware/api/v1"
	"github.com/TRIMM/redirects-traefik-middleware/internal/app"
	"log"
	"net/http"
)

func main() {
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

	// Register the handler
	http.HandleFunc("/", handleRequest(logger, redirectManager))
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func handleRequest(logger *app.Logger, redirectManager *app.RedirectManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request := r.URL.String()
		fmt.Println("Handler side: " + request)

		//Logging the incoming requests
		if err := logger.LogRequest(request); err != nil {
			log.Println("Failed to log request to file: ", err)
		}

		//Matching against the defined redirects
		redirectURL, ok := redirectManager.Trie.Match(request)
		if !ok {
			redirectURL = "@empty"
		}
		// Write the redirectURL to the response
		fmt.Fprint(w, redirectURL)
	}
}

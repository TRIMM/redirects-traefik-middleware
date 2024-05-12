package main

import (
	"fmt"
	"github.com/TRIMM/redirects-traefik-middleware/internal/app"
	"log"
)

func main() {
	trie := app.NewTrie()

	// Insert redirect rules
	trie.Insert("^/$", "/index.html")
	trie.Insert("^/articles/([0-9]+)$", "/posts/$1")
	trie.Insert("^/users/([0-9]+)$", "/profile/$1")

	// Test matching
	testCases := []string{
		"/",
		"/articles/123",
		"/articles/abc",
		"/users/123",
		"/users/abc",
		"/about-us",
		"/home/page/123",
		"/articles/123/abc",
	}

	for _, testCase := range testCases {
		redirectURL, matched := trie.Match(testCase)
		if matched {
			fmt.Printf("Redirecting %s to %s\n", testCase, redirectURL)
		} else {
			log.Println("No matching rule found for: \n", testCase)
		}
	}
	//log.Println("Starting redirects-traefik-middleware")
	//config := NewAppConfig()
	//
	//authData := api.NewAuthData(config.clientName, config.clientSecret, config.serverURL, config.jwtSecret)
	//graphqlClient := api.NewGraphQLClient(authData)
	//
	//logger := app.NewLogger(config.logFilePath, graphqlClient)
	//logger.SendLogsWeekly()
	//
	//redirectManager := app.NewRedirectManager(dbConnect(config.dbFilePath), graphqlClient)
	//redirectManager.PopulateMapWithDataFromDB()
	//redirectManager.PopulateTrieWithRedirects()
	//
	////Create channels for fetching redirects periodically
	//redirectsCh := make(chan []api.Redirect)
	//errCh := make(chan error)
	//
	//go func() {
	//	defer close(redirectsCh)
	//	defer close(errCh)
	//
	//	redirectManager.FetchRedirectsOverChannel(redirectsCh, errCh)
	//}()
	//go redirectManager.SyncRedirects(redirectsCh, errCh)
	//
	//handleRequests(logger, redirectManager)
}

// Register the handlers
//func handleRequests(logger *app.Logger, redirectManager *app.RedirectManager) {
//	http.HandleFunc("/", handlers.GetRedirectMatch(logger, redirectManager))
//	log.Fatal(http.ListenAndServe(":8081", nil))
//}

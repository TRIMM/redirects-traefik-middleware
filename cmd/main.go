package main

import (
	"database/sql"
	"github.com/joho/godotenv"
	"log"
	api "redneck-traefik-middleware/api/v1"
	"redneck-traefik-middleware/internal/app"
)

func main() {
	// Load env variables
	loadEnv()

	var tokenData = api.NewTokenData()
	var graphqlClient = api.NewGraphQLClient(tokenData)

	var logger = app.NewLogger("requests.log", graphqlClient)
	logger.SendLogsWeekly()

	var redirectManager = app.NewRedirectManager(dbConnect("redirects.db"), graphqlClient, tokenData, logger)
	redirectManager.PopulateMapWithDataFromDB()
	redirectManager.PopulateTrieWithRedirects()

	//Create channels for fetching redirects periodically
	var redirectsCh = make(chan []api.Redirect)
	var errCh = make(chan error)

	go func() {
		defer close(redirectsCh)
		defer close(errCh)

		redirectManager.FetchRedirectsOverChannel(redirectsCh, errCh)
	}()
	redirectManager.SyncRedirects(redirectsCh, errCh)

	// Keep the main goroutine running to let the child goroutines execute
	select {}
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func dbConnect(file string) *sql.DB {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		log.Fatal("Database connection issues: ", err)
	}
	return db
}

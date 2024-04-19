package main

import (
	"database/sql"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	loadEnv()
	var logger = NewLogger("requests.log")
	var redirectManager = NewRedirectManager(dbConnect("redirects.db"), logger)

	//Create channels for fetching redirects periodically
	var redirectsCh = make(chan []Redirect)
	var errCh = make(chan error)

	go func() {
		defer close(redirectsCh)
		defer close(errCh)

		redirectManager.FetchRedirectsOverChannel(redirectsCh, errCh)
	}()

	redirectManager.PopulateMapWithDataFromDB()
	redirectManager.SyncRedirects(redirectsCh, errCh)

	// Start a goroutine to send request logs periodically
	//TODO::see why it does not get here
	go func() {
		logger.SendLogs(redirectManager.tokenData)
	}()

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

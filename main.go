package main

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	loadEnv()
	var redirectManager = NewRedirectManager(dbConnect("redirects.db"))

	// Create channels for fetching redirects periodically
	var redirectsCh = make(chan []Redirect)
	var errCh = make(chan error)

	go fetchRedirectsOverChannel(redirectsCh, errCh)
	redirectManager.PopulateMapWithDataFromDB()
	redirectManager.SyncRedirects(redirectsCh, errCh)

	err := redirectManager.logger.LoadLoggedRequests()
	if err != nil {
		fmt.Println(err)
	}
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

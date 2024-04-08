package main

import (
	"database/sql"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	loadEnv()

	var redirectManager = NewRedirectManager(dbConnect("redirects.db"))
	redirectManager.PopulateMapWithSQLRedirects()

	// Create channels for fetching redirects periodically
	var redirectsCh = make(chan []Redirect)
	var errCh = make(chan error)

	go fetchRedirectsOverChannel(redirectsCh, errCh)
	redirectManager.SyncRedirects(redirectsCh, errCh)
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

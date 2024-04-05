package main

import (
	"github.com/joho/godotenv"
	"log"
)

func main() {
	loadEnv()
	var redirectsCh = make(chan []Redirect)
	var errCh = make(chan error)
	var redirectManager = NewRedirectManager()

	go fetchRedirectsOverChannel(redirectsCh, errCh)
	redirectManager.SyncRedirects(redirectsCh, errCh)
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

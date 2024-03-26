package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func fetchRedirectsOverChannel(redirectsCh chan<- []Redirect, errCh chan<- error) {
	for {
		select {
		case <-time.After(10 * time.Second):
			fetchedRedirects, err := fetchRedirectsQuery()
			if err != nil {
				errCh <- err
			} else {
				// Send fetched redirects over the channel
				redirectsCh <- fetchedRedirects
			}
		}
	}
}

func syncRedirects(redirectsCh <-chan []Redirect, errCh <-chan error) {
	redirectMap := make(map[string]*Redirect)
	//fetchedRedirectsMap := make(map[string]bool)

	var lastSyncTime time.Time

	for {
		select {
		case fetchedRedirects := <-redirectsCh:

			if len(fetchedRedirects) > 0 {
				if lastSyncTime.IsZero() {
					lastSyncTime = time.Now()
				}

				for _, fr := range fetchedRedirects {
					// Check if redirect exists in map
					if r, ok := redirectMap[fr.Id]; ok {
						// Update existing redirect
						if fr.UpdatedAt.After(lastSyncTime) {
							fmt.Println("Needs update")
							*r = fr
						}
					} else {
						// Add new redirect
						redirectMap[fr.Id] = &fr
					}
				}

				lastSyncTime = time.Now()
				fmt.Println("Redirects synced at:", lastSyncTime)
				printRedirects(redirectMap)
			}

		case err := <-errCh:
			log.Println("Error syncing redirects:", err)
		}
	}
}

func printRedirects(redirectMap map[string]*Redirect) {
	fmt.Println("Redirects:")
	for id, r := range redirectMap {
		fmt.Printf("ID: %s, FromUrl: %s, ToUrl: %s, UpdatedAt: %s\n", id, r.FromUrl, r.ToUrl, r.UpdatedAt)
	}
}

func getFullURL(req *http.Request) string {
	var proto = "https://"
	if req.TLS == nil {
		proto = "http://"
	}

	var host = req.URL.Host
	if len(host) == 0 {
		host = req.Host
	}

	var answer = proto + host + req.URL.Path
	return strings.ToLower(answer)
}

//func ServeHTTP(rw http.ResponseWriter, req *http.Request) {
//	var request = getFullURL(req)
//	// Revise
//	var answerUrl = req.URL.Host
//
//	for _, redirect := range redirects {
//		if redirect.FromUrl == request {
//			answerUrl = redirect.ToUrl
//			break
//		}
//	}
//
//	http.Redirect(rw, req, answerUrl, 302)
//}

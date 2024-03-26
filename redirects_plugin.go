package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type RedirectManager struct {
	redirects    map[string]*Redirect
	lastSyncTime time.Time
	//mu        sync.Mutex
}

func NewRedirectManager() *RedirectManager {
	return &RedirectManager{
		redirects:    make(map[string]*Redirect),
		lastSyncTime: time.Time{},
	}
}

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

func (rm *RedirectManager) SyncRedirects(redirectsCh <-chan []Redirect, errCh <-chan error) {
	for {
		select {
		case fetchedRedirects := <-redirectsCh:

			if len(fetchedRedirects) > 0 {
				if rm.lastSyncTime.IsZero() {
					rm.lastSyncTime = time.Now()
				}

				// Initialize a map of ids for quicker lookup
				var fetchedRedirectsIDs = make(map[string]bool)
				for _, fr := range fetchedRedirects {
					fetchedRedirectsIDs[fr.Id] = true
				}

				// Handle deletion
				for id := range rm.redirects {
					if !fetchedRedirectsIDs[id] {
						delete(rm.redirects, id)
						fmt.Println("Redirect deleted:", id)
					}
				}

				for _, fr := range fetchedRedirects {
					// Check if redirect exists in map
					if r, ok := rm.redirects[fr.Id]; ok {
						// Update existing redirect
						if fr.UpdatedAt.After(rm.lastSyncTime) {
							fmt.Println("Redirect updated:", fr.Id)
							*r = fr
						}
					} else {
						// Add new redirect
						rm.redirects[fr.Id] = &fr
						fmt.Println("Redirect added:", fr.Id)
					}
				}

				rm.lastSyncTime = time.Now()
				fmt.Println("Redirects synced at:", rm.lastSyncTime)
				printRedirects(rm.redirects)
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

func (rm *RedirectManager) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var request = getFullURL(req)
	// Revise
	var answerUrl = req.URL.Host

	for _, redirect := range rm.redirects {
		if redirect.FromUrl == request {
			answerUrl = redirect.ToUrl
			break
		}
	}

	http.Redirect(rw, req, answerUrl, 302)
}

package main

import (
	"fmt"
	"log"
	"time"
)

type RedirectManager struct {
	redirects    map[string]*Redirect
	lastSyncTime time.Time
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

				rm.handleOldRedirectsDeletion(&fetchedRedirects)
				rm.handleNewOrUpdatedRedirects(&fetchedRedirects)

				rm.lastSyncTime = time.Now()
				fmt.Println("Redirects synced at:", rm.lastSyncTime)
				printRedirects(rm.redirects)
				rm.PopulateTrieWithRedirects()
			}

		case err := <-errCh:
			log.Println("Error syncing redirects:", err)
		}
	}
}

// Handle deletion
func (rm *RedirectManager) handleOldRedirectsDeletion(fetchedRedirects *[]Redirect) {
	var fetchedRedirectsIDs = initializeRedirectMapIds(*fetchedRedirects)

	for id := range rm.redirects {
		if !fetchedRedirectsIDs[id] {
			delete(rm.redirects, id)
			fmt.Println("Redirect deleted:", id)
		}
	}
}

func (rm *RedirectManager) handleNewOrUpdatedRedirects(fetchedRedirects *[]Redirect) {
	for _, fr := range *fetchedRedirects {
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
}

// Initialize a map of ids for quicker lookup
func initializeRedirectMapIds(fetchedRedirects []Redirect) map[string]bool {
	var fetchedRedirectsIDs = make(map[string]bool)
	for _, fr := range fetchedRedirects {
		fetchedRedirectsIDs[fr.Id] = true
	}

	return fetchedRedirectsIDs
}

func printRedirects(redirectMap map[string]*Redirect) {
	fmt.Println("Redirects:")
	for id, r := range redirectMap {
		fmt.Printf("ID: %s, FromUrl: %s, ToUrl: %s, UpdatedAt: %s\n", id, r.FromUrl, r.ToUrl, r.UpdatedAt)
	}
}

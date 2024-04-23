package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
)

type RedirectManager struct {
	db           *sql.DB
	gqlClient    *GraphQLClient
	redirects    map[string]*Redirect
	trie         *Trie
	tokenData    *TokenData
	logger       *Logger
	lastSyncTime time.Time
}

func NewRedirectManager(db *sql.DB, gqlClient *GraphQLClient, tokenData *TokenData, logger *Logger) *RedirectManager {
	return &RedirectManager{
		db:           db,
		gqlClient:    gqlClient,
		redirects:    make(map[string]*Redirect),
		trie:         NewTrie(),
		tokenData:    tokenData,
		logger:       logger,
		lastSyncTime: time.Time{},
	}
}

func (rm *RedirectManager) FetchRedirectsOverChannel(redirectsCh chan<- []Redirect, errCh chan<- error) {
	for {
		select {
		//The time interval is experimental (for testing). For production change the time accordingly
		case <-time.After(10 * time.Second):
			fetchedRedirects, err := rm.gqlClient.ExecuteRedirectsQuery(rm.tokenData.ClientId)
			if err != nil {
				errCh <- err
			} else {
				// Send fetched redirects over the channel
				redirectsCh <- fetchedRedirects
			}
		}
	}
}

func (rm *RedirectManager) PopulateMapWithDataFromDB() {
	rows, err := rm.db.Query("SELECT * FROM redirects")
	if err != nil {
		log.Println("Error retrieving SQL records:", err)
	}

	defer func() {
		err := rows.Close()
		if err != nil {
			log.Println("Error closing rows:", err)
		}
	}()

	for rows.Next() {
		r := Redirect{}
		err = rows.Scan(&r.Id, &r.FromURL, &r.ToURL, &r.UpdatedAt)
		if err != nil {
			log.Println("Error scanning the SQL rows:", err)
		}
		rm.redirects[r.Id] = &r
	}
}

// SyncRedirects synchronizes the fetched redirects with the redirects map and the sqlite records
func (rm *RedirectManager) SyncRedirects(redirectsCh <-chan []Redirect, errCh <-chan error) {
	for {
		select {
		case fetchedRedirects := <-redirectsCh:
			if len(fetchedRedirects) > 0 {
				if rm.lastSyncTime.IsZero() {
					rm.lastSyncTime = time.Now().UTC()
				}

				rm.HandleOldRedirectsDeletion(&fetchedRedirects)
				rm.HandleNewOrUpdatedRedirects(&fetchedRedirects)
				rm.PopulateTrieWithRedirects()

				rm.lastSyncTime = time.Now().UTC()
				fmt.Println("Redirects synced at:", rm.lastSyncTime)
				printRedirects(rm.redirects)
				fmt.Printf("\n")
			}

		case err := <-errCh:
			log.Println("Error syncing redirects:", err)
		}
	}
}

func (rm *RedirectManager) HandleOldRedirectsDeletion(fetchedRedirects *[]Redirect) {
	var fetchedRedirectsIDs = initializeRedirectMapIds(*fetchedRedirects)

	for id := range rm.redirects {
		if !fetchedRedirectsIDs[id] {
			delete(rm.redirects, id)
			// Delete from the database
			err := rm.DeleteOldRedirect(id)
			if err != nil {
				log.Println("Error deleting old redirects:", err)
			} else {
				fmt.Println("Deleted old redirect:", id)
			}
		}
	}
}

func (rm *RedirectManager) HandleNewOrUpdatedRedirects(fetchedRedirects *[]Redirect) {
	for _, fr := range *fetchedRedirects {
		// Check if redirect exists in map
		if r, ok := rm.redirects[fr.Id]; ok {
			// Update existing redirect
			if fr.UpdatedAt.After(rm.lastSyncTime) {
				*r = fr
				fmt.Println("Redirect updated:", fr.Id)
				// Update the database record
				err := rm.UpsertRedirect(fr)
				if err != nil {
					log.Println("Error updating redirect in the database:", err)
				}
			}
		} else {
			// Add new redirect
			rm.redirects[fr.Id] = &fr
			fmt.Println("Redirect added:", fr.Id)

			// Store the database record
			err := rm.UpsertRedirect(fr)
			if err != nil {
				log.Println("Error storing new redirect in the database:", err)
			}
		}
	}
}

func (rm *RedirectManager) PopulateTrieWithRedirects() {
	rm.trie.Clear()
	for _, r := range rm.redirects {
		rm.trie.Insert(r.FromURL, r.ToURL)
	}
}

func (rm *RedirectManager) UpsertRedirect(r Redirect) error {
	stmt := `
			INSERT INTO redirects (id, fromURL, toURL, updatedAt)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE
			SET fromURL = EXCLUDED.fromURL, toURL = EXCLUDED.toURL, updatedAt = EXCLUDED.updatedAt;
			`

	_, err := rm.db.Exec(stmt, r.Id, r.FromURL, r.ToURL, r.UpdatedAt, r.FromURL, r.ToURL, r.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (rm *RedirectManager) DeleteOldRedirect(id string) error {
	stmt := `DELETE FROM redirects WHERE id = ?;`

	_, err := rm.db.Exec(stmt, id)
	if err != nil {
		return err
	}
	return nil
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
		fmt.Printf("ID: %s, FromURL: %s, ToURL: %s, UpdatedAt: %s\n", id, r.FromURL, r.ToURL, r.UpdatedAt)
	}
}

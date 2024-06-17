package app

import (
	"database/sql"
	"fmt"
	api "github.com/TRIMM/redirects-traefik-middleware/api/v1"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
)

type RedirectManager struct {
	db               *sql.DB
	gqlClient        *api.GraphQLClient
	redirects        map[string]*api.Redirect
	IndexedRedirects *IndexedRedirects
	lastSyncTime     time.Time
}

func NewRedirectManager(db *sql.DB, gqlClient *api.GraphQLClient) *RedirectManager {
	return &RedirectManager{
		db:               db,
		gqlClient:        gqlClient,
		redirects:        make(map[string]*api.Redirect),
		IndexedRedirects: NewIndexedRedirects(),
		lastSyncTime:     time.Time{},
	}
}

func (rm *RedirectManager) FetchRedirectsOverChannel(redirectsCh chan<- []api.Redirect, errCh chan<- error) {
	for {
		select {
		case <-time.After(10 * time.Hour):
			fetchedRedirects, err := rm.gqlClient.ExecuteRedirectsQuery()
			if err != nil {
				errCh <- err
			} else {
				// Send fetched redirects over the channel
				redirectsCh <- fetchedRedirects
			}
		}
	}
}

func (rm *RedirectManager) PopulateMapsWithDataFromDB() {
	_, err := rm.db.Exec(`
		CREATE TABLE IF NOT EXISTS redirects (
		    id TEXT PRIMARY KEY,
		    fromURL TEXT,
		    toURL TEXT,
		    updatedAt date
		)
	`)
	if err != nil {
		log.Println("Error creating redirects table:", err)
		return
	}

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
		r := api.Redirect{}
		err = rows.Scan(&r.Id, &r.FromURL, &r.ToURL, &r.UpdatedAt)
		if err != nil {
			log.Println("Error scanning the SQL rows:", err)
		}
		// Add to redirects map
		rm.redirects[r.Id] = &r
		// Add to IndexedRedirects
		rm.IndexedRedirects.IndexRule(r.FromURL, r.ToURL)
	}
}

// SyncRedirects synchronizes the fetched redirects with the redirects map and the sqlite records
func (rm *RedirectManager) SyncRedirects(redirectsCh <-chan []api.Redirect, errCh <-chan error) {
	for {
		select {
		case fetchedRedirects := <-redirectsCh:
			if len(fetchedRedirects) > 0 {
				if rm.lastSyncTime.IsZero() {
					rm.lastSyncTime = time.Now().UTC()
				}

				rm.HandleOldRedirectsDeletion(&fetchedRedirects)
				rm.HandleNewOrUpdatedRedirects(&fetchedRedirects)

				rm.lastSyncTime = time.Now().UTC()
				fmt.Println("Redirects synced at:", rm.lastSyncTime)
				printRedirects(rm.redirects)
			}
		case err := <-errCh:
			log.Println("Error syncing redirects:", err)
		}
	}
}

func (rm *RedirectManager) HandleOldRedirectsDeletion(fetchedRedirects *[]api.Redirect) {
	var fetchedRedirectsIDs = initializeRedirectMapIds(*fetchedRedirects)

	for id, r := range rm.redirects {
		if !fetchedRedirectsIDs[id] {
			delete(rm.redirects, id)
			// Delete from IndexedRedirects as well
			rm.IndexedRedirects.Delete(r.FromURL)

			// Delete from the database
			err := rm.DeleteOldRedirect(id)
			if err != nil {
				log.Println("Error deleting old redirects:", err)
			} else {
				log.Println("Deleted old redirect:", id)
			}
		}
	}
}

func (rm *RedirectManager) HandleNewOrUpdatedRedirects(fetchedRedirects *[]api.Redirect) {
	for _, fr := range *fetchedRedirects {
		// Check if redirect exists in map
		if r, ok := rm.redirects[fr.Id]; ok {
			if fr.UpdatedAt.After(rm.lastSyncTime) {
				*r = fr

				rm.IndexedRedirects.Update(r.FromURL, r.ToURL)
				log.Println("Redirect updated:", fr.Id)

				err := rm.UpsertRedirect(fr)
				if err != nil {
					log.Println("Error updating redirect in the database:", err)
				}
			}
		} else {
			rm.redirects[fr.Id] = &fr
			rm.IndexedRedirects.IndexRule(fr.FromURL, fr.ToURL)
			log.Println("Redirect added:", fr.Id)

			err := rm.UpsertRedirect(fr)
			if err != nil {
				log.Println("Error storing new redirect in the database:", err)
			}
		}
	}
}

func (rm *RedirectManager) UpsertRedirect(r api.Redirect) error {
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
func initializeRedirectMapIds(fetchedRedirects []api.Redirect) map[string]bool {
	var fetchedRedirectsIDs = make(map[string]bool)
	for _, fr := range fetchedRedirects {
		fetchedRedirectsIDs[fr.Id] = true
	}

	return fetchedRedirectsIDs
}

func printRedirects(redirectMap map[string]*api.Redirect) {
	for id, r := range redirectMap {
		fmt.Printf("ID: %s, FromURL: %s, ToURL: %s, UpdatedAt: %s\n", id, r.FromURL, r.ToURL, r.UpdatedAt)
	}
	fmt.Printf("\n")
}

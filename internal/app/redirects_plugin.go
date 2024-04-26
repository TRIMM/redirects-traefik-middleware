package app

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	api "redirects-traefik-middleware/api/v1"
	"strings"
)

type Config struct {
	ClientName   string `json:"clientName,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
	ServerURL    string `json:"serverURL,omitempty"`
	LogFilePath  string `json:"logFilePath,omitempty"`
	DBFilePath   string `json:"dbFilePath,omitempty"`
}

func CreateConfig() *Config {
	return &Config{}
}

type RedirectsPlugin struct {
	next             http.Handler
	name             string
	redirectsManager *RedirectManager
	logger           *Logger
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	authData := api.NewAuthData(config.ClientName, config.ClientSecret, config.ServerURL)
	graphqlClient := api.NewGraphQLClient(authData)

	logger := NewLogger(config.LogFilePath, graphqlClient)
	logger.SendLogsWeekly()

	var redirectManager = NewRedirectManager(dbConnect(config.DBFilePath), graphqlClient)
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

	return &RedirectsPlugin{
		next:             next,
		redirectsManager: redirectManager,
		logger:           logger,
	}, nil
}

/*
ServeHTTP intercepts a request and matches it against the existing rules presented in the Trie Data structure
If a match is found, it redirects accordingly
*/
func (rp *RedirectsPlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var request = getFullURL(req)
	if err := rp.logger.LogRequest(request); err != nil {
		log.Println("Failed to log request to file: ", err)
	}

	redirectURL, ok := rp.redirectsManager.trie.Match(request)
	if !ok {
		log.Println("No matching redirect rule found!")
	}

	http.Redirect(rw, req, redirectURL, http.StatusFound)
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

func dbConnect(file string) *sql.DB {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		log.Fatal("Database connection issues: ", err)
	}
	return db
}

package handlers

import (
	"fmt"
	"github.com/TRIMM/redirects-traefik-middleware/internal/app"
	"io"
	"log"
	"net/http"
)

func GetRedirectMatch(logger *app.Logger, redirectManager *app.RedirectManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestBody, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}

		request := string(requestBody)
		// Logging the incoming requests
		if err := logger.LogRequest(request); err != nil {
			log.Println("Failed to log request to file: ", err)
		}

		// Matching against the defined redirects
		redirectURL, ok := redirectManager.Trie.Match(request)
		if !ok {
			redirectURL = "@empty"
		}

		// Write the redirect URL to the response
		_, err = fmt.Fprintf(w, "%s", redirectURL)
		if err != nil {
			log.Println("Failed to write response:", err)
		}
	}
}

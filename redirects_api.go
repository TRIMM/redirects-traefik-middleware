package main

import (
	"context"
	"fmt"
	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
	"log"
	"os"
	"time"
)

type Redirect struct {
	Id        string    `graphql:"id"`
	FromUrl   string    `graphql:"fromURL"`
	ToUrl     string    `graphql:"toURL"`
	UpdatedAt time.Time `graphql:"updatedAt"`
}

func fetchRedirects(td *TokenData) ([]Redirect, error) {
	token, err := td.GetToken()
	if err != nil {
		log.Println("Failed to get token:", err)
		return nil, err
	}

	return executeRedirectsQuery(token, td.ClientId)
}

func executeRedirectsQuery(token, clientId string) ([]Redirect, error) {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	var client = graphql.NewClient(fmt.Sprintf("%s/graphql", os.Getenv("SERVER_URL")), httpClient)

	vars := map[string]interface{}{
		"clientId": graphql.String(clientId),
	}

	var redirectsQuery struct {
		Redirects []Redirect `graphql:"redirects(clientId: $clientId)"`
	}

	err := client.Query(context.Background(), &redirectsQuery, vars)
	if err != nil {
		log.Println("GraphQL server not reachable!", err)
		return nil, err
	}

	return redirectsQuery.Redirects, nil
}

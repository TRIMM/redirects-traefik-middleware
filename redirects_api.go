package main

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/shurcooL/graphql"
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

var redirectsQuery struct {
	Redirects []Redirect `graphql:"redirects(clientId: $clientId)"`
}

func fetchRedirectsQuery() ([]Redirect, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	graphqlUrl := os.Getenv("GRAPHQL_SERVER")
	clientId := os.Getenv("CLIENT_ID")
	client := graphql.NewClient(graphqlUrl, nil)

	vars := map[string]interface{}{
		"clientId": graphql.String(clientId),
	}

	err = client.Query(context.Background(), &redirectsQuery, vars)
	if err != nil {
		log.Println("GraphQL server not reachable!", err)
		return nil, err
	}

	return redirectsQuery.Redirects, nil
}

package main

import (
	"context"
	"github.com/shurcooL/graphql"
	"log"
	"time"
)

type Redirect struct {
	Id        string    `graphql:"id"`
	FromURL   string    `graphql:"fromURL"`
	ToURL     string    `graphql:"toURL"`
	UpdatedAt time.Time `graphql:"updatedAt"`
}

func (gql *GraphQLClient) executeRedirectsQuery(clientId string) ([]Redirect, error) {
	var redirectsQuery struct {
		Redirects []Redirect `graphql:"redirects(clientId: $clientId)"`
	}

	vars := map[string]interface{}{
		"clientId": graphql.String(clientId),
	}

	err := gql.client.Query(context.Background(), &redirectsQuery, vars)
	if err != nil {
		log.Println("GraphQL server not reachable!", err)
		return nil, err
	}

	return redirectsQuery.Redirects, nil
}

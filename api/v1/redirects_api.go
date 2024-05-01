package v1

import (
	"context"
	"fmt"
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

func (gql *GraphQLClient) ExecuteRedirectsQuery() ([]Redirect, error) {
	var redirectsQuery struct {
		Redirects []Redirect `graphql:"redirects(clientId: $clientId)"`
	}

	client := gql.GetClient()
	if client == nil {
		return nil, fmt.Errorf("GraphQL client not initialized")
	}
	vars := map[string]interface{}{
		"clientId": graphql.String(gql.TokenData.ClientId),
	}

	err := client.Query(context.Background(), &redirectsQuery, vars)
	if err != nil {
		log.Println("GraphQL server not reachable!", err)
		return nil, err
	}

	return redirectsQuery.Redirects, nil
}

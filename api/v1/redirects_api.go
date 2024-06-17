package v1

import (
	"context"
	"fmt"
	"github.com/shurcooL/graphql"
	"log"
	"os"
	"time"
)

type Redirect struct {
	Id        string    `graphql:"id"`
	FromURL   string    `graphql:"fromURL"`
	ToURL     string    `graphql:"toURL"`
	UpdatedAt time.Time `graphql:"updatedAt"`
}

type PageInfo struct {
	HasNextPage bool
	EndCursor   string
}

type RedirectEdge struct {
	Node   Redirect
	Cursor string
}

type RedirectConnection struct {
	Edges    []RedirectEdge
	PageInfo PageInfo
}

func (gql *GraphQLClient) ExecuteRedirectsQuery() ([]Redirect, error) {
	var allRedirects []Redirect
	var endCursor string
	for {
		redirects, pageInfo, err := gql.fetchRedirectsPage(endCursor)
		if err != nil {
			return nil, err
		}
		allRedirects = append(allRedirects, redirects...)
		if !pageInfo.HasNextPage {
			break
		}
		endCursor = pageInfo.EndCursor
		time.Sleep(100 * time.Millisecond)
	}
	return allRedirects, nil
}

func (gql *GraphQLClient) fetchRedirectsPage(cursor string) ([]Redirect, PageInfo, error) {
	var query struct {
		Redirects RedirectConnection `graphql:"redirects(hostId: $hostId, first: 100, after: $cursor)"`
	}

	vars := map[string]interface{}{
		"hostId": graphql.String(os.Getenv("HOST_ID")),
		"cursor": graphql.String(cursor),
	}

	client := gql.GetClient()
	if client == nil {
		return nil, PageInfo{}, fmt.Errorf("GraphQL client not initialized")
	}

	err := client.Query(context.Background(), &query, vars)
	if err != nil {
		log.Println("GraphQL server not reachable!", err)
		return nil, PageInfo{}, err
	}

	var redirects []Redirect
	for _, edge := range query.Redirects.Edges {
		redirects = append(redirects, edge.Node)
	}

	return redirects, query.Redirects.PageInfo, nil
}

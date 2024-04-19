package main

import (
	"context"
	"log"
	"time"
)

type LogRequestsInput struct {
	RequestURL string    `json:"requestURL"`
	HitTime    time.Time `json:"hitTime"`
}

func sendLogs(td *TokenData) {
	token, err := td.GetToken()
	if err != nil {
		log.Println("Failed to get token:", err)
	}

	executeLogRequestsMutation(token)
}

func executeLogRequestsMutation(token string) {
	var logMutation struct {
		LogResponse struct {
			Success bool   `graphql:"success"`
			Message string `graphql:"message"`
		} `graphql:"logRequests(logRequestsInput: $logRequestsInput)"`
	}

	vars := map[string]interface{}{
		"logRequestsInput": []LogRequestsInput{
			{
				RequestURL: "http://localhost:8080/",
				HitTime:    time.Now(),
			},
			{
				RequestURL: "http://localhost:8081/test",
				HitTime:    time.Now(),
			},
		},
	}

	var client = newGraphQLClient(token)
	err := client.Mutate(context.Background(), &logMutation, vars)
	if err != nil {
		log.Println("GraphQL server not reachable!", err)
	}

	log.Println("Mutation response:", logMutation.LogResponse)

}

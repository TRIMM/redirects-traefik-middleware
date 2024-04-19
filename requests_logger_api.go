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

type LogResponse struct {
	Success bool   `graphql:"success"`
	Message string `graphql:"message"`
}

func executeLogRequestsMutation(token string, requestsMap *map[string]time.Time) (LogResponse, error) {
	var logMutation struct {
		LogResponse `graphql:"logRequests(logRequestsInput: $logRequestsInput)"`
	}

	var logsInput []LogRequestsInput

	for key, val := range *requestsMap {
		logsInput = append(logsInput, LogRequestsInput{
			RequestURL: key,
			HitTime:    val,
		})
	}

	vars := map[string]interface{}{
		"logRequestsInput": logsInput,
	}

	var client = newGraphQLClient(token)
	err := client.Mutate(context.Background(), &logMutation, vars)
	if err != nil {
		log.Println("GraphQL server not reachable!", err)
		return LogResponse{}, err
	}

	return logMutation.LogResponse, nil
}

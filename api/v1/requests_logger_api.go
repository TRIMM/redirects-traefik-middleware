package v1

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

func (gql *GraphQLClient) ExecuteLogRequestsMutation(requestsMap *map[string]time.Time) (LogResponse, error) {
	var logMutation struct {
		LogResponse `graphql:"logRequests(logRequestsInput: $logRequestsInput)"`
	}

	var logsInput []LogRequestsInput

	// populate the logs input for the mutation
	for key, val := range *requestsMap {
		logsInput = append(logsInput, LogRequestsInput{
			RequestURL: key,
			HitTime:    val,
		})
	}

	vars := map[string]interface{}{
		"logRequestsInput": logsInput,
	}

	err := gql.GetClient().Mutate(context.Background(), &logMutation, vars)
	if err != nil {
		log.Println("GraphQL server not reachable!", err)
		return LogResponse{}, err
	}

	return logMutation.LogResponse, nil
}

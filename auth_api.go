package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type AuthBody struct {
	ClientName   string `json:"clientName"`
	ClientSecret string `json:"clientSecret"`
}

func NewAuthBody() *AuthBody {
	return &AuthBody{
		ClientName:   os.Getenv("CLIENT_NAME"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
	}
}

func (authBody *AuthBody) Auth() (string, error) {
	marshalled, err := json.Marshal(authBody)
	if err != nil {
		return "", fmt.Errorf("error while building the auth request: %v", err)
	}

	authEndpoint := fmt.Sprintf("%s/auth", os.Getenv("SERVER_URL"))
	res, err := http.Post(authEndpoint, "application/json", bytes.NewReader(marshalled))
	if err != nil {
		return "", fmt.Errorf("error while authenticating: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("authentication failed with status code %d", res.StatusCode)
	}

	var tokenResponse struct {
		Token string `json:"access_token"`
	}

	if err := json.NewDecoder(res.Body).Decode(&tokenResponse); err != nil {
		return "", fmt.Errorf("failed to decode token response: %v", err)
	}

	return tokenResponse.Token, nil
}

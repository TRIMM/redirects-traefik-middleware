package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"os"
	"time"

	"gopkg.in/square/go-jose.v2/jwt"
)

type AuthData struct {
	ClientName   string `json:"clientName"`
	ClientSecret string `json:"clientSecret"`
}

func NewAuthData() *AuthData {
	return &AuthData{
		ClientName:   os.Getenv("CLIENT_NAME"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
	}
}

type TokenData struct {
	Token    string
	ClientId string
}

func NewTokenData() *TokenData {
	return &TokenData{}
}

type GraphQLClient struct {
	client *graphql.Client
}

func NewGraphQLClient(tokenData *TokenData) *GraphQLClient {
	token, err := tokenData.GetToken()
	if err != nil {
		log.Println("Failed to get token:", err)
	}

	return &GraphQLClient{
		client: createGraphQLClient(token),
	}
}

func createGraphQLClient(token string) *graphql.Client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	var client = graphql.NewClient(fmt.Sprintf("%s/graphql", os.Getenv("SERVER_URL")), httpClient)

	return client
}

func (authData *AuthData) Auth() (string, error) {
	marshalled, err := json.Marshal(authData)
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

func (td *TokenData) GetToken() (string, error) {
	if len(td.Token) == 0 || isTokenExpired(td.Token) {
		auth := NewAuthData()
		token, err := auth.Auth()
		if err != nil {
			log.Println("Authentication failed:", err)
			return "", err
		}
		td.setClientIdFromClaims(token)
		td.Token = token

		return token, nil
	}

	return td.Token, nil
}

func isTokenExpired(tokenString string) bool {
	var claims = getClaimsFromToken(tokenString)

	return claims.Expiry.Time().Before(time.Now())
}

func (td *TokenData) setClientIdFromClaims(tokenString string) {
	var claims = getClaimsFromToken(tokenString)
	td.ClientId = claims.Subject
}

func getClaimsFromToken(tokenString string) jwt.Claims {
	parsedToken, err := jwt.ParseSigned(tokenString)
	if err != nil {
		log.Println("Failed to parse token:", err)
	}

	var claims jwt.Claims
	if err := parsedToken.UnsafeClaimsWithoutVerification(&claims); err != nil {
		log.Println("Failed to extract claims from token:", err)
	}

	return claims
}

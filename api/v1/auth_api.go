package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"time"

	"gopkg.in/square/go-jose.v2/jwt"
)

type AuthData struct {
	ClientName   string `json:"clientName"`
	ClientSecret string `json:"clientSecret"`
	ServerURL    string `json:"serverURL"`
}

type TokenData struct {
	Token    string
	ClientId string
}

type GraphQLClient struct {
	client *graphql.Client
}

func NewAuthData(clientName string, clientSecret string, serverURL string) *AuthData {
	return &AuthData{
		ClientName:   clientName,
		ClientSecret: clientSecret,
		ServerURL:    serverURL,
	}
}

func NewTokenData() *TokenData {
	return &TokenData{}
}

func NewGraphQLClient(tokenData *TokenData, authData *AuthData) *GraphQLClient {
	token, err := tokenData.GetToken(authData)
	if err != nil {
		log.Println("Failed to get token:", err)
	}

	return &GraphQLClient{
		client: createGraphQLClient(token, authData),
	}
}

func createGraphQLClient(token string, authData *AuthData) *graphql.Client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	var client = graphql.NewClient(fmt.Sprintf("%s/graphql", authData.ServerURL), httpClient)

	return client
}

func (authData *AuthData) Auth() (string, error) {
	marshalled, err := json.Marshal(authData)
	if err != nil {
		return "", fmt.Errorf("error while building the auth request: %v", err)
	}

	authEndpoint := fmt.Sprintf("%s/auth", authData.ServerURL)
	res, err := http.Post(authEndpoint, "application/json", bytes.NewReader(marshalled))
	if err != nil {
		return "", fmt.Errorf("error while authenticating: %v", err)
	}

	defer func() {
		err := res.Body.Close()
		if err != nil {
			log.Println("Error closing file:", err)
		}
	}()

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

func (td *TokenData) GetToken(authData *AuthData) (string, error) {
	if len(td.Token) == 0 || isTokenExpired(td.Token) {
		token, err := authData.Auth()
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

func (td *TokenData) setClientIdFromClaims(tokenString string) {
	var claims = getClaimsFromToken(tokenString)
	td.ClientId = claims.Subject
}

func isTokenExpired(tokenString string) bool {
	var claims = getClaimsFromToken(tokenString)

	return claims.Expiry.Time().Before(time.Now())
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

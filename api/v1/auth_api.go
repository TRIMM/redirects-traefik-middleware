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
	client    *graphql.Client
	TokenData *TokenData
	authData  *AuthData
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

func NewGraphQLClient(authData *AuthData) *GraphQLClient {
	return &GraphQLClient{
		TokenData: NewTokenData(),
		authData:  authData,
	}
}

// GetClient makes sure the access token is refreshed
func (gql *GraphQLClient) GetClient() *graphql.Client {
	if gql.TokenData.Token == "" || isTokenExpired(gql.TokenData.Token) {
		err := gql.GetNewAccessToken()
		if err != nil {
			log.Println("Failed to get token:", err)
			return nil
		}
	}

	return gql.client
}

func (gql *GraphQLClient) GetNewAccessToken() error {
	token, err := gql.Auth()
	if err != nil {
		log.Println("Authentication failed:", err)
		return err
	}
	gql.UpdateGraphQLClient(token)
	gql.SetClientIdFromClaims(token)
	gql.TokenData.Token = token

	return nil
}

func (gql *GraphQLClient) Auth() (string, error) {
	marshalled, err := json.Marshal(gql.authData)
	if err != nil {
		return "", fmt.Errorf("error while building the auth request: %v", err)
	}

	authEndpoint := fmt.Sprintf("%s/auth", gql.authData.ServerURL)
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

func (gql *GraphQLClient) UpdateGraphQLClient(token string) {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	gql.client = graphql.NewClient(fmt.Sprintf("%s/graphql", gql.authData.ServerURL), httpClient)
}

func (gql *GraphQLClient) SetClientIdFromClaims(tokenString string) {
	claims := getClaimsFromToken(tokenString)
	gql.TokenData.ClientId = claims.Subject
}

func isTokenExpired(tokenString string) bool {
	claims := getClaimsFromToken(tokenString)

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

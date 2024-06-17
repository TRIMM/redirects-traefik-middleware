package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"time"
)

type AuthData struct {
	ClientName   string `json:"clientName"`
	ClientSecret string `json:"clientSecret"`
	ServerURL    string `json:"serverURL"`
	JwtSecret    string `json:"jwtSecret"`
}

type TokenData struct {
	Token string
}

type GraphQLClient struct {
	client    *graphql.Client
	TokenData *TokenData
	authData  *AuthData
}

func NewAuthData(clientName string, clientSecret string, serverURL string, jwtSecret string) *AuthData {
	return &AuthData{
		ClientName:   clientName,
		ClientSecret: clientSecret,
		ServerURL:    serverURL,
		JwtSecret:    jwtSecret,
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
	if gql.TokenData.Token == "" || gql.isTokenExpired(gql.TokenData.Token) {
		err := gql.getNewAccessToken()
		if err != nil {
			log.Println("Failed to get token:", err)
			return nil
		}
	}

	return gql.client
}

func (gql *GraphQLClient) getNewAccessToken() error {
	token, err := gql.auth()
	if err != nil {
		log.Println("Authentication failed:", err)
		return err
	}
	gql.updateGraphQLClient(token)
	gql.TokenData.Token = token

	return nil
}

func (gql *GraphQLClient) auth() (string, error) {
	marshalled, err := json.Marshal(struct {
		ClientName   string `json:"clientName"`
		ClientSecret string `json:"clientSecret"`
	}{
		gql.authData.ClientName,
		gql.authData.ClientSecret,
	})
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
			log.Println("Error closing response body: ", err)
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

func (gql *GraphQLClient) updateGraphQLClient(token string) {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	gql.client = graphql.NewClient(fmt.Sprintf("%s/graphql", gql.authData.ServerURL), httpClient)
}

func (gql *GraphQLClient) isTokenExpired(tokenString string) bool {
	claims, err := gql.parseToken(tokenString)
	if err != nil {
		log.Println("Error parsing token:", err)
		return true
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		log.Println("Expiration time not found in token claims")
		return true
	}

	expiryTime := time.Unix(int64(exp), 0)
	return expiryTime.Before(time.Now().Add(10 * time.Second))
}

func (gql *GraphQLClient) parseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(gql.authData.JwtSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	return claims, nil
}

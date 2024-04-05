package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"gopkg.in/square/go-jose.v2/jwt"
)

type Auth struct {
	ClientName   string `json:"clientName"`
	ClientSecret string `json:"clientSecret"`
}

func NewAuth() *Auth {
	return &Auth{
		ClientName:   os.Getenv("CLIENT_NAME"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
	}
}

type TokenData struct {
	Token string
}

func NewTokenData() *TokenData {
	return &TokenData{}
}

func (authBody *Auth) Auth() (string, error) {

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

func (td *TokenData) GetToken() (string, error) {

	if len(td.Token) == 0 || isTokenExpired(td.Token) {
		authBody := NewAuth()
		token, err := authBody.Auth()
		if err != nil {
			log.Println("Authentication failed:", err)
			return "", err
		}
		td.Token = token
		return token, nil
	}

	return td.Token, nil
}

func isTokenExpired(tokenString string) bool {

	parsedToken, err := jwt.ParseSigned(tokenString)
	if err != nil {
		log.Println("Failed to parse token:", err)
		return true
	}

	var claims jwt.Claims
	if err := parsedToken.UnsafeClaimsWithoutVerification(&claims); err != nil {
		log.Println("Failed to extract claims from token:", err)
		return true
	}

	return claims.Expiry.Time().Before(time.Now())
}

package main

import (
	"database/sql"
	"github.com/joho/godotenv"
	"log"
	"os"
)

type AppConfig struct {
	clientName   string
	clientSecret string
	serverURL    string
	jwtSecret    string
	logFilePath  string
	dbFilePath   string
}

func NewAppConfig() *AppConfig {
	loadEnv()
	return &AppConfig{
		clientName:   os.Getenv("CLIENT_NAME"),
		clientSecret: os.Getenv("CLIENT_SECRET"),
		serverURL:    os.Getenv("SERVER_URL"),
		jwtSecret:    os.Getenv("JWT_SECRET"),
		logFilePath:  os.Getenv("LOG_FILE_PATH"),
		dbFilePath:   os.Getenv("DB_FILE_PATH"),
	}
}

func loadEnv() {
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		return
	}

	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file:", err)
		log.Println("Using default values")
	}
}

func dbConnect(file string) *sql.DB {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		log.Fatal("Database connection issues: ", err)
	}
	return db
}

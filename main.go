package main

import (
	"log"

	"os"

	"github.com/alex-emery/mailfeed/mail"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	username := os.Getenv("EMAIL")
	password := os.Getenv("PASSWORD")
	domain := os.Getenv("SERVER")

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("failed to create logger", err)
	}

	f, err := mail.New(logger, domain, username, password)
	if err != nil {
		log.Fatal("failed to create mail fetcher", err)
	}

	defer f.Close()
	f.SeqNum = 6155
	f.Fetch()
}

package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"

	"os"

	"github.com/alex-emery/mailfeed/mail"
	"github.com/alex-emery/mailfeed/newsletter"
	"github.com/alex-emery/mailfeed/rss"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	emailUsername := os.Getenv("EMAIL_USERNAME")
	emailPassword := os.Getenv("EMAIL_PASSWORD")
	emailServer := os.Getenv("EMAIL_SERVER")

	emailID := os.Getenv("EMAIL_ID")
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("failed to create logger", err)
	}

	feedChan := make(chan *newsletter.NewsLetter)
	m, err := mail.New(logger, emailServer, emailUsername, emailPassword, emailID, feedChan)
	if err != nil {
		log.Fatal("failed to create mail fetcher", err)
	}

	defer m.Close()

	go m.StartFetch()

	server := rss.New(logger, feedChan)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			fmt.Println("Error:", err)
		}
	}()

	// handle shutdown signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	err = server.Shutdown(context.Background())
	if err != nil {
		fmt.Println("Error:", err)
	}

}

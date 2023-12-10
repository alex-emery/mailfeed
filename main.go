package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"

	"os"

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

	feedChan := make(chan *NewsLetter)
	m, err := NewMail(logger, emailServer, emailUsername, emailPassword, emailID, feedChan)
	if err != nil {
		log.Fatal("failed to create mail fetcher", err)
	}

	defer m.Close()

	go m.StartFetch()

	feed := NewFeed("Good Title")
	go func() {
		for {
			appendToFood(feed, <-feedChan)
		}
	}()

	http.HandleFunc("/rss", func(w http.ResponseWriter, r *http.Request) {
		content, err := feed.ToRss()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/rss+xml")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Write([]byte(content))
	})

	port := 8080
	fmt.Printf("Server listening on :%d\n", port)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: nil,
	}

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

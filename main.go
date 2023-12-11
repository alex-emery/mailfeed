package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"

	"os"

	"github.com/alex-emery/mailfeed/mail"
	"github.com/alex-emery/mailfeed/newsletter"
	"github.com/alex-emery/mailfeed/rss"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
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

	handler := rss.New(logger, feedChan)

	port := 8080

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/rss", handler.GetFeed)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	go func() {
		logger.Info(fmt.Sprintf("RSS serving on :%d\n", port))
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

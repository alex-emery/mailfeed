package main

import (
	"flag"
	"log"
	"os/signal"

	"os"

	"github.com/alex-emery/mailfeed/internal/service"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	port := flag.String("port", "8080", "port to run server on")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	emailUsername := os.Getenv("EMAIL_USERNAME")
	emailPassword := os.Getenv("EMAIL_PASSWORD")
	emailServer := os.Getenv("EMAIL_SERVER")

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("failed to create logger", err)
	}

	defer logger.Sync()

	svc, err := service.New(logger, emailServer, emailUsername, emailPassword, *port)
	if err != nil {
		logger.Fatal("failed to create service", zap.Error(err))
	}

	go func() {
		if err := svc.Start(); err != nil {
			logger.Fatal("failed to start service", zap.Error(err))
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan

	if err := svc.Stop(); err != nil {
		log.Fatal("failed to shutdown service", err)
	}
}

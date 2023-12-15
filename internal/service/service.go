package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alex-emery/mailfeed/database"
	"github.com/alex-emery/mailfeed/mail"
	"github.com/alex-emery/mailfeed/newsletter"
	"github.com/alex-emery/mailfeed/rss"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
	"moul.io/chizap"
)

type Service struct {
	httpServer *http.Server
	mail       *mail.Mail
	logger     *zap.Logger
}

func New(logger *zap.Logger, emailServer, emailUsername, emailPassword, port string) (Service, error) {
	feedChan := make(chan *newsletter.NewsLetter)
	db, err := database.New(logger, "mailfeed.db")
	if err != nil {
		return Service{}, fmt.Errorf("failed to create database: %w", err)
	}

	m, err := mail.New(logger, emailServer, emailUsername, emailPassword, &db, feedChan)
	if err != nil {
		return Service{}, fmt.Errorf("failed to create mail fetcher: %w", err)
	}

	go m.StartFetch()

	rss, err := rss.New(logger, &db, feedChan)
	if err != nil {
		return Service{}, fmt.Errorf("failed to create rss server: %w", err)
	}

	r := chi.NewRouter()
	r.Use(chizap.New(logger, &chizap.Opts{
		WithReferer:   true,
		WithUserAgent: true,
	}))

	r.Post("/rss", rss.CreateFeed)
	r.Get("/rss/{id}", rss.GetFeed)

	return Service{
		mail: m,
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%s", port),
			Handler: r,
		},
		logger: logger,
	}, nil
}

func (svc *Service) Start() error {
	if err := svc.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start http server: %w", err)
	}

	return nil
}

func (svc *Service) Stop() error {
	svc.mail.Close()
	return svc.httpServer.Shutdown(context.Background())
}

package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/alex-emery/mailfeed/database"
	"github.com/alex-emery/mailfeed/internal/website"
	"github.com/alex-emery/mailfeed/mail"
	"github.com/alex-emery/mailfeed/newsletter"
	"github.com/alex-emery/mailfeed/rss"
	"github.com/go-chi/chi"
	"github.com/go-chi/httprate"
	"go.uber.org/zap"
	"moul.io/chizap"
)

type Service struct {
	httpServer *http.Server
	mail       *mail.Mail
	logger     *zap.Logger
}

type ServiceOptions struct {
	EmailServer   string
	EmailUsername string
	EmailPassword string
	DBPath        string
	Port          string
	Domain        string
}

func New(logger *zap.Logger, options ServiceOptions) (Service, error) {
	feedChan := make(chan *newsletter.NewsLetter)
	db, err := database.New(logger, options.DBPath)
	if err != nil {
		return Service{}, fmt.Errorf("failed to create database: %w", err)
	}

	m, err := mail.New(logger, options.EmailServer, options.EmailUsername, options.EmailPassword, &db, feedChan)
	if err != nil {
		return Service{}, fmt.Errorf("failed to create mail fetcher: %w", err)
	}

	go m.StartFetch()

	rss, err := rss.New(logger, &db, feedChan, options.Domain)
	if err != nil {
		return Service{}, fmt.Errorf("failed to create rss server: %w", err)
	}

	r := chi.NewRouter()
	r.Use(chizap.New(logger, &chizap.Opts{
		WithReferer:   true,
		WithUserAgent: true,
	}))

	r.Get("/", website.Serve)

	r.Route("/rss", func(r chi.Router) {
		r.Use(httprate.LimitByIP(2, 1*time.Minute))
		r.Post("/", rss.CreateFeed)
		r.Get("/{id}", rss.GetFeed)
	})

	return Service{
		mail: m,
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%s", options.Port),
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

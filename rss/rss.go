package rss

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/alex-emery/mailfeed/newsletter"
	"github.com/gorilla/feeds"
	"go.uber.org/zap"
)

func NewFeed(title string) *feeds.Feed {
	now := time.Now()
	feed := &feeds.Feed{
		Title:   title,
		Author:  &feeds.Author{Name: "Mail Feed", Email: "feed@mailfeed.io"},
		Created: now,
	}

	return feed
}

func (s *Server) AddToFeed(letter *newsletter.NewsLetter) {
	s.logger.Info("Adding to feed", zap.String("subject", letter.Subject))

	now := time.Now()
	s.feed.Items = append(s.feed.Items, &feeds.Item{
		Title:       letter.Subject,
		Description: letter.Body,
		Created:     now,
	})
}

type Server struct {
	feed     *feeds.Feed
	server   *http.Server
	logger   *zap.Logger
	feedChan <-chan *newsletter.NewsLetter
}

func New(logger *zap.Logger, feedChan <-chan *newsletter.NewsLetter) *Server {
	feed := NewFeed("Good Title")

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
	logger.Info(fmt.Sprintf("RSS serving on :%d\n", port))
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: nil,
	}

	return &Server{
		feed:     feed,
		server:   server,
		logger:   logger,
		feedChan: feedChan,
	}
}

func (s *Server) ListenAndServe() error {
	go func() {
		s.AddToFeed(<-s.feedChan)
	}()

	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

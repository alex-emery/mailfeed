package rss

import (
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
	logger   *zap.Logger
	feedChan <-chan *newsletter.NewsLetter
}

func (s *Server) GetFeed(w http.ResponseWriter, r *http.Request) {
	content, err := s.feed.ToRss()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/rss+xml")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Write([]byte(content))
}

func New(logger *zap.Logger, feedChan <-chan *newsletter.NewsLetter) *Server {
	feed := NewFeed("Good Title")

	// load feed in from DB.
	s := &Server{
		feed:     feed,
		logger:   logger,
		feedChan: feedChan,
	}
	go func() {
		s.AddToFeed(<-s.feedChan)
	}()

	return s
}

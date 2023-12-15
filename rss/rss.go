package rss

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/alex-emery/mailfeed/database"
	"github.com/alex-emery/mailfeed/database/sqlc"
	"github.com/alex-emery/mailfeed/newsletter"
	"github.com/go-chi/chi"
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
	if _, ok := s.feeds[letter.Inbox]; !ok {
		feed, err := s.db.GetFeed(context.Background(), letter.Inbox)
		if err != nil {
			s.logger.Error("Error getting inbox", zap.Error(err))
		}

		s.feeds[letter.Inbox] = NewFeed(feed.Name)
	}

	date := letter.Date.Format("2006-01-02 15:04:05")
	_, err := s.db.CreateFeedItem(context.Background(), sqlc.CreateFeedItemParams{
		FeedID:  letter.Inbox,
		Subject: letter.Subject,
		Body:    letter.Body,
		Date:    date,
	})

	if err != nil {
		s.logger.Error("Error creating feed item", zap.Error(err))
		return
	}

	s.feeds[letter.Inbox].Items = append(s.feeds[letter.Inbox].Items, &feeds.Item{
		Title:       letter.Subject,
		Description: letter.Body,
		Created:     letter.Date,
	})
}

type Server struct {
	feeds    map[string]*feeds.Feed
	logger   *zap.Logger
	feedChan <-chan *newsletter.NewsLetter
	db       *database.Database
}

type CreateFeedRequest struct {
	Name string
}

// Creates a feed. Feeds consist of a name and an ID.
// The name is used as the feeds title, and is human friendly.
// The ID is used as the email username and is random.
func (s *Server) CreateFeed(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	req := CreateFeedRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(500)
	}

	if req.Name == "" {
		w.WriteHeader(400)
		return
	}

	feed, err := s.db.CreateFeed(r.Context(), sqlc.CreateFeedParams{
		ID:   GenerateRandomString(4),
		Name: req.Name,
	})
	if err != nil {
		w.WriteHeader(500)
	}

	resp := fmt.Sprintf("{\"email\": \"%s\"}", feed.ID)

	w.Write([]byte(resp))
}

// Gets a feed for a given id, which is the username part of the email address.
func (s *Server) GetFeed(w http.ResponseWriter, r *http.Request) {
	inboxID := chi.URLParam(r, "id")

	if s.feeds[inboxID] == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	content, err := s.feeds[inboxID].ToRss()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/rss+xml")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	_, err = w.Write([]byte(content))
	if err != nil {
		s.logger.Error("Error writing response", zap.Error(err))
	}
}

func New(logger *zap.Logger, db *database.Database, feedChan <-chan *newsletter.NewsLetter) (*Server, error) {
	s := &Server{
		feeds:    make(map[string]*feeds.Feed),
		logger:   logger,
		feedChan: feedChan,
		db:       db,
	}

	rssFeeds, err := db.ListFeeds(context.Background())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to initialise, error listing inboxes: %v", err)
	}

	for _, inbox := range rssFeeds {
		s.logger.Debug("Initialising feed", zap.String("name", inbox.Name), zap.String("id", inbox.ID))
		feed := NewFeed(inbox.Name)
		s.feeds[inbox.ID] = feed

		items, err := db.ListFeedItems(context.Background(), inbox.ID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("failed to initialise, error listing feed items: %v", err)
		}
		for _, item := range items {
			s.logger.Debug("Adding item to feed", zap.String("subject", item.Subject))
			date, err := time.Parse("2006-01-02 15:04:05", item.Date)
			if err != nil {
				return nil, fmt.Errorf("failed to parse date: %v", err)
			}
			feed.Items = append(feed.Items, &feeds.Item{
				Title:       item.Subject,
				Description: item.Body,
				Created:     date,
			})
		}

	}

	go func() {
		s.AddToFeed(<-s.feedChan)
	}()

	return s, nil
}

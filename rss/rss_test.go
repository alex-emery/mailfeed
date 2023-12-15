package rss

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alex-emery/mailfeed/database"
	"github.com/alex-emery/mailfeed/newsletter"
	"github.com/go-chi/chi"
	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGetFeed(t *testing.T) {
	w := httptest.NewRecorder()

	r, err := http.NewRequest("GET", "/123", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	feed := &feeds.Feed{
		Title: "Test Feed",
	}

	feeds := map[string]*feeds.Feed{
		"123": feed,
	}

	logger := zap.NewNop()

	feedChan := make(chan *newsletter.NewsLetter)

	db, err := database.New(logger, ":memory:")
	require.NoError(t, err)

	s := &Server{
		feeds:    feeds,
		logger:   logger,
		feedChan: feedChan,
		db:       &db,
	}

	s.AddToFeed(&newsletter.NewsLetter{
		Inbox:   "123",
		Subject: "Test Subject",
		Body:    "Test Body",
	})

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "123")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	s.GetFeed(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	fp := gofeed.NewParser()
	response, _ := fp.Parse(w.Body)

	require.Equal(t, "Test Feed", response.Title)
	require.Equal(t, "Test Subject", response.Items[0].Title)
	require.Equal(t, "Test Body", response.Items[0].Description)
}

func TestCreateFeed(t *testing.T) {
	w := httptest.NewRecorder()

	feedName := "Test Feed"
	reqBody := CreateFeedRequest{
		Name: feedName,
	}
	reqBodyBytes, _ := json.Marshal(reqBody)

	r, err := http.NewRequest("POST", "/", bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	feeds := map[string]*feeds.Feed{}

	logger := zap.NewNop()

	feedChan := make(chan *newsletter.NewsLetter)

	db, err := database.New(logger, ":memory:")
	require.NoError(t, err)
	s := &Server{
		feeds:    feeds,
		logger:   logger,
		feedChan: feedChan,
		db:       &db,
	}

	s.CreateFeed(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	for _, v := range s.feeds {
		require.Equal(t, v.Title, feedName)
	}
}

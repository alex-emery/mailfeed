package rss

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alex-emery/mailfeed/newsletter"
	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGetFeed(t *testing.T) {
	w := httptest.NewRecorder()

	r, err := http.NewRequest("GET", "/feed", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	feed := &feeds.Feed{
		Title: "Test Feed",
	}

	logger := zap.NewNop()

	feedChan := make(chan *newsletter.NewsLetter)

	s := &Server{
		feed:     feed,
		logger:   logger,
		feedChan: feedChan,
	}

	s.AddToFeed(&newsletter.NewsLetter{
		Subject: "Test Subject",
		Body:    "Test Body",
	})

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

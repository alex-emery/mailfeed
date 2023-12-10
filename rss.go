package main

import (
	"time"

	"github.com/gorilla/feeds"
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

func appendToFood(feed *feeds.Feed, letter *NewsLetter) {
	now := time.Now()

	feed.Items = append(feed.Items, &feeds.Item{
		Title:       letter.Subject,
		Description: letter.Body,
		Created:     now,
	})
}

package newsletter

import "time"

// NewsLetter struct, is a common interface between email and rss.
type NewsLetter struct {
	Inbox   string
	Date    time.Time
	Subject string
	Body    string
}

func New(inbox, subject, body string, date time.Time) *NewsLetter {
	return &NewsLetter{
		Inbox:   inbox,
		Date:    date,
		Subject: subject,
		Body:    body,
	}
}

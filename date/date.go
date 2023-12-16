package date

import "time"

// ParseDate parses a date string into a time.Time object.
// we get two different formats from the IMAP server, so we need to try both.
func ParseDate(date string) (time.Time, error) {
	parsedTime, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", date)
	if err == nil {
		return parsedTime, nil
	}

	return time.Parse("Mon, 2 Jan 2006 15:04:05 -0700 (UTC)", date)
}

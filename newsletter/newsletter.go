package newsletter

type NewsLetter struct {
	Subject string
	Body    string
}

func New(subject, body string) *NewsLetter {
	return &NewsLetter{
		Subject: subject,
		Body:    body,
	}
}

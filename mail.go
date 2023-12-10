package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message"
	"github.com/emersion/go-message/textproto"
	"go.uber.org/zap"
)

type NewsLetter struct {
	Subject string
	Body    string
}

func NewNewsLetter(subject, body string) *NewsLetter {
	return &NewsLetter{
		Subject: subject,
		Body:    body,
	}
}

type Mail struct {
	c          *imapclient.Client
	SeqNum     uint32
	logger     *zap.Logger
	letterChan chan<- *NewsLetter
	fetchReady chan struct{}
	cleanups   []func() error
	emailID    string
}

func (m *Mail) startIdle(logger *zap.Logger, server, username, password string, fetchReady chan<- struct{}) error {
	options := imapclient.Options{
		UnilateralDataHandler: &imapclient.UnilateralDataHandler{
			Expunge: func(seqNum uint32) {
				logger.Info("message %v has been expunged", zap.Uint32("seqNum", seqNum))
			},
			Mailbox: func(data *imapclient.UnilateralDataMailbox) {
				if data.NumMessages != nil {
					logger.Info("a new message has been received")
					fetchReady <- struct{}{}
				}
			},
		},
	}

	c, err := newMailClient(server, username, password, &options)
	if err != nil {
		return fmt.Errorf("failed to create mail client: %v", err)
	}

	m.cleanups = append(m.cleanups, c.Close)
	_, err = c.Select("INBOX", nil).Wait()
	if err != nil {
		return fmt.Errorf("failed to select INBOX: %v", err)
	}

	logger.Debug("Starting idle")
	idle, err := c.Idle()
	if err != nil {
		return fmt.Errorf("failed to start IDLE: %v", err)
	}

	m.cleanups = append(m.cleanups, idle.Close)
	return nil
}

func newMailClient(server, username, password string, options *imapclient.Options) (*imapclient.Client, error) {
	c, err := imapclient.DialTLS(server, options)
	if err != nil {
		return nil, fmt.Errorf("failed to dial IMAP server: %v", err)
	}

	if err := c.Login(username, password).Wait(); err != nil {
		return nil, fmt.Errorf("failed to login: %v", err)
	}

	return c, nil
}

func NewMail(logger *zap.Logger, server, username, password, emailID string, feed chan<- *NewsLetter) (*Mail, error) {
	c, err := newMailClient(server, username, password, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create mail client: %v", err)
	}

	selectCmd, err := c.Select("INBOX", nil).Wait()
	if err != nil {
		return nil, fmt.Errorf("failed to select INBOX: %v", err)
	}

	logger.Debug("selected INBOX", zap.Uint32("UIDNext", selectCmd.UIDNext))

	fetchReady := make(chan struct{})

	mail := &Mail{
		c:          c,
		SeqNum:     selectCmd.UIDNext,
		logger:     logger,
		letterChan: feed,
		fetchReady: fetchReady,
		cleanups:   []func() error{c.Close},
		emailID:    emailID,
	}

	err = mail.startIdle(logger, server, username, password, fetchReady)
	if err != nil {
		return nil, fmt.Errorf("failed to start IDLE: %v", err)
	}

	return mail, nil
}

func (m *Mail) Close() {
	for _, cleanup := range m.cleanups {
		if err := cleanup(); err != nil {
			m.logger.Error("failed to cleanup", zap.Error(err))
		}

	}
}

func (m *Mail) StartFetch() {
	m.logger.Info("starting fetch loop")
	for {
		<-m.fetchReady
		m.Fetch()
	}
}

// Fetches a single email from the server.
func (m *Mail) Fetch() {
	seqSet := imap.SeqSetNum(m.SeqNum)
	fetchOptions := &imap.FetchOptions{
		UID:      true,
		Flags:    true,
		Envelope: true,
		BodySection: []*imap.FetchItemBodySection{
			// {Specifier: imap.PartSpecifierMIME},
			{Specifier: imap.PartSpecifierHeader},
			{Specifier: imap.PartSpecifierText},
		},
	}

	getMessages := func() []*imapclient.FetchMessageBuffer {
		messages, err := m.c.UIDFetch(seqSet, fetchOptions).Collect()
		if err != nil {
			m.logger.Error("failed to fetch messages", zap.Error(err))
			return nil
		}

		return messages
	}

	var messages []*imapclient.FetchMessageBuffer
	retry := 0
	for messages = getMessages(); messages == nil && retry < 3; messages = getMessages() {
		time.Sleep(1 * time.Second)
		retry += 1
	}

	for _, msg := range messages {
		m.logger.Info("message received", zap.Uint32("UID", msg.UID), zap.String("subject", msg.Envelope.Subject))
		var header message.Header
		var body string

		for k, buf := range msg.BodySection {
			if k.Specifier == imap.PartSpecifierHeader {
				reader := bufio.NewReader(bytes.NewReader(buf))

				txtHeader, err := textproto.ReadHeader(reader)
				if err != nil {
					m.logger.Error("failed to parse header", zap.Error(err))
					continue
				}

				header = message.Header{
					Header: txtHeader,
				}
			}
			if k.Specifier == imap.PartSpecifierText {
				body = string(buf)
			}
		}

		parsedMessage, err := message.New(header, strings.NewReader(body))
		if err != nil {
			m.logger.Error("failed to parse message", zap.Error(err))
			continue
		}

		if parsedMessage.Header.Get("To") != m.emailID {
			fmt.Println("not for me, skipping")
			continue
		}

		contents, err := ConvertEmail(*parsedMessage)
		if err != nil {
			m.logger.Error("failed to convert email", zap.Error(err))
			continue
		}

		m.letterChan <- NewNewsLetter(msg.Envelope.Subject, contents)

		m.SeqNum = msg.UID + 1
	}
}

func ConvertEmail(message message.Entity) (string, error) {
	// Retrieve the Content-Type header and parse it to get the boundary value
	mediaType, params, err := mime.ParseMediaType(message.Header.Get("Content-Type"))
	if err != nil {
		return "", fmt.Errorf("error parsing MIME header: %v", err)
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		parts, err := ParsePart(message.Body, params["boundary"])
		if err != nil {
			return "", fmt.Errorf("error parsing MIME parts: %v", err)
		}

		if strings.HasPrefix(mediaType, "multipart/alternative") { // try html first, then plain text
			for _, part := range parts {
				if part.ContentType == "text/html" {
					return part.Content, nil
				}
			}
			for _, part := range parts {
				if part.ContentType == "plain/text" {
					return part.Content, nil
				}
			}

			return "", fmt.Errorf("no text/html or text/plain part found")
		}

		body := ""
		for _, part := range parts {
			body += "\n" + part.Content
		}

		return body, nil
	}

	contents, err := io.ReadAll(message.Body)
	return string(contents), err
}

type Part struct {
	ContentType string
	Content     string
}

func ParsePart(mimeData io.Reader, boundary string) ([]Part, error) {
	reader := multipart.NewReader(mimeData, boundary)
	if reader == nil {
		return nil, fmt.Errorf("error creating MIME multipart reader")
	}

	var parts []Part
	for {
		new_part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, err
		}

		mediaType, _, err := mime.ParseMediaType(new_part.Header.Get("Content-Type"))
		if err != nil {
			return nil, err
		}

		if strings.HasPrefix(mediaType, "multipart/") {
			nestedParts, err := ParsePart(new_part, boundary)
			if err != nil {
				return nil, err
			}

			parts = append(parts, nestedParts...)

		} else {
			partString, err := PartToString(new_part)
			if err != nil {
				return nil, err
			}

			parts = append(parts, Part{ContentType: mediaType, Content: partString})
		}

	}

	return parts, nil
}

func PartToString(part *multipart.Part) (string, error) {
	part_data, err := io.ReadAll(part)
	if err != nil {
		return "", fmt.Errorf("error reading MIME part - %v", err)
	}

	content_transfer_encoding := strings.ToUpper(part.Header.Get("Content-Transfer-Encoding"))

	switch {

	case strings.Compare(content_transfer_encoding, "BASE64") == 0:
		decoded_content, err := base64.StdEncoding.DecodeString(string(part_data))
		if err != nil {
			return "", fmt.Errorf("error decoding base64 - %v", err)
		}

		return string(decoded_content), nil

	case strings.Compare(content_transfer_encoding, "QUOTED-PRINTABLE") == 0:
		decoded_content, err := io.ReadAll(quotedprintable.NewReader(bytes.NewReader(part_data)))
		if err != nil {
			return "", fmt.Errorf("error decoding quoted-printable - %v", err)
		}

		return string(decoded_content), nil

	default:
		return string(part_data), nil
	}
}

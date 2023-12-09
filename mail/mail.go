package mail

import (
	"fmt"
	"log"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"go.uber.org/zap"
)

type Mail struct {
	c      *imapclient.Client
	SeqNum uint32
	logger *zap.Logger
}

func New(logger *zap.Logger, server, username, password string) (*Mail, error) {
	c, err := imapclient.DialTLS(server, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to dial IMAP server: %v", err)
	}

	// defer c.Close()

	if err := c.Login(username, password).Wait(); err != nil {
		log.Fatalf("failed to login: %v", err)
	}

	selectCmd, err := c.Select("INBOX", nil).Wait()
	if err != nil {
		log.Fatalf("failed to select INBOX: %v", err)
	}

	logger.Debug("selected INBOX", zap.Uint32("UIDNext", selectCmd.UIDNext))
	return &Mail{
		c:      c,
		SeqNum: selectCmd.UIDNext,
		logger: logger,
	}, nil
}

func (f *Mail) Close() {
	f.c.Close()
}

func (f *Mail) Fetch() {
	seqSet := imap.SeqSetNum(f.SeqNum)
	fetchOptions := &imap.FetchOptions{
		UID:      true,
		Flags:    true,
		Envelope: true,
		BodySection: []*imap.FetchItemBodySection{
			// {Specifier: imap.PartSpecifierHeader},
			{Specifier: imap.PartSpecifierText},
		},
	}

	getMessages := func() []*imapclient.FetchMessageBuffer {
		messages, err := f.c.UIDFetch(seqSet, fetchOptions).Collect()
		if err != nil {
			fmt.Println("failed to fetch messages", err)
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
		fmt.Printf("message %d: %s\n", msg.UID, msg.Envelope.Subject)

		for k, buf := range msg.BodySection {
			switch k.Specifier {
			case imap.PartSpecifierText:
				body := string(buf)
				fmt.Println(body)
				break
			}
		}

		f.SeqNum = msg.UID + 1
	}
}

type NewsLetter struct {
	Subject string
	Body    string
}

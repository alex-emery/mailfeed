package mail

import (
	"bufio"
	"bytes"
	"os"
	"testing"

	"github.com/emersion/go-message"
	"github.com/emersion/go-message/textproto"
	"github.com/stretchr/testify/require"
)

func TestConvertEmailHandlesAlternative(t *testing.T) {
	buf, err := os.ReadFile("testdata/header.txt")
	require.NoError(t, err)

	reader := bufio.NewReader(bytes.NewReader(buf))

	txtHeader, err := textproto.ReadHeader(reader)
	require.NoError(t, err)

	header := message.Header{
		Header: txtHeader,
	}

	buf, err = os.ReadFile("testdata/body.txt")
	require.NoError(t, err)

	parsedMessage, err := message.New(header, bytes.NewReader(buf))
	require.NoError(t, err)

	email, err := ConvertEmail(*parsedMessage)
	require.NoError(t, err)

	require.Equal(t, "<html><head><meta charset=\"utf-8\"><title>The News Letter</title>\r\n</head><body>Wow this is like the text/html version.</body></html>", email)

}

package request

import (
	"io"
	// "strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n
	if n > cr.numBytesPerRead {
		n = cr.numBytesPerRead
		cr.pos -= n - cr.numBytesPerRead
	}
	return n, nil
}

func TestRequestFromReader(t *testing.T) {
	// Test: Good GET Request line
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Good GET Request line with path
	reader = &chunkReader{
		data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 1,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// // Test: Good GET Request line
	// r, err := RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	// require.NoError(t, err)
	// require.NotNil(t, r)
	// assert.Equal(t, "GET", r.RequestLine.Method)
	// assert.Equal(t, "/", r.RequestLine.RequestTarget)
	// assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	//
	// // Test: Good GET Request line with path
	// r, err = RequestFromReader(strings.NewReader("GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	// require.NoError(t, err)
	// require.NotNil(t, r)
	// assert.Equal(t, "GET", r.RequestLine.Method)
	// assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	// assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	//
	// r, err = RequestFromReader(strings.NewReader("POST /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	// require.NoError(t, err)
	// require.NotNil(t, r)
	// assert.Equal(t, "POST", r.RequestLine.Method)
	// assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	// assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	//
	// // Test: Invalid order
	// _, err = RequestFromReader(strings.NewReader("/coffee POST HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	// require.Error(t, err)
	//
	// // Test: Invalid number of parts in request line
	// _, err = RequestFromReader(strings.NewReader("/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	// require.Error(t, err)
}

func TestHeaderFromReader(t *testing.T) {
	// Test: Standard Headers
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069", r.Headers["host"])
	assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
	assert.Equal(t, "*/*", r.Headers["accept"])

	// Test: Malformed Header
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)
}

//!go test ./internal/request -v

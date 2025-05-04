package request

import (
	"io"
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
	end := cr.pos + cr.numBytesPerRead
	if end > len(cr.data) {
		end = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:end])
	cr.pos += n
	return n, nil
}

func TestHeaderFromReader(t *testing.T) {
	// Valid GET request with headers
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	req, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, req)

	assert.Equal(t, "GET", req.RequestLine.Method)
	assert.Equal(t, "/", req.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", req.RequestLine.HttpVersion)

	assert.Equal(t, "localhost:42069", req.Headers["host"])
	assert.Equal(t, "curl/7.81.0", req.Headers["user-agent"])
	assert.Equal(t, "*/*", req.Headers["accept"])

	//  Malformed header (missing colon)
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
		numBytesPerRead: 3,
	}
	req, err = RequestFromReader(reader)
	require.Error(t, err)
	require.Nil(t, req)

	//  Valid POST request with multiple reads
	reader = &chunkReader{
		data:            "POST /submit HTTP/1.1\r\nContent-Type: text/plain\r\nContent-Length: 0\r\n\r\n",
		numBytesPerRead: 5,
	}
	req, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, req)

	assert.Equal(t, "POST", req.RequestLine.Method)
	assert.Equal(t, "/submit", req.RequestLine.RequestTarget)
	assert.Equal(t, "text/plain", req.Headers["content-type"])
	assert.Equal(t, "0", req.Headers["content-length"])
}

func TestBodyFromReader(t *testing.T) {
	// Test: Standard Body
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 13\r\n" +
			"\r\n" +
			"hello world!\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "hello world!\n", string(r.Body))

	// Test: Body shorter than reported content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 20\r\n" +
			"\r\n" +
			"partial content",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)
}

//!go test ./internal/request -v

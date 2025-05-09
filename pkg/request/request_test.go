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

func TestBodyFromReader(t *testing.T) {
	t.Run("Standard Body", func(t *testing.T) {
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
	})

	t.Run("Empty Body, 0 reported content length", func(t *testing.T) {
		reader := &chunkReader{
			data: "POST /empty HTTP/1.1\r\n" +
				"Host: localhost\r\n" +
				"Content-Length: 0\r\n" +
				"\r\n",
			numBytesPerRead: 4,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "", string(r.Body))
	})

	t.Run("Empty Body, no reported content length", func(t *testing.T) {
		reader := &chunkReader{
			data: "POST /empty HTTP/1.1\r\n" +
				"Host: localhost\r\n" +
				"\r\n",
			numBytesPerRead: 4,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "", string(r.Body))
	})

	t.Run("Body shorter than reported content length", func(t *testing.T) {
		reader := &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 20\r\n" +
				"\r\n" +
				"partialcontent",
			numBytesPerRead: 3,
		}
		r, err := RequestFromReader(reader)
		require.Error(t, err)
		assert.Nil(t, r)
	})

	t.Run("No Content-Length but Body Exists", func(t *testing.T) {
		reader := &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				" Host: localhost\r\n" +
				"\r\n" +
				"extra body here",
			numBytesPerRead: 4,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "extra body here", string(r.Body))
	})
}

//!go test ./internal/request -v

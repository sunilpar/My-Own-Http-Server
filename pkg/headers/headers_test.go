package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidSingleHeader(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)

	require.NoError(t, err)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)
}

func TestValidSingleHeaderWithExtraWhitespace(t *testing.T) {
	headers := NewHeaders()
	data := []byte("    Host:     localhost:42069   \r\n")
	n, done, err := headers.Parse(data)

	require.NoError(t, err)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 34, n)
	assert.False(t, done)
}

func TestValidTwoHeadersWithExistingHeaders(t *testing.T) {
	headers := NewHeaders()

	data1 := []byte("Host: localhost\r\n")
	n1, done1, err1 := headers.Parse(data1)
	require.NoError(t, err1)
	assert.False(t, done1)
	assert.Equal(t, "localhost", headers["host"])
	assert.Equal(t, 17, n1)

	data2 := []byte("User-Agent: curl/7.64.1\r\n")
	n2, done2, err2 := headers.Parse(data2)
	require.NoError(t, err2)
	assert.False(t, done2)
	assert.Equal(t, "curl/7.64.1", headers["user-agent"])
	assert.Equal(t, 25, n2)
}

func TestValidDone(t *testing.T) {
	headers := NewHeaders()
	data := []byte("\r\n")
	n, done, err := headers.Parse(data)

	require.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.True(t, done)
}

func TestInvalidSpacingHeader(t *testing.T) {
	headers := NewHeaders()
	data := []byte("   Host : localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)

	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func TestInvalidCharacterInKey(t *testing.T) {
	headers := NewHeaders()
	data := []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid character in header key")
	assert.Equal(t, 0, n)
	assert.False(t, done)
}
func TestMultipleHeaderValuesAreCombined(t *testing.T) {
	headers := NewHeaders()
	// First header
	data1 := []byte("Set-Person: lane-loves-go\r\n")
	n1, done1, err1 := headers.Parse(data1)
	require.NoError(t, err1)
	require.False(t, done1)
	assert.Equal(t, "lane-loves-go", headers["set-person"])
	assert.Equal(t, 27, n1)
	// Second header with same key
	data2 := []byte("Set-Person: prime-loves-zig\r\n")
	n2, done2, err2 := headers.Parse(data2)
	require.NoError(t, err2)
	require.False(t, done2)
	assert.Equal(t, "lane-loves-go, prime-loves-zig", headers["set-person"])
	assert.Equal(t, 29, n2)
	// Third header with same key
	data3 := []byte("Set-Person: tj-loves-ocaml\r\n")
	n3, done3, err3 := headers.Parse(data3)
	require.NoError(t, err3)
	require.False(t, done3)
	assert.Equal(t, "lane-loves-go, prime-loves-zig, tj-loves-ocaml", headers["set-person"])
	assert.Equal(t, 28, n3)
}

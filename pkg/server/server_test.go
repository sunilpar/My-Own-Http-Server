package server_test

import (
	"bufio"
	"fmt"
	"httpfromtcp/internal/server"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func sendRawRequest(t *testing.T, address string, raw string) (string, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	_, err = fmt.Fprintln(conn, raw)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	resp, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return "", err
	}

	return resp, nil
}

func TestValidRequestReturns200(t *testing.T) {
	s, err := server.Serve(42069)
	assert.NoError(t, err)
	defer s.Close()

	time.Sleep(100 * time.Millisecond) // Give the server time to start

	raw := "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"
	resp, err := sendRawRequest(t, "localhost:42069", raw)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(resp, "200 OK"))
}

func TestMalformedHeaderRequest(t *testing.T) {
	s, err := server.Serve(42069)
	assert.NoError(t, err)
	defer s.Close()

	time.Sleep(100 * time.Millisecond) // Give the server time to start

	// Send malformed header (missing colon)
	raw := "GET / HTTP/1.1\r\nInvalidHeader\r\n\r\n"
	_, err = sendRawRequest(t, "localhost:42069", raw)

	// Expect error or empty response (because server logs and closes)
	assert.Error(t, err)
}

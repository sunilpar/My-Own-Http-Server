package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var reason string
	switch statusCode {
	case StatusOK:
		reason = "OK"
	case StatusBadRequest:
		reason = "Bad Request"
	case StatusInternalServerError:
		reason = "Internal Server Error"
	default:
		reason = ""
	}
	_, err := fmt.Fprintf(w, "HTTP/1.1 %d %s\r\n", statusCode, reason)
	wf := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reason)
	_, err = w.Write([]byte(wf))
	fmt.Printf("HTTP/1.1 %d %s\r\n", statusCode, reason)
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

func WriteHeaders(w io.Writer, h headers.Headers) error {
	for key, value := range h {
		_, err := fmt.Fprintf(w, "%s: %s\r\n", key, value)
		fmt.Printf("%v: %v\n", key, value)
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(w, "\r\n")
	return err
}

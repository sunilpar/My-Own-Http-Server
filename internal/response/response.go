package response

import (
	"errors"
	"fmt"
	"httpfromtcp/internal/headers"
	"net"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func statusText(code StatusCode) string {
	switch code {
	case StatusOK:
		return "OK"
	case StatusBadRequest:
		return "Bad Request"
	case StatusInternalServerError:
		return "Internal Server Error"
	default:
		return ""
	}
}

type writerState int

const (
	stateInitial writerState = iota
	stateStatusWritten
	stateHeadersWritten
	stateBodyWritten
)

type Writer struct {
	conn   net.Conn
	state  writerState
	Header headers.Headers
}

func NewWriter(conn net.Conn) *Writer {
	return &Writer{
		conn:   conn,
		state:  stateInitial,
		Header: headers.NewHeaders(),
	}
}

func (w *Writer) WriteStatusLine(code StatusCode) error {
	if w.state != stateInitial {
		return errors.New("status line already written")
	}
	_, err := fmt.Fprintf(w.conn, "HTTP/1.1 %d %s\r\n", code, statusText(code))
	if err == nil {
		w.state = stateStatusWritten
	}
	return err
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	if w.state != stateStatusWritten {
		return errors.New("must write status line before headers")
	}
	for key, val := range h {
		_, err := fmt.Fprintf(w.conn, "%s: %s\r\n", key, val)
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(w.conn, "\r\n")
	if err == nil {
		w.state = stateHeadersWritten
	}
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != stateHeadersWritten {
		return 0, errors.New("must write headers before body")
	}
	w.state = stateBodyWritten
	return w.conn.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.state != stateHeadersWritten && w.state != stateBodyWritten {
		return 0, errors.New("must write headers before chunked body")
	}
	w.state = stateBodyWritten

	chunkSize := len(p)
	_, err := fmt.Fprintf(w.conn, "%x\r\n", chunkSize)
	if err != nil {
		return 0, err
	}
	n, err := w.conn.Write(p)
	if err != nil {
		return n, err
	}
	_, err = fmt.Fprint(w.conn, "\r\n")
	if err != nil {
		return n, err
	}
	return n, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.state != stateBodyWritten {
		return 0, errors.New("no chunked body started")
	}
	return fmt.Fprint(w.conn, "0\r\n\r\n")
}

package request

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strings"
	"unicode"
)

type parserState int

const (
	stateInitialized parserState = iota
	stateParsingHeaders
	stateDone
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	state       parserState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, 8)
	readTo := 0
	req := &Request{
		state:   stateInitialized,
		Headers: headers.NewHeaders(),
	}
	for {
		if readTo == len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}
		n, err := reader.Read(buf[readTo:])
		if err != nil && err != io.EOF {
			return nil, err
		}
		readTo += n

		parsed, perr := req.parse(buf[:readTo])
		if perr != nil {
			return nil, perr
		}
		if parsed == 0 && err == io.EOF {
			return nil, fmt.Errorf("incomplete request")
		}
		if parsed > 0 {
			copy(buf, buf[parsed:readTo])
			readTo -= parsed
		}
		if req.state == stateDone {
			return req, nil
		}
	}
}

func parseRequestLine(data []byte) (RequestLine, int, error) {
	s := string(data)
	i := strings.Index(s, "\r\n")
	if i == -1 {
		return RequestLine{}, 0, nil
	}
	line := s[:i]
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return RequestLine{}, 0, fmt.Errorf("request line must have exactly 3 parts")
	}
	method := parts[0]
	target := parts[1]
	version := parts[2]

	for _, ch := range method {
		if !unicode.IsUpper(ch) {
			return RequestLine{}, 0, fmt.Errorf("method must be all uppercase letters")
		}
	}
	if method != "GET" && method != "POST" {
		return RequestLine{}, 0, fmt.Errorf("method must be GET or POST")
	}
	if version != "HTTP/1.1" {
		return RequestLine{}, 0, fmt.Errorf("unsupported HTTP version: %s", version)
	}
	return RequestLine{
		Method:        method,
		RequestTarget: target,
		HttpVersion:   "1.1",
	}, i + 2, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalParsed := 0
	for r.state != stateDone {
		n, err := r.parseSingle(data[totalParsed:])
		if err != nil {
			return totalParsed, err
		}
		if n == 0 {
			break
		}
		totalParsed += n
	}
	return totalParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case stateInitialized:
		rl, n, err := parseRequestLine(data)
		if err != nil || n == 0 {
			return n, err
		}
		r.RequestLine = rl
		r.state = stateParsingHeaders
		return n, nil

	case stateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.state = stateDone
		}
		return n, nil

	default:
		return 0, fmt.Errorf("unknown state")
	}
}

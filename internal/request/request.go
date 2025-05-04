// package request
//
// import (
// 	"fmt"
// 	"httpfromtcp/internal/headers"
// 	"io"
// 	"strconv"
// 	"strings"
// 	"unicode"
// )
//
// type parserState int
//
// const (
// 	stateInitialized parserState = iota
// 	stateParsingHeaders
// 	stateParsingBody
// 	stateDone
// )
//
// type Request struct {
// 	RequestLine RequestLine
// 	Headers     headers.Headers
// 	Body        []byte
// 	state       parserState
// 	bodyLength  int
// }
//
// type RequestLine struct {
// 	HttpVersion   string
// 	RequestTarget string
// 	Method        string
// }
//
// func RequestFromReader(reader io.Reader) (*Request, error) {
// 	buf := make([]byte, 8)
// 	readTo := 0
// 	req := &Request{
// 		state:   stateInitialized,
// 		Headers: headers.NewHeaders(),
// 	}
//
// 	reachedEOF := false
//
// 	for {
// 		if readTo == len(buf) {
// 			newBuf := make([]byte, len(buf)*2)
// 			copy(newBuf, buf)
// 			buf = newBuf
// 		}
//
// 		n, err := reader.Read(buf[readTo:])
// 		if err != nil && err != io.EOF {
// 			return nil, err
// 		}
// 		if err == io.EOF {
// 			reachedEOF = true
// 		}
//
// 		readTo += n
// 		parsed, perr := req.parse(buf[:readTo], reachedEOF)
// 		if perr != nil {
// 			return nil, perr
// 		}
// 		if parsed == 0 && reachedEOF {
// 			// If headers and body are done, return
// 			if req.state == stateParsingBody {
// 				// If there's no content-length, and we reach EOF, we're done
// 				req.state = stateDone
// 				return req, nil
// 			}
// 			if req.state != stateDone {
// 				return nil, fmt.Errorf("incomplete request")
// 			}
// 		}
//
// 		if parsed > 0 {
// 			copy(buf, buf[parsed:readTo])
// 			readTo -= parsed
// 		}
//
// 		if req.state == stateDone {
// 			return req, nil
// 		}
// 	}
// }
//
// func parseRequestLine(data []byte) (RequestLine, int, error) {
// 	s := string(data)
// 	i := strings.Index(s, "\r\n")
// 	if i == -1 {
// 		return RequestLine{}, 0, nil
// 	}
// 	line := s[:i]
// 	parts := strings.Split(line, " ")
// 	if len(parts) != 3 {
// 		return RequestLine{}, 0, fmt.Errorf("request line must have exactly 3 parts")
// 	}
// 	method := parts[0]
// 	target := parts[1]
// 	version := parts[2]
//
// 	for _, ch := range method {
// 		if !unicode.IsUpper(ch) {
// 			return RequestLine{}, 0, fmt.Errorf("method must be all uppercase letters")
// 		}
// 	}
// 	if method != "GET" && method != "POST" {
// 		return RequestLine{}, 0, fmt.Errorf("method must be GET or POST")
// 	}
// 	if version != "HTTP/1.1" {
// 		return RequestLine{}, 0, fmt.Errorf("unsupported HTTP version: %s", version)
// 	}
// 	return RequestLine{
// 		Method:        method,
// 		RequestTarget: target,
// 		HttpVersion:   "1.1",
// 	}, i + 2, nil
// }
//
// func (r *Request) parse(data []byte, eof bool) (int, error) {
// 	totalParsed := 0
// 	for r.state != stateDone {
// 		n, err := r.parseSingle(data[totalParsed:], eof)
// 		if err != nil {
// 			return totalParsed, err
// 		}
// 		if n == 0 {
// 			break
// 		}
// 		totalParsed += n
// 	}
// 	return totalParsed, nil
// }
//
// func (r *Request) parseSingle(data []byte, eof bool) (int, error) {
// 	switch r.state {
// 	case stateInitialized:
// 		rl, n, err := parseRequestLine(data)
// 		if err != nil || n == 0 {
// 			return n, err
// 		}
// 		r.RequestLine = rl
// 		r.state = stateParsingHeaders
// 		return n, nil
//
// 	case stateParsingHeaders:
// 		n, done, err := r.Headers.Parse(data)
// 		if err != nil {
// 			return 0, err
// 		}
// 		if done {
// 			contentLenStr := r.Headers.Get("Content-Length")
// 			if contentLenStr != "" {
// 				length, err := strconv.Atoi(contentLenStr)
// 				if err != nil || length < 0 {
// 					return 0, fmt.Errorf("invalid Content-Length")
// 				}
// 				r.bodyLength = length
// 				if length == 0 {
// 					r.state = stateDone
// 				} else {
// 					r.state = stateParsingBody
// 				}
// 			} else {
// 				// No content length = allow any body up to EOF
// 				r.state = stateParsingBody
// 			}
// 		}
// 		return n, nil
//
// 	case stateParsingBody:
// 		r.Body = append(r.Body, data...)
// 		if r.bodyLength > 0 {
// 			fmt.Printf("r.body:%v r.bodylen:%v\n", len(r.Body), r.bodyLength)
// 			if len(r.Body) > r.bodyLength {
// 				return 0, fmt.Errorf("body larger than Content-Length")
// 			}
// 			if len(r.Body) == r.bodyLength {
// 				r.state = stateDone
// 			}
// 			return len(data), nil
// 		}
// 		// No Content-Length: we only know we're done if EOF reached
// 		if eof {
// 			r.state = stateDone
// 		}
// 		return len(data), nil
//
// 	default:
// 		return 0, fmt.Errorf("unknown state")
// 	}
// }

package request

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type parserState int

const (
	stateInitialized parserState = iota
	stateParsingHeaders
	stateParsingBody
	stateDone
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	state       parserState
	bodyLength  int
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

	reachedEOF := false

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
		if err == io.EOF {
			reachedEOF = true
		}

		readTo += n
		parsed, perr := req.parse(buf[:readTo], reachedEOF)
		if perr != nil {
			return nil, perr
		}
		if parsed > 0 {
			copy(buf, buf[parsed:readTo])
			readTo -= parsed
		}

		if reachedEOF {
			if req.state == stateParsingBody {
				if req.bodyLength > 0 && len(req.Body) < req.bodyLength {
					return nil, fmt.Errorf("incomplete body: expected %d bytes, got %d", req.bodyLength, len(req.Body))
				}
				// If no Content-Length, treat whatever is received as valid body
				req.state = stateDone
			} else if req.state != stateDone {
				return nil, fmt.Errorf("incomplete request")
			}
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

func (r *Request) parse(data []byte, eof bool) (int, error) {
	totalParsed := 0
	for r.state != stateDone {
		n, err := r.parseSingle(data[totalParsed:], eof)
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

func (r *Request) parseSingle(data []byte, eof bool) (int, error) {
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
			contentLenStr := r.Headers.Get("Content-Length")
			if contentLenStr != "" {
				length, err := strconv.Atoi(contentLenStr)
				if err != nil || length < 0 {
					return 0, fmt.Errorf("invalid Content-Length")
				}
				r.bodyLength = length
				if length == 0 {
					r.state = stateDone
				} else {
					r.state = stateParsingBody
				}
			} else {
				r.state = stateParsingBody
			}
		}
		return n, nil

	case stateParsingBody:
		r.Body = append(r.Body, data...)
		if r.bodyLength > 0 {
			if len(r.Body) > r.bodyLength {
				return 0, fmt.Errorf("body larger than Content-Length")
			}
			if len(r.Body) == r.bodyLength {
				r.state = stateDone
			}
		} else if eof {
			// No content-length and EOF reached
			r.state = stateDone
		}
		return len(data), nil

	default:
		return 0, fmt.Errorf("unknown state")
	}
}

package server

import (
	"bufio"
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"log"
	"net"
	"sync/atomic"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", port, err)
	}
	s := &Server{
		listener: ln,
	}
	go s.listen()
	return s, nil
}

func (s *Server) Close() error {
	if s.closed.CompareAndSwap(false, true) {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("Error accepting connection: %v\n", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	_, err := request.RequestFromReader(bufio.NewReader(conn))
	if err != nil {
		log.Printf("Malformed request: %v\n", err)
		return
	}
	// fmt.Printf("\n--request line--\n")
	// fmt.Printf("method: %v\n", r.RequestLine.Method)
	// fmt.Printf("requestTarget: %v\n", r.RequestLine.RequestTarget)
	// fmt.Printf("httpversion: %v\n", r.RequestLine.HttpVersion)
	// fmt.Printf("\n--headers--\n")
	// for i, v := range r.Headers {
	// 	fmt.Printf("[%v]: %v\n", i, v)
	// }
	// fmt.Printf("\n--Body--\n")
	// fmt.Printf("r.body: %v\n\n\n", string(r.Body))

	const body = "sunil World!"
	contentLen := len(body) + 2

	if err := response.WriteStatusLine(conn, response.StatusOK); err != nil {
		log.Printf("Failed to write status line: %v\n", err)
		return
	}

	h := response.GetDefaultHeaders(contentLen)
	if err := response.WriteHeaders(conn, h); err != nil {
		log.Printf("Failed to write headers: %v\n", err)
		return
	}

	if _, err := conn.Write([]byte("\r\n")); err != nil {
		log.Printf("Failed to write CRLF after headers: %v\n", err)
		return
	}
	fmt.Printf("%v\n", body)

	if _, err := conn.Write([]byte(body)); err != nil {
		log.Printf("Failed to write body: %v\n", err)
	}
}

package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/sunilpar/My-Own-Http-Server/internal/request"
	"github.com/sunilpar/My-Own-Http-Server/internal/response"
)

type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
	handler  Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", port, err)
	}
	s := &Server{
		listener: ln,
		handler:  handler,
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

	req, err := request.RequestFromReader(bufio.NewReader(conn))
	if err != nil {
		log.Printf("Malformed request: %v\n", err)
		w := response.NewWriter(conn)
		w.WriteStatusLine(response.StatusBadRequest)
		w.Header.Set("Content-Type", "text/plain")
		body := []byte("Bad Request")
		w.Header.Set("Content-Length", fmt.Sprintf("%d", len(body)+2))
		w.WriteHeaders(w.Header)
		w.WriteBody(body)
		return
	}

	respWriter := response.NewWriter(conn)
	s.handler(respWriter, req)
}

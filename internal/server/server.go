// package server
//
// import (
// 	"bufio"
// 	"bytes"
// 	"fmt"
// 	"httpfromtcp/internal/request"
// 	"httpfromtcp/internal/response"
// 	"io"
// 	"log"
// 	"net"
// 	"sync/atomic"
// )
//
// type Handler func(w io.Writer, req *request.Request) *HandlerError
//
// type HandlerError struct {
// 	StatusCode int
// 	Message    string
// }
//
// type Server struct {
// 	listener net.Listener
// 	closed   atomic.Bool
// 	handler  Handler
// }
//
// func Serve(port int, handler Handler) (*Server, error) {
// 	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to listen on port %d: %w", port, err)
// 	}
// 	s := &Server{
// 		listener: ln,
// 		handler:  handler,
// 	}
// 	go s.listen()
// 	return s, nil
// }
//
// func (s *Server) Close() error {
// 	if s.closed.CompareAndSwap(false, true) {
// 		return s.listener.Close()
// 	}
// 	return nil
// }
//
// func (s *Server) listen() {
// 	for {
// 		conn, err := s.listener.Accept()
// 		if err != nil {
// 			if s.closed.Load() {
// 				return
// 			}
// 			log.Printf("Error accepting connection: %v\n", err)
// 			continue
// 		}
// 		go s.handle(conn)
// 	}
// }
//
// func (s *Server) handle(conn net.Conn) {
// 	defer conn.Close()
// 	req, err := request.RequestFromReader(bufio.NewReader(conn))
// 	if err != nil {
// 		log.Printf("Malformed request: %v\n", err)
// 		herr := &HandlerError{
// 			StatusCode: int(response.StatusBadRequest),
// 			Message:    err.Error(),
// 		}
// 		writeHandlerError(conn, herr)
// 		return
// 	}
// 	var buf bytes.Buffer
// 	handlerErr := s.handler(&buf, req)
// 	if handlerErr != nil {
// 		writeHandlerError(conn, handlerErr)
// 		return
// 	}
//
// 	body := buf.Bytes()
// 	contentLen := len(body)
//
// 	if err := response.WriteStatusLine(conn, response.StatusOK); err != nil {
// 		log.Printf("Failed to write status line: %v\n", err)
// 		return
// 	}
// 	headers := response.GetDefaultHeaders(contentLen)
// 	if err := response.WriteHeaders(conn, headers); err != nil {
// 		log.Printf("Failed to write headers: %v\n", err)
// 		return
// 	}
// 	if _, err := conn.Write([]byte("\r\n")); err != nil {
// 		log.Printf("Failed to write CRLF: %v\n", err)
// 		return
// 	}
// 	if _, err := conn.Write(body); err != nil {
// 		log.Printf("Failed to write body: %v\n", err)
// 	}
// }
//
// func writeHandlerError(w io.Writer, herr *HandlerError) {
// 	if err := response.WriteStatusLine(w, response.StatusCode(herr.StatusCode)); err != nil {
// 		log.Printf("Failed to write error status line: %v\n", err)
// 		return
// 	}
// 	headers := response.GetDefaultHeaders(len(herr.Message))
// 	if err := response.WriteHeaders(w, headers); err != nil {
// 		log.Printf("Failed to write error headers: %v\n", err)
// 		return
// 	}
// 	w.Write([]byte("\r\n"))
// 	w.Write([]byte(herr.Message))
// }

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
		w.Header.Set("Content-Type", "text/html")
		body := []byte(`
			<html>
			  <head><title>400 Bad Request</title></head>
			  <body>
				<h1>Bad Request</h1>
				<p>Your request honestly kinda sucked.</p>
			  </body>
			</html>`)
		w.Header.Set("Content-Length", fmt.Sprintf("%d", len(body)+2))
		w.WriteHeaders(w.Header)
		w.WriteBody(body)
		return
	}

	respWriter := response.NewWriter(conn)
	s.handler(respWriter, req)
}

// package main
//
// import (
// 	"httpfromtcp/internal/server"
// 	"log"
// 	"os"
// 	"os/signal"
// 	"syscall"
// )
//
// const port = 42069
//
// func main() {
// 	server, err := server.Serve(port)
// 	if err != nil {
// 		log.Fatalf("Error starting server: %v", err)
// 	}
// 	defer server.Close()
// 	log.Println("Server started on port", port)
//
// 	sigChan := make(chan os.Signal, 1)
// 	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
// 	<-sigChan
// 	log.Println("\n--Server gracefully stopped--")
// }

package main

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

func myHandler(w io.Writer, req *request.Request) *server.HandlerError {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		return &server.HandlerError{
			StatusCode: 400,
			Message:    "Your problem is not my problem\n",
		}
	case "/myproblem":
		return &server.HandlerError{
			StatusCode: 500,
			Message:    "Woopsie, my bad\n",
		}
	default:
		fr := "All good, frfr\n"
		w.Write([]byte(fr))
		return nil
	}
}

func main() {
	srv, err := server.Serve(port, myHandler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer srv.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("\n--Server gracefully stopped--")
}

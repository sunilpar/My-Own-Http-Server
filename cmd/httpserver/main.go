package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

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

func myHandler(w *response.Writer, req *request.Request) {
	var status response.StatusCode
	var body string

	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		status = response.StatusBadRequest
		body = `
<html>
  <head><title>400 Bad Request</title></head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`
	case "/myproblem":
		status = response.StatusInternalServerError
		body = `
<html>
  <head><title>500 Internal Server Error</title></head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`
	default:
		status = response.StatusOK
		body = `
<html>
  <head><title>200 OK</title></head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`
	}

	w.WriteStatusLine(status)
	w.Header.Set("Content-Type", "text/html")
	w.Header.Set("Content-Length", fmt.Sprintf("%d", len(body)))
	w.WriteHeaders(w.Header)
	w.WriteBody([]byte(body))
}

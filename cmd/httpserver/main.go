package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
	path := req.RequestLine.RequestTarget

	if strings.HasPrefix(path, "/httpbin/") {
		proxyToHttpbin(w, req, path)
		return
	}

	if path == "/video" {
		serveVideo(w)
		return
	}

	var status response.StatusCode
	var body string

	switch path {
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

func proxyToHttpbin(w *response.Writer, req *request.Request, path string) {
	targetPath := strings.TrimPrefix(path, "/httpbin")
	url := "https://httpbin.org" + targetPath

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error proxying request: %v", err)
		w.WriteStatusLine(response.StatusInternalServerError)
		w.Header.Set("Content-Type", "text/plain")
		w.Header.Set("Content-Length", "21")
		w.WriteHeaders(w.Header)
		w.WriteBody([]byte("Proxying failed, sorry."))
		return
	}
	defer resp.Body.Close()

	w.WriteStatusLine(response.StatusOK)

	w.Header.Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header.Del("Content-Length")
	w.Header.Set("Transfer-Encoding", "chunked")
	w.Header.Set("Trailer", "X-Content-SHA256, X-Content-Length")
	w.WriteHeaders(w.Header)

	var fullBody []byte
	buf := make([]byte, 1024)

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			fullBody = append(fullBody, chunk...)

			if _, writeErr := w.WriteChunkedBody(chunk); writeErr != nil {
				log.Println("Error writing chunk:", writeErr)
				break
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("Error reading response:", err)
			break
		}
	}

	w.WriteChunkedBodyDone()

	sum := sha256.Sum256(fullBody)
	hash := hex.EncodeToString(sum[:])
	length := fmt.Sprintf("%d", len(fullBody))

	trailer := headers.NewHeaders()
	trailer.Set("X-Content-SHA256", hash)
	trailer.Set("X-Content-Length", length)
	if err := w.WriteTrailers(trailer); err != nil {
		log.Println("Error writing trailers:", err)
	}
}

func serveVideo(w *response.Writer) {
	data, err := os.ReadFile("assets/vim.mp4")
	if err != nil {
		log.Printf("Error reading video file: %v", err)
		w.WriteStatusLine(response.StatusInternalServerError)
		w.Header.Set("Content-Type", "text/plain")
		body := "Video not found"
		w.Header.Set("Content-Length", fmt.Sprintf("%d", len(body)))
		w.WriteHeaders(w.Header)
		w.WriteBody([]byte(body))
		return
	}

	w.WriteStatusLine(response.StatusOK)
	w.Header.Set("Content-Type", "video/mp4")
	w.Header.Set("Content-Length", fmt.Sprintf("%d", len(data)))
	w.WriteHeaders(w.Header)
	w.WriteBody(data)
}

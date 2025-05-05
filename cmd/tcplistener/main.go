package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"log"
	"net"
)

const port = ":42069"

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("error listening for TCP traffic: %s\n", err.Error())
	}
	defer listener.Close()

	fmt.Println("Listening for TCP traffic on", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("error: %s\n", err.Error())
		}
		fmt.Println("Accepted connection from", conn.RemoteAddr())

		rn, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatalf("error: %s\n", err.Error())
		}
		fmt.Printf("\nrequest line:\n")
		fmt.Printf("method:%v\n", rn.RequestLine.Method)
		fmt.Printf("requestTarget:%v\n", rn.RequestLine.RequestTarget)
		fmt.Printf("httpversion:%v\n", rn.RequestLine.HttpVersion)
		fmt.Printf("\nheaders:\n")
		for i, v := range rn.Headers {
			fmt.Printf("[%v]:%v\n", i, v)
		}
		fmt.Printf("\nBody:\n")
		fmt.Printf("r.body:%v\n", string(rn.Body))

		fmt.Println("\nConnection to ", conn.RemoteAddr(), "closed")
	}
}

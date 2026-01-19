package main

import (
	"fmt"
	"net"

	"http-scratch/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		panic(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		fmt.Printf("connection established\n")

		req, err := request.RequestFromReader(conn)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HTTPVersion)
		fmt.Printf("Headers:\n")
		for k, v := range req.Headers {
			fmt.Printf("- %s: %s\n", k, v)
		}
		fmt.Printf("Body:\n")
		fmt.Printf("%s\n", string(req.Body))
	}
}

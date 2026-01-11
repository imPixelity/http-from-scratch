package main

import (
	"errors"
	"fmt"
	"io"
	"net"
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

		for line := range getLinesChannel(conn) {
			fmt.Printf("read: %s\n", line)
		}

		fmt.Printf("connection closed\n")
	}
}

func getLinesChannel(conn net.Conn) <-chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)
		defer conn.Close()
		str := ""
		buffer := make([]byte, 8)
		for {
			n, err := conn.Read(buffer)
			isNewline := false

			for k, v := range buffer[:n] {
				if v == '\n' {
					str += string(buffer[:k])
					ch <- str
					isNewline = true
					str = ""
					str += string(buffer[k+1 : n])
					break
				}
			}

			if !isNewline {
				str += string(buffer[:n])
			}

			if err != nil {
				if str != "" {
					ch <- str
				}
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Printf("%v", err)
			}
		}
	}()

	return ch
}

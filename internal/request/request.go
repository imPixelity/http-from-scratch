package request

import (
	"fmt"
	"io"
	"log"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req, err := io.ReadAll(reader)
	if err != nil {
		log.Fatal("read fail")
	}

	requestLine, err := parseRequestLine(req)
	if err != nil {
		log.Fatal("parse fail")
	}

	fmt.Printf("method: %s\n", requestLine.Method)
	fmt.Printf("destination: %s\n", requestLine.RequestTarget)
	fmt.Printf("version: %s\n", requestLine.HTTPVersion)

	return nil, nil
}

func parseRequestLine(req []byte) (*RequestLine, error) {
	requestLine := new(RequestLine)
	for line := range strings.SplitSeq(string(req), "\r\n") {
		data := strings.Split(line, " ")

		if data[0] != strings.ToUpper(data[0]) {
			return nil, fmt.Errorf("method is not capital")
		}

		if !strings.Contains(data[2], "HTTP/") {
			return nil, fmt.Errorf("version not started with HTTP")
		}

		requestLine.Method = data[0]
		requestLine.RequestTarget = data[1]
		requestLine.HTTPVersion = data[2]

		break
	}

	return requestLine, nil
}

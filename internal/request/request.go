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

		if len(data) != 3 {
			return nil, fmt.Errorf("bad request")
		}

		if data[0] != strings.ToUpper(data[0]) {
			return nil, fmt.Errorf("method is not capital")
		}

		if data[2] != "HTTP/1.1" {
			return nil, fmt.Errorf("version not supported")
		}

		requestLine.Method = data[0]
		requestLine.RequestTarget = data[1]
		requestLine.HTTPVersion = strings.TrimPrefix(data[2], "HTTP/")

		break
	}

	return requestLine, nil
}

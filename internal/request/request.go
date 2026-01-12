package request

import (
	"fmt"
	"io"
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
		return nil, fmt.Errorf("read fail")
	}

	request := &Request{}
	err = parseRequestLine(request, req)
	if err != nil {
		return nil, fmt.Errorf("parse fail")
	}

	return request, nil
}

func parseRequestLine(request *Request, req []byte) error {
	var requestLine RequestLine
	for line := range strings.SplitSeq(string(req), "\r\n") {
		data := strings.Split(line, " ")

		if len(data) != 3 {
			return fmt.Errorf("bad request")
		}

		if data[0] != strings.ToUpper(data[0]) {
			return fmt.Errorf("method is not capital")
		}

		if data[2] != "HTTP/1.1" {
			return fmt.Errorf("version not supported")
		}

		requestLine.Method = data[0]
		requestLine.RequestTarget = data[1]
		requestLine.HTTPVersion = strings.TrimPrefix(data[2], "HTTP/")

		request.RequestLine = requestLine
		break
	}

	return nil
}

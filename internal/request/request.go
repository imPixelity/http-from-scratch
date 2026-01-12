package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
}

var (
	ErrMalformedReqLine       = errors.New("malformed request line")
	ErrUnsupportedHTTPVersion = errors.New("unsupported HTTP version")
	Separator                 = "\r\n"
)

func parseRequestLine(str string) (RequestLine, string, error) {
	idx := strings.Index(str, Separator)

	if idx == -1 {
		return RequestLine{}, "", ErrMalformedReqLine
	}

	startLine := str[:idx]
	restOfMsg := str[idx+len(Separator):]

	parts := strings.Split(startLine, " ")
	if len(parts) != 3 {
		return RequestLine{}, "", ErrMalformedReqLine
	}

	if parts[0] != strings.ToUpper(parts[0]) {
		return RequestLine{}, "", ErrMalformedReqLine
	}

	httpParts := strings.Split(parts[2], "/")
	if len(httpParts) != 2 || httpParts[0] != "HTTP" || httpParts[1] != "1.1" {
		return RequestLine{}, "", ErrMalformedReqLine
	}

	rl := RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HTTPVersion:   httpParts[1],
	}

	return rl, restOfMsg, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("unable to io.ReadAll: %w", err)
	}

	rl, _, err := parseRequestLine(string(data))
	if err != nil {
		return nil, fmt.Errorf("unable to parse request line: %w", err)
	}

	return &Request{
		RequestLine: rl,
	}, nil
}

package request

import (
	"bytes"
	"errors"
	"io"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

var (
	ErrReadFile         = errors.New("fail to read from reader")
	ErrMalformedReqLine = errors.New("malformed request line")
)

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf, err := io.ReadAll(reader)
	if err != nil {
		return nil, ErrReadFile
	}

	reqLine, err := parseRequestLine(buf)
	if err != nil {
		return nil, err
	}

	return &Request{
		RequestLine: reqLine,
	}, nil
}

func parseRequestLine(b []byte) (RequestLine, error) {
	if idx := bytes.Index(b, []byte("\r\n")); idx != -1 {
		reqLine := b[:idx]
		parts := bytes.Split(reqLine, []byte(" "))

		if len(parts) != 3 {
			return RequestLine{}, ErrMalformedReqLine
		}

		if ok := validateFormat(parts[0], parts[2]); !ok {
			return RequestLine{}, ErrMalformedReqLine
		}

		return RequestLine{
			Method:        string(parts[0]),
			RequestTarget: string(parts[1]),
			HTTPVersion:   string(bytes.TrimPrefix(parts[2], []byte("HTTP/"))),
		}, nil
	}
	return RequestLine{}, ErrMalformedReqLine
}

func validateFormat(method []byte, version []byte) bool {
	// Method only uppercase
	if !bytes.Equal(method, bytes.ToUpper(method)) {
		return false
	}

	// Version only accepting 1.1
	if !bytes.Equal(version, []byte("HTTP/1.1")) {
		return false
	}

	return true
}

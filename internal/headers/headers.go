package headers

import (
	"bytes"
	"errors"
	"strings"
)

var (
	ErrMalformedHeaderLine = errors.New("malformed header line")
	ErrMalformedHeaderKey  = errors.New("malformed header key")
	CRLF                   = []byte("\r\n")
)

type Headers map[string]string

func (h Headers) Get(key string) string {
	return h[strings.ToLower(key)]
}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)
	if v, ok := h[key]; ok {
		h[key] = v + "," + value
		return
	}
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	isDone := false

	for {
		idx := bytes.Index(data[read:], Separator)
		if idx == -1 {
			return read, isDone, nil
		}

		// Empty Header
		if idx == 0 {
			isDone = true
			read += idx + len(Separator)
			break
		}

		key, value, err := parseHeader(data[read : read+idx])
		if err != nil {
			return 0, isDone, err
		}

		if !isToken([]byte(key)) {
			return 0, isDone, ErrMalformedHeaderName
		}

		read += idx + len(Separator)
		h.Set(key, value)
	}

	return read, isDone, nil
}

func NewHeaders() Headers {
	return Headers{}
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	for {
		idx := bytes.Index(data[read:], CRLF)
		if idx == -1 {
			break
		}

		// Empty Header
		if idx == 0 {
			read += len(CRLF)
			return read, true, nil
		}

		key, value, err := parseHeader(data[read : read+idx])
		if err != nil {
			return read, false, err
		}

		if !validateToken([]byte(key)) {
			return read, false, ErrMalformedHeaderKey
		}

		read += idx + len(CRLF)
		h.Set(key, value)
	}
	return read, false, nil
}

func parseHeader(data []byte) (string, string, error) {
	if idx := bytes.Index(data, []byte(":")); idx != -1 {
		// Key not empty
		if len(data[:idx]) == 0 {
			return "", "", ErrMalformedHeaderLine
		}

		// No WS before colon
		if data[idx-1] == ' ' {
			return "", "", ErrMalformedHeaderLine
		}
	}

	parts := bytes.SplitN(data, []byte(":"), 2)
	for i := range parts {
		parts[i] = bytes.TrimSpace(parts[i])
	}

	return string(parts[0]), string(parts[1]), nil
}

func validateToken(key []byte) bool {
	for _, ch := range key {
		found := false
		if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			found = true
		}

		switch ch {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			found = true
		}

		if !found {
			return false
		}
	}
	return true
}

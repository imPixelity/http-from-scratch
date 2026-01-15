package headers

import (
	"bytes"
	"errors"
	"strings"
)

var (
	ErrMalformedLine       = errors.New("malformed header line")
	ErrMalformedField      = errors.New("malformed field")
	ErrMalformedHeaderName = errors.New("malformed header name")
	Separator              = []byte("\r\n")
)

type Headers map[string]string

func (h Headers) Get(name string) string {
	return h[strings.ToLower(name)]
}

func (h Headers) Set(name, value string) {
	h[strings.ToLower(name)] = value
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	isDone := false

	for {
		idx := bytes.Index(data[read:], Separator)
		if idx == -1 {
			return 0, isDone, ErrMalformedLine
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

func parseHeader(fieldLine []byte) (string, string, error) {
	if idx := bytes.Index(fieldLine, []byte(":")); idx != -1 {
		if fieldLine[idx-1] == ' ' {
			return "", "", ErrMalformedLine
		}
	}

	parts := bytes.SplitN(fieldLine, []byte(":"), 2)

	if len(parts) != 2 {
		return "", "", ErrMalformedField
	}

	for i := range parts {
		parts[i] = bytes.Trim(parts[i], " ")
	}

	return string(parts[0]), string(parts[1]), nil
}

func isToken(b []byte) bool {
	for _, c := range b {
		found := false
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			found = true
		}

		switch c {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			found = true
		}

		if !found {
			return false
		}
	}
	return true
}

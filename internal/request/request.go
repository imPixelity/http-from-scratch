package request

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"

	"http-scratch/internal/headers"
)

var (
	ErrMalformedReqLine       = errors.New("malformed request line")
	ErrUnsupportedHTTPVersion = errors.New("unsupported HTTP version")
	ErrReqInErrorState        = errors.New("request in error state")
	Separator                 = []byte("\r\n")
)

type parserState string

const (
	StateInit    parserState = "init"
	StateHeaders parserState = "headers"
	StateBody    parserState = "body"
	StateDone    parserState = "done"
	StateError   parserState = "error"
)

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        string
	state       parserState
}

func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

func (r *Request) hasBody() bool {
	cl := getInt(r.Headers, "content-length", 0)
	return cl > 0
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

outer:
	for {
		currentData := data[read:]

		switch r.state {
		case StateError:
			return 0, ErrReqInErrorState

		case StateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}

			if n == 0 {
				break outer
			}

			r.RequestLine = rl
			read += n
			r.state = StateHeaders

		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}

			if n == 0 {
				break outer
			}

			read += n

			if done {
				if r.hasBody() {
					r.state = StateBody
				} else {
					r.state = StateDone
				}
			}

		case StateBody:
			cl := getInt(r.Headers, "content-length", 0)

			if cl == 0 {
				panic("yes")
			}

			remaining := min(cl-len(r.Body), len(currentData))
			r.Body += string(currentData[:remaining])
			read += remaining

			if len(r.Body) == cl {
				r.state = StateDone
			}

		case StateDone:
			break outer

		default:
			panic("yo")
		}
	}
	return read, nil
}

func newRequest() *Request {
	return &Request{state: StateInit, Headers: headers.NewHeaders(), Body: ""}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	// Could overrun
	buf := make([]byte, 1024)
	bufLen := 0
	for !request.done() {
		// Read from buffer
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			return nil, err
		}

		// Parse read buffer
		bufLen += n
		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}

		// Move to beginning
		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}

	return request, nil
}

func parseRequestLine(b []byte) (RequestLine, int, error) {
	idx := bytes.Index(b, Separator)

	// Return, not enough data
	if idx == -1 {
		return RequestLine{}, 0, nil
	}

	startLine := b[:idx]
	read := idx + len(Separator)

	// Split start line by 3
	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return RequestLine{}, 0, ErrMalformedReqLine
	}

	// Method only uppercase
	if string(parts[0]) != strings.ToUpper(string(parts[0])) {
		return RequestLine{}, 0, ErrMalformedReqLine
	}

	// Take version from HTTP/ and only accepting 1.1
	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return RequestLine{}, 0, ErrMalformedReqLine
	}

	rl := RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HTTPVersion:   string(httpParts[1]),
	}

	return rl, read, nil
}

func getInt(h headers.Headers, name string, defaultVal int) int {
	valStr, ok := h.Get(name)
	if !ok {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

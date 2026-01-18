package request

import (
	"bytes"
	"errors"
	"io"
	"strconv"

	"http-scratch/internal/headers"
)

var (
	ErrConvertType      = errors.New("fail to convert data type")
	ErrReadFile         = errors.New("fail to read from reader")
	ErrParseHeaders     = errors.New("fail to parse headers")
	ErrParseBody        = errors.New("fail to parse body")
	ErrMalformedReqLine = errors.New("malformed request line")
	Separator           = []byte("\r\n")
)

type StateStatus int

const (
	StateInit    parserState = "init"
	StateHeaders parserState = "headers"
	StateDone    parserState = "done"
	StateError   parserState = "error"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	state       StateStatus
}

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	state       parserState
}

func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
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
				return 0, err
			}

			if n == 0 {
				break outer
			}

			read += n

			if done {
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
	return &Request{state: StateInit, Headers: headers.NewHeaders()}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := &Request{Headers: headers.NewHeaders(), state: StateInit}
	buf := make([]byte, 1024)
	bufPos := 0

	for req.state != StateDone {
		consumed := 0
		n, err := reader.Read(buf[bufPos:])

		// Handle EOF separately, still process buffer first
		isEOF := false
		if err != nil {
			if errors.Is(err, io.EOF) {
				isEOF = true
			} else {
				return nil, ErrReadFile
			}
		}

		bufPos += n

		switch req.state {
		case StateInit:
			consumed, err = req.parse(buf[:bufPos])
			if err != nil {
				return nil, err
			}
		case StateParseHeaders:
			done := false
			consumed, done, err = req.Headers.Parse(buf[:bufPos])
			if err != nil {
				return nil, err
			}
			if done {
				req.state = StateParseBody
			}
		case StateParseBody:
			clStr := req.Headers.Get("content-length")
			if clStr == "" {
				break
			}

			cl, err := strconv.Atoi(clStr)
			if err != nil {
				return nil, ErrConvertType
			}

			req.Body = append(req.Body, buf[:bufPos]...)
			consumed = len(buf[:bufPos])

			if len(req.Body) > cl {
				return nil, ErrParseBody
			}
			if len(req.Body) == cl {
				req.state = StateDone
			}
		}

		// Move unconsumed bytes to the front
		copy(buf, buf[consumed:bufPos])
		bufPos -= consumed

		// Now check EOF after processing remaining buffer
		if isEOF {
			if req.state == StateParseHeaders {
				return nil, ErrParseHeaders
			}

			if req.state == StateParseBody {
				cl := req.Headers.Get("content-length")
				if cl != "" {
					// Body was expected, but incomplete
					return nil, ErrParseBody
				}
				// No Content-Length, no body expected
				req.state = StateDone
			}
		}
	}
	return req, nil
}

func parseRequestLine(b []byte) (RequestLine, int, error) {
	if idx := bytes.Index(b, Separator); idx != -1 {
		reqLine := b[:idx]
		read := idx + len(Separator)
		parts := bytes.Split(reqLine, []byte(" "))

		if len(parts) != 3 {
			return RequestLine{}, 0, ErrMalformedReqLine
		}

		if ok := validateFormat(parts[0], parts[2]); !ok {
			return RequestLine{}, 0, ErrMalformedReqLine
		}

		return RequestLine{
			Method:        string(parts[0]),
			RequestTarget: string(parts[1]),
			HTTPVersion:   string(bytes.TrimPrefix(parts[2], []byte("HTTP/"))),
		}, read, nil
	}
	return RequestLine{}, 0, nil
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

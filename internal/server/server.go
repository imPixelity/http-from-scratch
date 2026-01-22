package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync/atomic"

	"http-scratch/internal/request"
	"http-scratch/internal/response"
)

var (
	ErrListenServer = errors.New("fail to listen")
	ErrCloseServer  = errors.New("fail to close")
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

type Server struct {
	listener  net.Listener
	handler   Handler
	isRunning atomic.Bool
}

func (s *Server) Close() error {
	if !s.isRunning.CompareAndSwap(true, false) {
		return ErrCloseServer
	}
	s.listener.Close()
	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	if !s.isRunning.Load() {
		conn.Close()
		return
	}

	headers := response.GetDefaultHeaders(0)
	request, err := request.RequestFromReader(conn)
	if err != nil {
		response.WriteStatusLine(conn, response.StatusBadRequest)
		response.WriteHeaders(conn, headers)
		return
	}

	writer := bytes.NewBuffer([]byte{})
	handlerError := s.handler(writer, request)

	var body []byte
	status := response.StatusOK
	if handlerError != nil {
		status = handlerError.StatusCode
		body = []byte(handlerError.Message)
	} else {
		body = writer.Bytes()
	}

	headers.Replace("Content-Length", strconv.Itoa(len(body)))
	response.WriteStatusLine(conn, status)
	response.WriteHeaders(conn, headers)
	conn.Write(body)
}

func Serve(port int, handler Handler) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, ErrListenServer
	}

	server := &Server{
		listener: listener,
		handler:  handler,
	}
	server.isRunning.Store(true)

	go server.listen()
	return server, nil
}

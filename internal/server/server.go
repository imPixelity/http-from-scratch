package server

import (
	"errors"
	"fmt"
	"log"
	"net"
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

type Handler func(w *response.Writer, req *request.Request)

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
			if !s.isRunning.Load() {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	if !s.isRunning.Load() {
		conn.Close()
		return
	}

	responseWriter := response.NewWriter(conn)
	request, err := request.RequestFromReader(conn)
	if err != nil {
		responseWriter.WriteStatusLine(response.StatusBadRequest)
		responseWriter.WriteHeaders(response.GetDefaultHeaders(0))
		return
	}

	s.handler(responseWriter, request)
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

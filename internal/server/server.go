package server

import (
	"errors"
	"fmt"
	"net"
	"sync/atomic"

	"http-scratch/internal/response"
)

var (
	ErrListenServer = errors.New("fail to listen")
	ErrCloseServer  = errors.New("fail to close")
)

type Server struct {
	listener  net.Listener
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
	response.WriteStatusLine(conn, response.StatusOK)
	response.WriteHeaders(conn, headers)
}

func Serve(port int) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, ErrListenServer
	}

	server := &Server{
		listener: listener,
	}
	server.isRunning.Store(true)

	go server.listen()
	return server, nil
}

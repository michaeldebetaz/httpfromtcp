package server

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync/atomic"

	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
)

type Handler func(w io.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (handlerErr *HandlerError) Write(w io.Writer) error {
	if err := response.WriteStatusLine(w, handlerErr.StatusCode); err != nil {
		return fmt.Errorf("error writing status line: %w", err)
	}

	contentLen := len(handlerErr.Message)
	if err := response.WriteHeaders(w, response.GetDefaultHeaders(contentLen)); err != nil {
		return fmt.Errorf("error writing headers: %w", err)
	}

	if _, err := w.Write([]byte(handlerErr.Message)); err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}

	return nil
}

type Server struct {
	isOpen   *atomic.Bool
	listener net.Listener
	handler  Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("error creating listener: %w", err)
	}

	server := &Server{isOpen: &atomic.Bool{}, listener: listener, handler: handler}

	server.open()

	go server.listen()

	return server, nil
}

func (s *Server) open() {
	s.isOpen.Store(true)
}

func (s *Server) Close() error {
	if err := s.listener.Close(); err != nil {
		return fmt.Errorf("error closing the listener: %w", err)
	}

	s.isOpen.Store(false)

	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if !s.isOpen.Load() {
				return
			}
			slog.Error("Error accepting connection", "error", err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		slog.Error("Error parsing request", "err", err)
		return
	}

	var body bytes.Buffer

	handlerErr := s.handler(&body, req)
	if handlerErr != nil {
		if err := handlerErr.Write(conn); err != nil {
			slog.Error("Error writing handler error", "err", err)
			return
		}
		return
	}

	if err := response.WriteStatusLine(conn, response.OK); err != nil {
		slog.Error("Error writing status line", "err", err)
		return
	}

	if err := response.WriteHeaders(conn, response.GetDefaultHeaders(body.Len())); err != nil {
		slog.Error("Error writing headers", "err", err)
		return
	}

	if _, err := conn.Write(body.Bytes()); err != nil {
		slog.Error("Error writing response from buffer", "err", err)
		return
	}
}

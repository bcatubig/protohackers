package tcp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	Addr    string
	Handler Handler
	BaseCtx context.Context

	ln         net.Listener
	inShutdown atomic.Bool
	mu         sync.Mutex
	activeConn map[*Conn]struct{}
	logger     *slog.Logger
}

func (s *Server) WithLogger(l *slog.Logger) {
	s.logger = l
}

func (s *Server) ListenAndServe() error {
	if s.logger == nil {
		s.logger = slog.New(slog.DiscardHandler)
	}

	s.activeConn = make(map[*Conn]struct{})

	addr := s.Addr

	if addr == "" {
		addr = "0.0.0.0:8000"
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		if s.isShutdown() {
			return ErrServerClosed
		}
		return err
	}
	s.logger.Info(fmt.Sprintf("Listening on: %s", addr))

	ocl := &onceCloseListener{Listener: ln}
	s.ln = ocl
	defer ocl.Close()

	return s.Serve(ocl)
}

func (s *Server) Serve(ln net.Listener) error {
	for {
		c, err := ln.Accept()
		if err != nil {
			if s.isShutdown() {
				return ErrServerClosed
			}

			var oe *net.OpError
			if errors.As(err, &oe) {
				if oe.Temporary() {
					time.Sleep(5 * time.Millisecond)
					continue
				}
			}

			return err
		}

		conn := s.newConn(c)
		s.addConn(conn)
		go conn.serve()
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("server shutting down")
	s.inShutdown.Store(true)

	s.ln.Close()

	for c := range s.activeConn {
		c.close()
		s.removeConn(c)
	}

	return nil
}

func (s *Server) isShutdown() bool {
	return s.inShutdown.Load()
}

func (s *Server) newConn(rwc net.Conn) *Conn {
	return &Conn{rwc: rwc, server: s}
}

func (s *Server) addConn(c *Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.activeConn[c] = struct{}{}
}

func (s *Server) removeConn(c *Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.activeConn, c)
}

type onceCloseListener struct {
	net.Listener
	once sync.Once
	err  error
}

func (c *onceCloseListener) Close() error {
	c.once.Do(c.close)
	return c.err
}

func (c *onceCloseListener) close() {
	c.err = c.Listener.Close()
}

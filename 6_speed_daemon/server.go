package main

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
)

var ErrServerClosed = errors.New("tcp: Server closed")

type Server struct {
	ln         net.Listener
	activeConn map[*conn]struct{}
	inShutdown atomic.Bool
	mu         sync.Mutex
}

type onceCloseListener struct {
	net.Listener
	once       sync.Once
	closeError error
}

func (oc *onceCloseListener) Close() error {
	oc.once.Do(oc.close)
	return oc.closeError
}

func (oc *onceCloseListener) close() {
	oc.closeError = oc.Listener.Close()
}

func NewServer(addr string) (*Server, error) {
	s := &Server{
		activeConn: make(map[*conn]struct{}),
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	s.ln = &onceCloseListener{Listener: ln}

	return s, nil
}

func (s *Server) Serve() error {
	defer s.ln.Close()

	for {
		rw, err := s.ln.Accept()
		if err != nil {
			if s.shuttingDown() {
				return ErrServerClosed
			}
			logger.Error(err.Error())
			continue
		}
		c := s.newConn(rw)
		s.addConn(c)
		go c.serve()
	}
}

func (s *Server) Shutdown() error {
	s.inShutdown.Store(true)

	s.mu.Lock()
	defer s.mu.Unlock()

	for c := range s.activeConn {
		c.rwc.Close()
		delete(s.activeConn, c)
	}

	return nil
}

func (s *Server) shuttingDown() bool {
	return s.inShutdown.Load()
}

func (s *Server) addConn(c *conn) {
	s.mu.Lock()
	s.activeConn[c] = struct{}{}
	s.mu.Unlock()
}

func (s *Server) newConn(rwc net.Conn) *conn {
	c := &conn{
		server: s,
		rwc:    rwc,
	}

	return c
}

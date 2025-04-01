package main

import (
	"errors"
	"math"
	"net"
	"sync"
	"sync/atomic"
)

var ErrServerClosed = errors.New("tcp: Server closed")

type Server struct {
	ln          net.Listener
	inShutdown  atomic.Bool
	mu          sync.Mutex
	activeConn  map[*conn]struct{}
	dispatchers []*Dispatcher
}

func NewServer(addr string) (*Server, error) {
	s := &Server{
		activeConn: make(map[*conn]struct{}),
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	s.ln = ln

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

func speed(distance int, time int) int {
	return distance / (time / 3600)
}

func currentDay(timestamp uint32) int {
	return int(math.Floor(float64(timestamp) / 86400))
}

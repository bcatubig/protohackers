package main

import (
	"errors"
	"math"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var ErrServerClosed = errors.New("tcp: Server closed")

type Server struct {
	Addr       string
	inShutdown atomic.Bool
	mu         sync.Mutex
	activeConn map[*conn]struct{}
}

func (s *Server) ListenAndServe() error {
	if s.shuttingDown() {
		return ErrServerClosed
	}

	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	return s.Serve(ln)
}

func (s *Server) Serve(l net.Listener) error {
	defer l.Close()

	for {
		rw, err := l.Accept()
		if err != nil {
			var ne net.Error
			if s.shuttingDown() {
				return ErrServerClosed
			}

			if errors.As(err, &ne) {
				logger.Error("accept error", "error", err.Error())
				time.Sleep(5 * time.Millisecond)
				continue
			}

			return err
		}

		c := s.newConn(rw)
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

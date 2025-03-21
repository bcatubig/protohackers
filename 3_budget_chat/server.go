package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
)

type Server struct {
	l          net.Listener
	inShutdown atomic.Bool
	mu         sync.Mutex
	activeConn map[*conn]struct{}
}

func NewServer(addr string) (*Server, error) {
	l, err := net.Listen("tcp", addr)

	if err != nil {
		return nil, err
	}

	return &Server{
		l:          l,
		activeConn: make(map[*conn]struct{}),
		mu:         sync.Mutex{},
	}, nil
}

func (s *Server) ListenAndServe() error {
	for {
		c, err := s.l.Accept()

		if err != nil {
			logger.Error("connection error", "error", err.Error())
			continue
		}

		nConn := &conn{ip: c.RemoteAddr().String(), rwc: c}

		s.addConn(nConn)
		go s.handle(nConn)
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.inShutdown.Store(true)

	logger.Info("shutting down server")

	chanDone := make(chan struct{})

	go func() {
		for c := range s.activeConn {
			c.close()
			s.removeConn(c)
		}
		chanDone <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		logger.Error("timed out waiting for server to shutdown")
	case <-chanDone:
		logger.Info("shutdown successful")
	}

	return nil
}

func (s *Server) handle(c *conn) {
	defer func() {
		logger.Info("closing connection", "ip", c.ip)
		s.removeConn(c)
		c.rwc.Close()
	}()

	if c.username == "" {
		io.Copy(c.rwc, strings.NewReader("Welcome to budgetchat! What shall I call you?\n"))

		// read in username
		buf := bufio.NewReaderSize(c.rwc, 64)
		username, err := buf.ReadString('\n')

		if err != nil {
			logger.Error("error reading username", "error", err.Error(), "ip", c.ip)
			return
		}

		if len(username) < 1 {
			logger.Error("Username less than 1 character", "ip", c.ip)
			io.Copy(c.rwc, strings.NewReader("error: username must be at least 1 character"))
			return
		}

		c.username = strings.TrimSuffix(username, "\n")

		for sC := range s.activeConn {
			if sC == c {
				continue
			}
			if c.username == sC.username {
				logger.Error("duplicate username", "username", c.username, "ip", c.ip)
				io.Copy(c.rwc, strings.NewReader(fmt.Sprintf("error: the username %s is already taken", c.username)))
				return
			}
		}

		io.Copy(c.rwc, strings.NewReader(fmt.Sprintf("\nWelcome, %s!\n", c.username)))
	}

	// handle the thing
	for {
		if s.inShutdown.Load() {
			return
		}

		buf := bufio.NewReaderSize(c.rwc, 2048)

		line, err := buf.ReadString('\n')
		line = strings.TrimSuffix(line, "\n")

		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			logger.Error(err.Error())
			continue
		}

		// Send this line to all clients except current client
		logger.Info("read line", "line", line)
	}
}

func (s *Server) addConn(c *conn) {
	s.mu.Lock()
	s.activeConn[c] = struct{}{}
	s.mu.Unlock()
}

func (s *Server) removeConn(c *conn) {
	s.mu.Lock()
	delete(s.activeConn, c)
	s.mu.Unlock()
}

type conn struct {
	rwc      net.Conn
	ip       string
	username string
}

func (c conn) String() string {
	return fmt.Sprintf("%s - %s", c.ip, c.username)
}

func (c *conn) close() {
	logger.Info("closing connection", "ip", c.ip)
	io.Copy(c.rwc, strings.NewReader("info: server closed connection"))
	c.rwc.Close()
}

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

func (s *Server) Broadcast(c *conn, msg string) {
	logger.Info("broadcasting message", "msg", msg, "from", c.username)
	wg := &sync.WaitGroup{}
	wg.Add(len(s.activeConn) - 1)

	s.mu.Lock()
	for cn := range s.activeConn {
		if cn == c {
			continue
		}
		go func() {
			defer wg.Done()
			io.Copy(cn.rwc, strings.NewReader(msg))
		}()
	}
	s.mu.Unlock()

	wg.Wait()
}

func (s *Server) SendMessage(c *conn, msg string) {
	msg = fmt.Sprintf("[%s] %s\n", c.username, msg)
	s.Broadcast(c, msg)
}

func (s *Server) UserAction(c *conn, action string) {
	for cn := range s.activeConn {
		if cn == c {
			continue
		}

		var msg string

		switch action {
		case "join":
			msg = fmt.Sprintf("* %s has entered the room\n", c.username)
		case "leave":
			msg = fmt.Sprintf("* %s has left the room\n", c.username)
		}

		s.Broadcast(c, msg)
	}
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
		logger.Info("read username", "username", c.username, "ip", c.ip)

		if err != nil {
			logger.Error("error reading username", "error", err.Error(), "ip", c.ip)
			return
		}

		c.username = strings.TrimSuffix(username, "\n")

		if len(c.username) < 1 {
			logger.Error("Username less than 1 character", "ip", c.ip)
			io.Copy(c.rwc, strings.NewReader("error: username must be at least 1 character\n"))
			return
		}

		switch c.username {
		case "":
			logger.Error("Username is a new line", "ip", c.ip)
			io.Copy(c.rwc, strings.NewReader("error: username must be at least 1 character\n"))
			return
		}

		s.mu.Lock()
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
		s.mu.Unlock()

		s.UserAction(c, "join")
		io.Copy(c.rwc, strings.NewReader(fmt.Sprintf("* The room contains: %s\n", s.getUsers(c))))
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
				s.UserAction(c, "leave")
				return
			}
			logger.Error(err.Error())
			continue
		}

		// Don't print empty lines
		switch line {
		case "", "\t":
			continue
		}

		// Send this line to all clients except current client
		s.SendMessage(c, line)
	}
}

func (s *Server) getUsers(c *conn) string {
	var result []string

	s.mu.Lock()
	for cn := range s.activeConn {
		if cn.username == c.username {
			continue
		}
		result = append(result, cn.username)
	}
	s.mu.Unlock()

	return strings.Join(result, " ")
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

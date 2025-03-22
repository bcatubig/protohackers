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
	activeConn map[*conn]struct{}
	clientMsgs chan clientMessage
	mu         sync.Mutex
}

func NewServer(addr string) (*Server, error) {
	l, err := net.Listen("tcp", addr)

	if err != nil {
		return nil, err
	}

	return &Server{
		l:          l,
		activeConn: make(map[*conn]struct{}),
		clientMsgs: make(chan clientMessage),
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

func (s *Server) broadcast(c *conn, msg string) {
	logger.Info("broadcasting message", "msg", msg)
	wg := &sync.WaitGroup{}

	for cn := range s.activeConn {
		if cn == c || !cn.joined {
			continue
		}

		wg.Add(1)

		go func() {
			defer wg.Done()
			s.sendMessage(cn, msg)
		}()
	}

	wg.Wait()
}

func (s *Server) handle(c *conn) {
	defer func() {
		s.removeConn(c)
		c.close()
	}()

	logger.Info("client connected", "ip", c.ip)

	// header
	s.sendMessage(c, "Welcome to budgetchat! What shall I call you?")

	err := s.handleJoin(c)

	if err != nil {
		logger.Error("client failed to join", "error", err.Error())
		s.sendMessage(c, err.Error())
		return
	}

	for {
		if s.inShutdown.Load() {
			return
		}

		buf := bufio.NewReaderSize(c.rwc, 2048)

		line, err := buf.ReadString('\n')
		line = strings.TrimSuffix(line, "\n")
		line = strings.TrimSuffix(line, "\r")

		if err != nil {
			if errors.Is(err, io.EOF) {
				logger.Info("client disconnected", "username", c.username, "ip", c.ip)
				s.handleDisconnect(c)
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
		s.handleData(c, line)
	}
}

func (s *Server) getUsers(c *conn) string {
	var result []string

	logger.Info("getting users", "ip", c.ip)

	for cn := range s.activeConn {
		if cn.username == c.username {
			continue
		}

		if !c.joined {
			continue
		}

		result = append(result, cn.username)
	}

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

func (s *Server) sendMessage(conn *conn, msg string) {
	io.Copy(conn, strings.NewReader(fmt.Sprintf("%s\n", msg)))
}

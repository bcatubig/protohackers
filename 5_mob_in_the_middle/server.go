package main

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	tonyBogusCoinAddr = "7YWHMfk9JZe0LM0g1ZauHuiSxhI"
)

var reWalletAddress = regexp.MustCompile(`(7[a-zA-Z0-9]{25,34}\b)`)

type Server struct {
	l              net.Listener
	activeConn     map[*conn]struct{}
	isShuttingDown atomic.Bool
	mu             sync.Mutex
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

func (s *Server) ListenAndServe() {
	for {
		c, err := s.l.Accept()
		if err != nil {
			logger.Error(err.Error())
			continue
		}

		conn := &conn{
			conn: c,
			ip:   c.LocalAddr().String(),
		}

		s.addConn(conn)

		go s.handle(conn)
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.isShuttingDown.Store(true)
	return nil
}

func (s *Server) handle(c *conn) {
	defer func() {
		s.removeConn(c)
		c.Close()
	}()

	// Connect to upstream
	upstream, err := net.Dial("tcp", "chat.protohackers.com:16963")
	if err != nil {
		logger.Error("failed to connect to upstream", "ip", c.ip)
		return
	}
	defer upstream.Close()

	// handle data from upstream
	go func() {
		reader := bufio.NewReaderSize(upstream, 2048)

		for {
			if s.isShuttingDown.Load() {
				return
			}

			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}

			// check for wallet address
			if reWalletAddress.MatchString(line) {
				splitLine := strings.Split(line, " ")

				for _, w := range splitLine {
					if strings.HasPrefix(w, "7") {
						if len(w) > 35 {
							// not a valid address
							continue
						}
						line = reWalletAddress.ReplaceAllString(line, tonyBogusCoinAddr)
					}
				}
				logger.Info("modified wallet address", "line", line)
			}

			io.Copy(c, strings.NewReader(line))
		}
	}()

	// handle data from client
	reader := bufio.NewReaderSize(c.conn, 2048)

	for {
		if s.isShuttingDown.Load() {
			logger.Info("server shutting down, exiting handler", "ip", c.ip)
			return
		}

		// Inspect line for wallet
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				logger.Info("client disconnected", "ip", c.ip)
				return
			}

			return
		}

		logger.Info("read line", "line", line, "ip", c.ip)

		// check for wallet address
		if reWalletAddress.MatchString(line) {
			splitLine := strings.Split(line, " ")

			for _, w := range splitLine {
				if strings.HasPrefix(w, "7") {
					if len(w) > 35 {
						// not a valid address
						continue
					}
					line = reWalletAddress.ReplaceAllString(line, tonyBogusCoinAddr)
				}
			}
			logger.Info("modified wallet address", "line", line)
		}

		// write to upstream
		io.Copy(upstream, strings.NewReader(line))
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

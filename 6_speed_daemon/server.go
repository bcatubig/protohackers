package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tidwall/btree"
)

type Server struct {
	l          net.Listener
	activeConn map[*conn]struct{}
	db         *btree.Map[string, string]
	inShutdown atomic.Bool
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
		db:         new(btree.Map[string, string]),
		mu:         sync.Mutex{},
	}, nil
}

func (s *Server) ListenAndServe() {
	for {

		c, err := s.l.Accept()
		if err != nil {
			continue
		}

		conn := &conn{
			conn: c,
			ip:   c.RemoteAddr().String(),
		}

		go s.handle(conn)
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.inShutdown.Store(true)

	s.l.Close()

	s.mu.Lock()
	for c := range s.activeConn {
		c.Close()
	}
	s.mu.Unlock()

	return nil
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

func (s *Server) handle(c *conn) {
	defer func() {
		c.Close()
		s.removeConn(c)
	}()

	for {
		//c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))

		buf := make([]byte, 1024)

		_, err := c.Read(buf)

		if err != nil {
			logger.Error(err.Error())
			return
		}

		bufR := bytes.NewBuffer(buf)

		mType := bufR.Next(1)

		var msgType uint8
		_, err = binary.Decode(mType, binary.BigEndian, &msgType)
		if err != nil {
			logger.Error(err.Error())
			return
		}

		logger.Info("got data", "type", msgType)

		if msgType == 64 {
			var interval uint32
			rawInterval := bufR.Next(4)
			binary.Decode(rawInterval, binary.BigEndian, &interval)
			logger.Info("interval", "val", interval)

			if interval > 0 {
				go func() {
					ticker := time.NewTicker((time.Duration(interval) / 10) * time.Second)
					for range ticker.C {
						binary.Write(c, binary.BigEndian, uint8(65))
					}

				}()

			}
		}

	}
}

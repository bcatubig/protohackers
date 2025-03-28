package main

import (
	"bufio"
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
		reader := bufio.NewReader(c)

		var mType MsgType
		err := binary.Read(reader, binary.BigEndian, &mType)

		if err != nil {
			logger.Error(err.Error())
			return
		}

		logger.Info("mType", "type", string(mType))

		type wantHeartbeat struct {
			Interval uint32
		}

		if mType == MsgTypeWantHeartbeat {
			// read next 4 bytes
			msg := &wantHeartbeat{}

			err = binary.Read(reader, binary.BigEndian, msg)

			if err != nil {
				logger.Error(err.Error())
				return
			}

			logger.Info("got wantHeartbeat msg", "data", msg)
			if msg.Interval > 0 {
				logger.Info("registering heartbeat", "interval", msg.Interval, "ip", c.ip)
				s.registerHeartbeat(c, msg.Interval)
			}
		}
	}
}

func (s *Server) registerHeartbeat(c *conn, interval uint32) {
	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Second / 10)
		for range ticker.C {
			//logger.Info("sending heartbeat", "ip", c.ip)
			err := binary.Write(c, binary.BigEndian, MsgHeartbeat)

			if err != nil {
				return
			}
		}
	}()
}

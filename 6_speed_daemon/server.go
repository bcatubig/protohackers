package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"strings"
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

		s.addConn(conn)

		go s.handle(conn)
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.inShutdown.Store(true)

	s.l.Close()

	s.mu.Lock()
	for c := range s.activeConn {
		c.Close()
		s.removeConn(c)
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

	reader := bufio.NewReaderSize(c, 1024)

	for {
		var mType MsgType
		err := binary.Read(reader, binary.BigEndian, &mType)

		if err != nil {
			if errors.Is(err, io.EOF) {
				logger.Info("client disconnected", "ip", c.ip)
				return
			}

			logger.Error(err.Error())
			return
		}

		switch mType {
		case MsgTypeWantHeartbeat:
			msg, err := parseWantHeartbeat(reader)
			if err != nil {
				logger.Error("failed to parse wantHeartbeat msg", "error", err.Error())
				continue
			}

			if msg.Interval > 0 {
				s.registerHeartbeat(c, msg.Interval)
			}
		case MsgTypePlate:
			p, err := parsePlate(reader)
			if err != nil {
				logger.Error("failed to parse plate", "error", err.Error())
				continue
			}
			logger.Info("read plate", "plate", p.Plate, "timestamp", p.Timestamp, "ip", c.ip)
		case MsgTypeIAmCamera:
			logger.Info("got IAmCamera msg")
			camera, err := parseCamera(reader)
			if err != nil {
				logger.Error("failed to parse camera", "error", err.Error())
				continue
			}
			s.addCamera(c, camera)
		case MsgTypeIAmDispatcher:
			logger.Info("got IAmDispatcher msg")
			d, err := parseDispatcher(reader)

			if err != nil {
				logger.Error("failed to parse dispatcher msg", "error", err.Error(), "ip", c.ip)
				continue
			}

			logger.Info("got dispatcher", "num_roads", len(d.Roads), "roads", d.Roads, "ip", c.ip)
		default:
			got, err := hex.DecodeString(string(mType))
			if err != nil {
				logger.Error(err.Error())
				continue
			}
			logger.Info("got unknown message type", "type", string(got))
		}
	}
}

func (s *Server) addCamera(c *conn, camera *Camera) {
	logger.Info("registering camera", "road", camera.Road, "mile", "camera.Mile", "limit_mph", camera.LimitMPH, "ip", c.ip)
}

func (s *Server) registerHeartbeat(c *conn, interval uint32) {
	logger.Info("registering heartbeat", "interval", interval, "ip", c.ip)
	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Second / 10)
		for range ticker.C {
			err := binary.Write(c, binary.BigEndian, MsgHeartbeat)

			if err != nil {
				return
			}
		}
	}()
}

func (s *Server) sendError(c *conn, msg string) {
	err := binary.Write(c, binary.BigEndian, MsgTypeError)

	if err != nil {
		return
	}

	_, err = io.Copy(c, strings.NewReader(msg))

	if err != nil {
		return
	}
}

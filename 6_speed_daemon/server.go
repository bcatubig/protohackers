package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	l          net.Listener
	activeConn map[*conn]struct{}
	inShutdown atomic.Bool
	mu         sync.Mutex
	dispatcher *DispatcherService
}

func NewServer(addr string) (*Server, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	dispatcher := NewDispatcherService()

	return &Server{
		l:          l,
		activeConn: make(map[*conn]struct{}),
		mu:         sync.Mutex{},
		dispatcher: dispatcher,
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
				logger.Info("client disconnected")
				return
			}

			logger.Error(err.Error())
			return
		}

		switch mType {
		case MsgTypeIAmCamera:
			logger.Info("camera connected", "ip", c.ip)
			c.isCamera = true
			camera, err := parseCamera(reader)
			if err != nil {
				logger.Error("failed to parse camera", "error", err.Error())
				return
			}
			logger.Info("new camera", "road", camera.Road, "mile", camera.Mile, "limit_mph", camera.LimitMPH)
		case MsgTypeIAmDispatcher:
			logger.Info("dispatcher connected")
			c.isDispatcher = true
			d, err := parseDispatcher(reader)
			if err != nil {
				logger.Error("failed to parse dispatcher msg", "error", err.Error())
				return
			}
			logger.Info("got dispatcher", "num_roads", len(d.Roads), "roads", d.Roads)
			s.dispatcher.RegisterDispatcher(d)

		case MsgTypeWantHeartbeat:
			logger.Info("got heartbeat request")
			if c.hasHeartbeat {
				logger.Error("client already has an active heartbeat check")
			}
			c.hasHeartbeat = true
			msg, err := parseWantHeartbeat(reader)
			if err != nil {
				logger.Error("failed to parse wantHeartbeat msg", "error", err.Error())
				return
			}

			if msg.Interval > 0 {
				s.registerHeartbeat(c, msg.Interval)
			}
		case MsgTypePlate:
			if !c.isCamera {
				s.sendError(c, fmt.Sprintf("%s is not a valid camera: cannot send plate data", c.ip))
				continue
			}
			p, err := parsePlate(reader)
			if err != nil {
				logger.Error("failed to parse plate", "error", err.Error())
			}
			logger.Info("read plate", "plate", p.Plate, "timestamp", p.Timestamp)
		}
	}
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

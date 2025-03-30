package main

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
)

var ErrServerClosed = errors.New("tcp: Server closed")

type contextKey struct {
	name string
}

func (s contextKey) String() string {
	return "tcp server context value " + s.name
}

var serverContextKey = &contextKey{"tcp-server"}

type Server struct {
	addr       string
	activeConn map[*conn]struct{}
	inShutdown atomic.Bool
	mu         sync.Mutex
}

type onceCloseListener struct {
	net.Listener
	once       sync.Once
	closeError error
}

func (oc *onceCloseListener) Close() error {
	oc.once.Do(oc.close)
	return oc.closeError
}

func (oc *onceCloseListener) close() {
	oc.closeError = oc.Listener.Close()
}

func (s *Server) ListenAndServe() error {
	addr := s.addr

	if addr == "" {
		addr = "0.0.0.0:8000"
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return s.Serve(ln)
}

func (s *Server) Serve(l net.Listener) error {
	l = &onceCloseListener{Listener: l}
	defer l.Close()

	baseCtx := context.Background()

	ctx := context.WithValue(baseCtx, serverContextKey, s)

	for {
		rw, err := l.Accept()
		if err != nil {
			if s.shuttingDown() {
				return ErrServerClosed
			}
			logger.Error(err.Error())
			continue
		}
		c := s.newConn(rw)
		go c.serve(ctx)
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.inShutdown.Store(true)

	return nil
}

func (s *Server) newConn(rwc net.Conn) *conn {
	c := &conn{
		server: s,
		rwc:    rwc,
	}

	return c
}

func (s *Server) shuttingDown() bool {
	return s.inShutdown.Load()
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

// func (s *Server) registerHeartbeat(c *conn, interval uint32) {
// 	logger.Info("registering heartbeat", "interval", interval, "ip", c.ip)
// 	go func() {
// 		ticker := time.NewTicker(time.Duration(interval) * time.Second / 10)
// 		for range ticker.C {
// 			err := binary.Write(c, binary.BigEndian, MsgHeartbeat)
// 			if err != nil {
// 				return
// 			}
// 		}
// 	}()
// }
//
// func (s *Server) sendError(c *conn, msg string) {
// 	err := binary.Write(c, binary.BigEndian, MsgTypeError)
// 	if err != nil {
// 		return
// 	}
//
// 	_, err = io.Copy(c, strings.NewReader(msg))
// 	if err != nil {
// 		return
// 	}
// }

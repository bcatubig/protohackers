package main

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/tidwall/btree"
)

type Server struct {
	l              *net.UDPConn
	db             *btree.Map[string, string]
	isShuttingDown atomic.Bool
	chanDone       chan struct{}

	mu sync.Mutex
}

func NewServer(addr string) (*Server, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	l, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}

	db := new(btree.Map[string, string])

	s := &Server{
		l:        l,
		db:       db,
		mu:       sync.Mutex{},
		chanDone: make(chan struct{}),
	}

	return s, nil
}

func (s *Server) ListenAndServe() {
	logger.Info("listening for udp connections")

	for {
		if s.isShuttingDown.Load() {
			logger.Info("server is shutting down. exiting")
			return
		}

		buf := make([]byte, 1024)

		n, addr, err := s.l.ReadFromUDP(buf)
		if err != nil {
			logger.Error(err.Error())
			continue
		}

		logger.Info("raw message", "data", string(buf[:n]))

		bb := bytes.NewBuffer(buf)

		data, err := bb.ReadString('\x00')
		if err != nil {
			logger.Error(err.Error())
			continue
		}

		data = strings.TrimRight(data, "\x00")

		client := &client{
			addr: addr,
			ip:   addr.String(),
			data: data,
		}

		logger.Info("got message", "addr", addr.String(), "length", n, "msg", client.data)

		go s.handle(client)
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	logger.Info("shutting down server")
	s.isShuttingDown.Store(true)
	logger.Info("killing udp connection")
	s.l.Close()
	return nil
}

func (s *Server) handle(c *client) {
	if strings.Contains(c.data, "version") {
		s.handleVersion(c)
	} else if strings.Contains(c.data, "=") {
		s.handleInsert(c)
	} else {
		s.handleRetrieve(c)
	}
}

func (s *Server) sendData(c *client, data string) {
	s.l.WriteToUDP([]byte(fmt.Sprintf("%s", data)), c.addr)
}

package main

import (
	"encoding/binary"
	"net"
	"time"
)

type Server struct {
	ln net.Listener
}

func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", "0.0.0.0:8000")
	if err != nil {
		return err
	}
	defer ln.Close()

	s.ln = ln

	return s.Serve(ln)
}

func (s *Server) Serve(ln net.Listener) error {
	for {
		c, err := ln.Accept()
		if err != nil {
			logger.Error("accept error", "error", err.Error())
			time.Sleep(5 * time.Millisecond)
			continue
		}

		go s.handle(c)
	}
}

func (s *Server) handle(c net.Conn) {
	ip := c.RemoteAddr().String()
	defer func() {
		logger.Info("client disconnected", "ip", ip)
		c.Close()
	}()

	var b [1]byte
	for {
		n, err := c.Read(b[:])
		if n < 1 || err != nil {
			return
		}

		var mType MsgType
		n, err = binary.Decode(b[:], binary.BigEndian, &mType)
		if n < 1 || err != nil {
			continue
		}

		switch mType {
		case WantHeartbeatMsg:
			logger.Info("got wantHeartbeat msg")
		case IAmCameraMsg:
			logger.Info("got IAmCamera msg")
		case IAmDispatcherMsg:
			logger.Info("Got IAmDispatcher msg")
		case PlateMsg:
			logger.Info("Got PlateMsg")

		}
	}
}

package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"time"
)

type Server struct {
	l net.Listener
}

func NewServer(addr string) (*Server, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Server{l: l}, nil
}

func (s *Server) ListenAndServe() error {
	for {
		c, err := s.l.Accept()
		if err != nil {
			logger.Error(err.Error())
			continue
		}

		logger.Info("new connection", "ip", c.RemoteAddr().String())

		c.SetDeadline(time.Time{})
		go s.handle(c)
	}
}

func (s *Server) handle(c net.Conn) {
	defer c.Close()

	defer func() {
		if r := recover(); r != nil {
			logger.Error("recovered from panic", "r", r)
		}
	}()

	db := NewDB()

	for {
		buf := new(bytes.Buffer)

		w, err := io.CopyN(buf, c, 9)

		if err != nil {
			if errors.Is(err, io.EOF) {
				logger.Info("EOF. client disconnected", "ip", c.RemoteAddr().String())
				return
			}
			logger.Error("error copying data", "error", err.Error(), "ip", c.RemoteAddr().String())
			continue
		}

		if w < 9 {
			logger.Error("did not read 9 bytes of data", "ip", c.RemoteAddr().String())
			continue
		}

		op := buf.Next(1)
		num1 := buf.Next(4)
		num2 := buf.Next(4)

		switch string(op) {
		case "I":
			timestamp := parseInt32(num1)
			cents := parseInt32(num2)
			logger.Info("got insert message", "timestamp", timestamp, "cents", cents, "ip", c.RemoteAddr().String())
			db.Insert(timestamp, cents)
		case "Q":
			minTime := parseInt32(num1)
			maxTime := parseInt32(num2)
			logger.Info("got query message", "minTime", minTime, "maxTime", maxTime, "ip", c.RemoteAddr().String())

			if minTime > maxTime {
				logger.Error("minTime greater than maxTime", "minTime", minTime, "maxTime", maxTime, "ip", c.RemoteAddr().String())
				binary.Write(c, binary.BigEndian, int32(0))
				continue
			}

			mean := db.Mean(minTime, maxTime)

			binary.Write(c, binary.BigEndian, int32(mean))

		default:
			logger.Error("invalid message")
		}
	}
}

func parseInt32(data []byte) int32 {
	var result int32

	reader := bytes.NewReader(data)
	err := binary.Read(reader, binary.BigEndian, &result)
	if err != nil {
		panic(err)
	}

	return result
}

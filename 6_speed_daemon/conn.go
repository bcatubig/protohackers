package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

type conn struct {
	server *Server
	rwc    net.Conn
	ip     string

	camera *Camera

	isDispatcher bool
}

func (c *conn) close() {
	c.rwc.Close()
}

func (c *conn) serve() {
	if ra := c.rwc.RemoteAddr(); ra != nil {
		c.ip = ra.String()
	}

	logger.Info("client connected", "ip", c.ip)

	defer func() {
		if err := recover(); err != nil {
			logger.Error("panic in handler", "error", err, "ip", c.ip)
		}
		logger.Info("client disconnected", "ip", c.ip)
		c.close()
	}()

	b := make([]byte, 256)
	for {
		n, err := c.rwc.Read(b)
		if err != nil {
			return
		}

		if n < 1 {
			continue
		}

		fmt.Printf("buf: %v\n", b[:n])
	}
}

func (c *conn) registerHeartbeat(interval uint32) {
	logger.Info("registering heartbeat", "interval", interval, "ip", c.ip)
	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Second / 10)
		for range ticker.C {
			err := binary.Write(c.rwc, binary.BigEndian, MsgHeartbeat)
			if err != nil {
				return
			}
		}
	}()
}

func (c *conn) sendError(msg string) {
	err := binary.Write(c.rwc, binary.BigEndian, MsgError)
	if err != nil {
		return
	}

	_, err = io.Copy(c.rwc, strings.NewReader(msg+"\n"))
	if err != nil {
		return
	}
}

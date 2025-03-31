package main

import (
	"encoding/binary"
	"io"
	"net"
	"strings"
	"time"
)

type conn struct {
	server *Server
	rwc    net.Conn
	ip     string

	isCamera     bool
	isDispatcher bool
}

func (c *conn) close() {
	c.rwc.Close()
}

func (c *conn) Write(b []byte) (int, error) {
	return c.rwc.Write(b)
}

func (c *conn) Read(b []byte) (int, error) {
	return c.rwc.Read(b)
}

func (c *conn) serve() {
	if ra := c.rwc.RemoteAddr(); ra != nil {
		c.ip = ra.String()
	}

	defer func() {
		if err := recover(); err != nil {
			logger.Error("conn panic", "error", err)
		}
		logger.Info("client disconnected", "ip", c.ip)
		c.close()
	}()

	// Read data from connection

	var msgType MsgType
	for {
		err := binary.Read(c, binary.BigEndian, &msgType)
		if err != nil {
			logger.Error(err.Error())
			return
		}

		switch msgType {
		case MsgWantHeartbeat:
			logger.Info("got heartbeat request")
			hb, err := parseWantHeartbeat(c)
			if err != nil {
				logger.Error(err.Error())
				return
			}
			if hb.Interval > 0 {
				c.registerHeartbeat(hb.Interval)
			}
		}
	}
}

func (c *conn) registerHeartbeat(interval uint32) {
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

func (c *conn) sendError(msg string) {
	err := binary.Write(c, binary.BigEndian, MsgError)
	if err != nil {
		return
	}

	_, err = io.Copy(c, strings.NewReader(msg))
	if err != nil {
		return
	}
}

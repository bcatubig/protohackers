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

	isCamera     bool
	isDispatcher bool

	camera     *Camera
	dispatcher *Dispatcher
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
			hb, err := parseWantHeartbeat(c)
			if err != nil {
				logger.Error(err.Error())
				return
			}
			if hb.Interval > 0 {
				c.registerHeartbeat(hb.Interval)
			}
		case MsgIAmCamera:
			if c.isCamera || c.isDispatcher {
				c.sendError("already registered, cannot re-register")
				continue
			}
			camera, err := parseCamera(c)
			if err != nil {
				logger.Error("failed to parse camera", "error", err.Error())
				continue
			}
			fmt.Println(camera)
		case MsgIAmDispatcher:
			if c.isCamera || c.isDispatcher {
				c.sendError("already registered, cannot re-register")
				continue
			}
			dispatcher, err := parseDispatcher(c)
			if err != nil {
				logger.Error("failed to parse dispatcher", "error", err.Error())
			}
			dispatcher.conn = c
			c.server.dispatcherSvc.RegisterDispatcher(dispatcher)
		case MsgPlate:
			if !c.isCamera {
				logger.Error("plate event from non-camera")
				continue
			}

			p, err := parsePlate(c)
			if err != nil {
				logger.Error("failed to parse plate", "error", err.Error())
				continue
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

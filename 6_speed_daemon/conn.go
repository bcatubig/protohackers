package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
)

type conn struct {
	server *Server
	rwc    net.Conn
	ip     string
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

	for {
		b := make([]byte, 1024)
		n, err := c.rwc.Read(b)
		if err != nil {
			return
		}

		// Process msg
		buf := bytes.NewBuffer(b[:n])

		var mType MsgType
		err = binary.Read(buf, binary.BigEndian, &mType)
		if err != nil {
			logger.Error("error reading msg type", "error", err)
			continue
		}

		switch mType {
		case MsgWantHeartbeat:
			logger.Info("got heartbeat request")
		default:
			c.sendError(fmt.Sprintf("invalid msg: %v", mType))
			continue
		}

		logger.Info("got data", "n", n, "msg_type", mType)
		fmt.Println("after", buf.String())
	}
}

//	func (c *conn) registerHeartbeat(interval uint32) {
//		logger.Info("registering heartbeat", "interval", interval, "ip", c.ip)
//		go func() {
//			ticker := time.NewTicker(time.Duration(interval) * time.Second / 10)
//			for range ticker.C {
//				err := binary.Write(c, binary.BigEndian, MsgHeartbeat)
//				if err != nil {
//					return
//				}
//			}
//		}()
//	}
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

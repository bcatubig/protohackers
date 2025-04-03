package main

import (
	"bytes"
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

	camera *Camera

	isCamera     bool
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

	for {
		b := make([]byte, 1024)
		n, err := c.rwc.Read(b)
		if err != nil {
			return
		}

		if n < 1 {
			continue
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
			hb, err := parseWantHeartbeat(buf)
			if err != nil {
				logger.Error("failed to parse WantHeartbeat request", "error", err.Error(), "ip", c.ip)
				continue
			}
			if hb.Interval > 0 {
				c.registerHeartbeat(hb.Interval)
			}
		case MsgIAmDispatcher:
			d, err := parseDispatcher(buf)
			if err != nil {
				logger.Error("failed to parse IAmDispatcher request", "error", err, "ip", c.ip)
				continue
			}
			c.isDispatcher = true
			c.server.dispatcherService.RegisterDispatcher(c, d.Roads)
		case MsgIAmCamera:
			cam, err := parseCamera(buf)
			if err != nil {
				logger.Error("failed to parse camera", "error", err.Error(), "ip", c.ip)
				continue
			}
			c.isCamera = true
			c.camera = cam
			c.server.dispatcherService.RegisterCamera(c, cam)
		case MsgPlate:
			if c.camera == nil {
				logger.Info("camera event from non-camera", "ip", c.ip)
				continue
			}
			p, err := parsePlate(buf)
			if err != nil {
				logger.Error("failed to parse plate", "error", err.Error(), "ip", c.ip)
				continue
			}
			c.server.dispatcherService.NewEvent(&CameraEvent{
				Timestamp: p.Timestamp,
				Plate:     p.Plate,
				Mile:      c.camera.Mile,
				Road:      c.camera.Road,
				LimitMPH:  c.camera.LimitMPH,
			})
		default:
			c.sendError("invalid msg")
			continue
		}
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

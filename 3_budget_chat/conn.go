package main

import (
	"fmt"
	"net"
)

type conn struct {
	rwc      net.Conn
	ip       string
	username string
	joined   bool
}

func (c *conn) Write(b []byte) (int, error) {
	return c.rwc.Write(b)
}

func (c *conn) Read(b []byte) (int, error) {
	return c.rwc.Read(b)
}

func (c conn) String() string {
	return fmt.Sprintf("%s - %s", c.ip, c.username)
}

func (c *conn) close() {
	logger.Info("closing connection", "ip", c.ip)
	c.rwc.Close()
}

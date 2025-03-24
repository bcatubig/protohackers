package main

import (
	"net"
)

type conn struct {
	conn net.Conn
	ip   string
}

func (c *conn) Read(b []byte) (int, error) {
	return c.conn.Read(b)
}

func (c *conn) Write(b []byte) (int, error) {
	return c.conn.Write(b)
}

func (c *conn) Close() error {
	return c.conn.Close()
}

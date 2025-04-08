package tcp

import (
	"net"
)

type Conn struct {
	server     *Server
	rwc        net.Conn
	remoteAddr string
}

func (c *Conn) Read(b []byte) (int, error) {
	return c.rwc.Read(b)
}

func (c *Conn) Write(b []byte) (int, error) {
	return c.rwc.Write(b)
}

func (c *Conn) Close() error {
	return c.rwc.Close()
}

func (c *Conn) close() {
	c.rwc.Close()
}

func (c *Conn) serve() {
	defer func() {
		c.close()
		c.server.removeConn(c)
		c.server.logger.Info("client disconnected", "ip", c.remoteAddr)
	}()

	if ra := c.rwc.RemoteAddr(); ra != nil {
		c.remoteAddr = ra.String()
	}

	c.server.logger.Info("client connected", "ip", c.remoteAddr)

	c.server.Handler.Serve(c)
}

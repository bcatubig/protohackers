package server

import (
	"context"
	"net"
	"time"
)

type Conn struct {
	server     *Server
	rwc        net.Conn
	remoteAddr string
}

func (c *Conn) close() {
	c.rwc.Close()
}

func (c *Conn) serve(ctx context.Context) {
	defer func() {
		c.close()
		c.server.removeConn(c)
	}()

	if ra := c.rwc.RemoteAddr(); ra != nil {
		c.remoteAddr = ra.String()
	}

	select {
	case <-ctx.Done():
		return
	case <-time.Tick(5 * time.Millisecond):
	}

	c.server.Handler.Serve(c.rwc)
}

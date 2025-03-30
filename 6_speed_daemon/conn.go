package main

import (
	"context"
	"net"
)

type conn struct {
	server     *Server
	cancelCtx  context.CancelFunc
	rwc        net.Conn
	remoteAddr string
}

func (c *conn) serve(ctx context.Context) {}

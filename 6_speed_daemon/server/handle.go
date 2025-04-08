package server

import (
	"net"
)

type Handler interface {
	Serve(net.Conn)
}

type HandlerFunc func(net.Conn)

func (f HandlerFunc) Serve(c net.Conn) {
	f(c)
}

package main

import "net"

type client struct {
	addr *net.UDPAddr
	ip   string
	data string
}

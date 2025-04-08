package main

import (
	"encoding/binary"
	"sync"

	"github.com/bcatubig/protohackers/6_speed_daemon/server"
	"github.com/tidwall/btree"
)

type Mux struct {
	mu   sync.Mutex
	tree btree.Map[uint8, server.Handler]
}

func (m *Mux) Register(op uint8, h server.Handler) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tree.Set(op, h)
}

func (m *Mux) Serve(c *server.Conn) {
	// find the handler for the incoming connection
	b := make([]byte, 256)
	for {
		n, err := c.Read(b)
		if err != nil {
			return
		}

		if n < 1 {
			continue
		}

		var mType uint8
		_, err = binary.Decode(b[:1], binary.BigEndian, &mType)
		if err != nil {
			logger.Error(err.Error())
			continue
		}

		// find handler
		h, ok := m.tree.Get(mType)

		if !ok {
			continue
		}

		h.Serve(c)
	}
}

func NewMux() *Mux {
	return &Mux{}
}

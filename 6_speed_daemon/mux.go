package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"sync"

	"github.com/bcatubig/protohackers/pkg/tcp"
)

type Mux struct {
	mu sync.Mutex
}

func NewMux() *Mux {
	return &Mux{}
}

func (m *Mux) Serve(c *tcp.Conn) {
	// find the handler for the incoming connection
	br := bufio.NewReader(c)

	for {
		var mType MsgType
		err := binary.Read(br, binary.BigEndian, &mType)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}

			logger.Error(err.Error())
			continue
		}

		switch mType {
		case WantHeartbeatMsg:
			handleWantHeartbeat(br)
		}
	}
}

func handleWantHeartbeat(r io.Reader) {
	var interval uint32
	err := binary.Read(r, binary.BigEndian, &interval)
	if err != nil {
		logger.Info("nope")
		return
	}

	if interval > 0 {
		logger.Info("got heartbeat interval", "interval", interval)
	}
}

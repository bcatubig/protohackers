package main

import (
	"fmt"
	"strings"
)

func (s *Server) handleInsert(c *client) {
	k, v, _ := strings.Cut(c.data, "=")

	s.mu.Lock()
	s.db.Set(k, v)
	s.mu.Unlock()
}

func (s *Server) handleRetrieve(c *client) {
	val, ok := s.db.Get(c.data)

	if !ok {
		s.sendData(c, "key=")
		return
	}

	s.sendData(c, fmt.Sprintf("%s=%s", c.data, val))
}

func (s *Server) handleVersion(c *client) {
	s.sendData(c, "version=bcatubig Key-Value Store 1.0")
}

package main

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
)

type clientMessageType int

const (
	clientJoined clientMessageType = iota
	clientDisconnected
	clientData
)

type clientMessage struct {
	msgType clientMessageType
	conn    *conn
	data    string
}

func (m clientMessage) String() string {
	return fmt.Sprintf("%d %s %s", m.msgType, m.conn.username, m.data)
}

func (s *Server) handleJoin(conn *conn) error {
	// read in username
	buf := bufio.NewReaderSize(conn, 64)
	username, err := buf.ReadString('\n')

	if err != nil {
		logger.Error("error reading username", "error", err.Error(), "ip", conn.ip)
		return errors.New("error reading username")
	}

	username = strings.TrimSuffix(username, "\n")

	logger.Info("parsed username", "username", username)

	if username == "" {
		logger.Error("Username is a new line", "ip", conn.ip)
		return errors.New("username must be at least 1 character")
	}

	if len(username) > 64 {
		logger.Error("username is too long", "username", username)
		return fmt.Errorf("username %s is too long: max 64 chars", username)
	}

	for c := range s.activeConn {
		if c == conn {
			continue
		}

		if username == conn.username {
			logger.Error("duplicate username", "username", c.username, "ip", c.ip)
			return fmt.Errorf("* username %s is already taken", username)
		}
	}

	conn.username = username
	conn.joined = true

	s.broadcast(conn, fmt.Sprintf("* %s has entered the room", conn.username))
	s.sendMessage(conn, fmt.Sprintf("* The room contains: %s", s.getUsers(conn)))

	return nil
}

func (s *Server) handleDisconnect(conn *conn) {
	conn.joined = false
	s.broadcast(conn, fmt.Sprintf("* %s has left the room", conn.username))
}

func (s *Server) handleData(conn *conn, data string) {
	if strings.HasPrefix("*", data) {
		return
	}
	s.broadcast(conn, fmt.Sprintf("[%s] %s", conn.username, data))
}

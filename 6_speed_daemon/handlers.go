package main

import (
	"bytes"
	"encoding/binary"
)

type MsgType uint8

const (
	MsgTypeError         MsgType = 16
	MsgTypePlate         MsgType = 32
	MsgTypeTicket        MsgType = 33
	MsgTypeWantHeartbeat MsgType = 64
	MsgTypeHeartbeat     MsgType = 65
	MsgTypeIAmCamera     MsgType = 128
	MsgTypeIAmDispatcher MsgType = 129
)

func parseMsgType(b *bytes.Buffer) (MsgType, error) {
	var result MsgType

	data := b.Next(1)

	_, err := binary.Decode(data, binary.BigEndian, &result)
	if err != nil {
		return 0, err
	}

	return result, nil
}

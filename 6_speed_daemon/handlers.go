package main

import (
	"bytes"
	"encoding/binary"
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

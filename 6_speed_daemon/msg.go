package main

type MsgType uint8

const (
	WantHeartbeatMsg MsgType = 0x40
	HeartbeatMsg     MsgType = 0x41
)

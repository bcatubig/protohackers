package main

type MsgType uint8

const (
	ErrorMsg         MsgType = 0x10
	PlateMsg         MsgType = 0x20
	TicketMsg        MsgType = 0x21
	WantHeartbeatMsg MsgType = 0x40
	HeartbeatMsg     MsgType = 0x41
	IAmCameraMsg     MsgType = 0x80
	IAmDispatcherMsg MsgType = 0x81
)

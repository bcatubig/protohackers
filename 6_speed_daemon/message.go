package main

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

var MsgHeartbeat uint8 = 65

type MsgError struct {
	Msg string
}

type MsgPlate struct {
	Plate     string
	Timestamp uint32
}

type MsgTicket struct {
	Plate          string
	Road           uint16
	Mile1          uint16
	Mile1Timestamp uint32
	Mile2          uint16
	Mile2Timestamp uint32
	Speed          uint16
}

type MsgWantHeartbeat struct {
	Interval uint32
}

type MsgIAmCamera struct {
	Road     uint16
	Mile     uint16
	LimitMPH uint16
}

type MsgIAmDispatcher struct {
	NumRoads uint8
	Roads    []uint16
}

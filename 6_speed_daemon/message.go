package main

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

type MsgHeartbeat struct{}

type MsgIAmCamera struct {
	Road     uint16
	Mile     uint16
	LimitMPH uint16
}

type MsgIAmDispatcher struct {
	NumRoads uint8
	Roads    []uint16
}

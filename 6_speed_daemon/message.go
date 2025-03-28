package main

import (
	"encoding/binary"
	"fmt"
	"io"
)

type (
	MsgType       uint8
	Road          uint16
	Mile          uint16
	SpeedLimitMPH uint16
)

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

type Plate struct {
	Plate     string
	Timestamp uint32
}

func parsePlate(r io.Reader) (*Plate, error) {
	result := &Plate{}

	var pLength uint8
	binary.Read(r, binary.BigEndian, &pLength)

	pData := make([]byte, pLength)
	_, err := io.ReadFull(r, pData)
	if err != nil {
		return nil, err
	}

	result.Plate = string(pData)

	var pTimestamp uint32
	err = binary.Read(r, binary.BigEndian, &pTimestamp)
	if err != nil {
		return nil, err
	}

	result.Timestamp = pTimestamp

	return result, nil
}

type Ticket struct {
	Plate          string
	Road           uint16
	Mile1          uint16
	Mile1Timestamp uint32
	Mile2          uint16
	Mile2Timestamp uint32
	Speed          uint16
}

type WantHeartbeat struct {
	Interval uint32
}

func parseWantHeartbeat(r io.Reader) (*WantHeartbeat, error) {
	result := &WantHeartbeat{}

	err := binary.Read(r, binary.BigEndian, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

type Camera struct {
	Road     uint16
	Mile     uint16
	LimitMPH uint16
}

func parseCamera(r io.Reader) (*Camera, error) {
	result := &Camera{}

	// parse road
	err := binary.Read(r, binary.BigEndian, result.Road)
	if err != nil {
		return nil, err
	}

	err = binary.Read(r, binary.BigEndian, result.Mile)
	if err != nil {
		return nil, err
	}

	err = binary.Read(r, binary.BigEndian, result.LimitMPH)
	if err != nil {
		return nil, err
	}

	return result, nil
}

type Dispatcher struct {
	Roads []uint16
}

func (d Dispatcher) String() string {
	return fmt.Sprintf("dispatcher: [%v]", d.Roads)
}

func parseDispatcher(r io.Reader) (*Dispatcher, error) {
	result := &Dispatcher{
		Roads: make([]uint16, 0),
	}

	var numRoads uint8
	err := binary.Read(r, binary.BigEndian, &numRoads)
	if err != nil {
		return nil, err
	}

	for range numRoads {
		var roadNum uint16
		err = binary.Read(r, binary.BigEndian, &roadNum)
		if err != nil {
			return nil, err
		}

		result.Roads = append(result.Roads, roadNum)
	}

	return result, nil
}

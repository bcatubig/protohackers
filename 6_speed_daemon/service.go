package main

import (
	"fmt"
	"math"
	"sort"
	"sync"
)

type DispatcherService struct {
	mu            sync.Mutex
	dispatchers   map[uint16][]*conn
	cameras       map[uint16][]*Camera
	events        []*CameraEvent
	ticketHistory []*TicketEvent
}

func NewDispatcherService() *DispatcherService {
	svc := &DispatcherService{
		mu:          sync.Mutex{},
		dispatchers: make(map[uint16][]*conn),
		cameras:     make(map[uint16][]*Camera),
		events:      make([]*CameraEvent, 0),
	}

	return svc
}

func (s *DispatcherService) RegisterDispatcher(c *conn, roads []uint16) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, r := range roads {
		s.dispatchers[r] = append(s.dispatchers[r], c)
	}

	logger.Info("registered dispatcher", "ip", c.ip, "roads", roads)
}

func (s *DispatcherService) RegisterCamera(c *conn, camera *Camera) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cameras[camera.Road] = append(s.cameras[camera.Road], camera)
	logger.Info("registered camera", "road", camera.Road, "mile", camera.Mile, "limit_mph", camera.LimitMPH, "ip", c.ip)
}

type TicketEvent struct {
	Day   int
	Plate string
}

type CameraEvent struct {
	Timestamp uint32
	Plate     string
	Mile      uint16
	Road      uint16
	LimitMPH  uint16
}

func (s *DispatcherService) NewEvent(e *CameraEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, e)

	go s.handleEvent(e)
}

func (s *DispatcherService) getPlateEvents(plate string) []*CameraEvent {
	s.mu.Lock()
	defer s.mu.Unlock()

	var result []*CameraEvent
	for _, ce := range s.events {
		if ce.Plate == plate {
			result = append(result, ce)
		}
	}

	return result
}

func (s *DispatcherService) getTicketEvents(plate string) []*TicketEvent {
	s.mu.Lock()
	defer s.mu.Unlock()

	var result []*TicketEvent
	for _, t := range s.ticketHistory {
		if t.Plate == plate {
			result = append(result, t)
		}
	}

	return result
}

func (s *DispatcherService) getDispatcher(road uint16) *conn {
	s.mu.Lock()
	defer s.mu.Unlock()

	dispatchers, ok := s.dispatchers[road]

	if !ok {
		return nil
	}

	return dispatchers[0]
}

func (s *DispatcherService) handleEvent(e *CameraEvent) {
	plate := e.Plate

	events := s.getPlateEvents(plate)

	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp < events[j].Timestamp
	})

	if len(events)%2 != 0 {
		return
	}

	for i := 0; i < len(events); i += 2 {
		e1 := events[i]
		e2 := events[i+1]

		fmt.Printf("%v\n", e1)
		fmt.Printf("%v\n", e2)

		time := math.Floor(float64(e1.Timestamp) - float64(e2.Timestamp))
		distance := math.Floor(float64(e1.Mile) - float64(e2.Mile))

		fmt.Println("Time:", time)
		fmt.Println("Distance:", distance)

		speed := speed(int(distance), int(time))

		if speed-float64(e1.LimitMPH) > 0.5 {
			// Get the day
			day := currentDay(e2.Timestamp)

			// Get current tickets
			tickets := s.getTicketEvents(plate)

			for _, t := range tickets {
				if t.Day == day {
					// Don't send ticket as they already have one
					return
				}
			}

			// Send ticket

			// Grab any dispatcher for road
			d := s.getDispatcher(e1.Road)

			if d == nil {
				// Save ticket to be sent later
				panic("can't save ticket: not implemented")
			}

			logger.Info("sending ticket", "plate", plate, "speed", speed, "day", day)
		}

	}
}

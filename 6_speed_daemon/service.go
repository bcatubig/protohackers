package main

import (
	"sync"
)

type EventType int

const (
	RegisterDispatcher = iota
	RegisterCamera
	DisconnectDispatcher
	DisconnectCamera
	PlateEvent
	TicketEvent
)

type DispatcherService struct {
	wg *sync.WaitGroup

	cameras       map[*Camera]struct{}
	dispatchers   map[*Dispatcher]chan TicketRecord
	events        map[*Plate]map[uint16][]SeenHistory
	ticketHistory map[*Plate][]PlateTicketHistory

	cameraEvents     chan CameraEvent
	dispatcherEvents chan DispatcherEvent
}

func NewDispatcherService() *DispatcherService {
	d := &DispatcherService{
		wg:               &sync.WaitGroup{},
		cameras:          make(map[*Camera]struct{}),
		dispatchers:      make(map[*Dispatcher]chan TicketRecord),
		events:           make(map[*Plate]map[uint16][]SeenHistory),
		ticketHistory:    make(map[*Plate][]PlateTicketHistory),
		cameraEvents:     make(chan CameraEvent),
		dispatcherEvents: make(chan DispatcherEvent),
	}

	return d
}

type DispatcherEvent struct {
	eventType EventType

	dispatcher *Dispatcher
	chanRecord chan TicketRecord
}

type CameraEvent struct {
	eventType EventType

	camera *Camera
	plate  *Plate
}

type TicketRecord struct {
	day    int
	ticket Ticket
}

type PlateTicketHistory struct {
	Day  int
	Road uint16
}

type SeenHistory struct {
	Timestamp uint32
	Camera    *Camera
}

func (s *DispatcherService) Start() {
	go func() {
		s.wg.Add(1)
		defer s.wg.Done()

		logger.Info("dispatcher_service: handling dispatcher events")
		for event := range s.dispatcherEvents {
			s.handleDispatcherEvent(event)
		}
	}()

	go func() {
		s.wg.Add(1)
		defer s.wg.Done()
		logger.Info("dispatcher_service: listening for camera events")
		for event := range s.cameraEvents {
			s.handleCameraEvent(event)
		}
	}()
}

func (s *DispatcherService) Shutdown() {
	close(s.cameraEvents)
	close(s.dispatcherEvents)
	s.wg.Done()
}

func (s *DispatcherService) RegisterDispatcher(d *Dispatcher) chan TicketRecord {
	chanRecord := make(chan TicketRecord)

	s.dispatcherEvents <- DispatcherEvent{
		eventType:  RegisterDispatcher,
		dispatcher: d,
		chanRecord: chanRecord,
	}

	return chanRecord
}

func (s *DispatcherService) DisconnectDispatcher(d *Dispatcher) {
	s.dispatcherEvents <- DispatcherEvent{
		eventType:  DisconnectDispatcher,
		dispatcher: d,
	}
}

func (s *DispatcherService) RegisterCamera(c *Camera) {
	s.cameraEvents <- CameraEvent{
		eventType: RegisterCamera,
		camera:    c,
	}
}

func (s *DispatcherService) PlateEvent(p *Plate, c *Camera) {
	s.cameraEvents <- CameraEvent{
		eventType: PlateEvent,
		camera:    c,
		plate:     p,
	}
}

func (s *DispatcherService) DisconnectCamera(c *Camera) {
	s.cameraEvents <- CameraEvent{
		eventType: DisconnectCamera,
		camera:    c,
	}
}

func (s *DispatcherService) handleCameraEvent(e CameraEvent) {
	switch e.eventType {
	case RegisterCamera:
		s.cameras[e.camera] = struct{}{}
	case DisconnectCamera:
		delete(s.cameras, e.camera)
	case PlateEvent:
		logger.Info("handling plate event")
		s.handlePlate(e.camera, e.plate)
	}
}

func (s *DispatcherService) handleDispatcherEvent(d DispatcherEvent) {
	switch d.eventType {
	case RegisterDispatcher:
		s.dispatchers[d.dispatcher] = d.chanRecord
	case DisconnectDispatcher:
		delete(s.dispatchers, d.dispatcher)
	case TicketEvent:
		logger.Info("sending ticket")
		s.dispatchers[d.dispatcher] <- TicketRecord{}
	}
}

func (s *DispatcherService) handlePlate(c *Camera, p *Plate) {
	s.events[p][c.Road] = append(s.events[p][c.Road], SeenHistory{
		Timestamp: p.Timestamp,
		Camera:    c,
	})
	logger.Info("added event", "timestamp", p.Timestamp, "plate", p.Plate, "road", c.Road, "mile", c.Mile, "limit_mph", c.LimitMPH)
}

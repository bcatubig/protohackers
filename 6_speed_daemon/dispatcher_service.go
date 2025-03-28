package main

import "sync"

type DispatcherService struct {
	mu          sync.Mutex
	dispatchers map[*Dispatcher]struct{}
}

func NewDispatcherService() *DispatcherService {
	s := &DispatcherService{
		mu:          sync.Mutex{},
		dispatchers: make(map[*Dispatcher]struct{}),
	}

	return s
}

func (s *DispatcherService) RegisterDispatcher(d *Dispatcher) {
	s.mu.Lock()
	s.dispatchers[d] = struct{}{}
	s.mu.Unlock()

	logger.Info("registered dispatcher", "dispatcher", d)
}

func (s *DispatcherService) SendTicket() {
}

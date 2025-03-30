package main

type DispatcherService struct{}

func (s *DispatcherService) RegisterDispatcher(c *conn) {
}

func (s *DispatcherService) ReportPlate(road uint16, plate string, timestamp uint32)

package main

import (
	"slices"
	"sync"
)

type Event struct {
	Timestamp  uint32
	Plate      string
	CameraRoad uint16
	CameraMile uint16
}

type DB struct {
	mu   sync.RWMutex
	data []Event
}

func NewDB() *DB {
	return &DB{
		data: make([]Event, 1024),
	}
}

func (d *DB) Insert(e Event) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.data = append(d.data, e)
}

func (d *DB) Filter(test func(e Event) bool) []Event {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := slices.Collect(func(yield func(Event) bool) {
		for _, n := range d.data {
			if test(n) {
				if !yield(n) {
					return
				}
			}
		}
	})

	return result
}

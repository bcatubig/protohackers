package main

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDBInsert(t *testing.T) {
	// ctx := context.Background()

	t.Run("Happy path", func(t *testing.T) {
		d := NewDB()

		wg := &sync.WaitGroup{}
		wg.Add(2)

		go func() {
			defer wg.Done()
			d.Insert(Event{
				Timestamp:  0,
				Plate:      "UN1X",
				CameraRoad: 123,
				CameraMile: 8,
			})
		}()

		go func() {
			defer wg.Done()
			d.Insert(Event{
				Timestamp:  45,
				Plate:      "UN1X",
				CameraRoad: 123,
				CameraMile: 9,
			})
		}()

		wg.Wait()

		got := d.Filter(func(e Event) bool {
			if e.Plate == "UN1X" && e.Timestamp == 45 {
				return true
			}

			return false
		})

		assert.Len(t, got, 2)
		t.Log(got)
	})
}

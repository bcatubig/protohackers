package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDBInsert(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		db := NewDB()

		db.Insert(12345, 101)
		db.Insert(12346, 102)
		db.Insert(12347, 100)
		db.Insert(40960, 5)

		got := db.Mean(12288, 16384)

		assert.Equal(t, 101, got)
	})
}

package main

import (
	"github.com/tidwall/btree"
)

type DB struct {
	db *btree.Map[int32, int32]
}

func NewDB() *DB {
	return &DB{
		db: new(btree.Map[int32, int32]),
	}
}

func (d *DB) Insert(timestamp, cents int32) {
	d.db.Set(timestamp, cents)
}

func (d *DB) Mean(minTime, maxTime int32) int {
	var result int
	var count int
	d.db.Ascend(minTime, func(key, value int32) bool {
		if key >= minTime && key <= maxTime {
			result += int(value)
			count++

			return true
		}

		return false
	})

	if count == 0 {
		return 0
	}

	logger.Info("mean results", "total_items", d.db.Len(), "sum", result, "count", count)

	return result / count
}

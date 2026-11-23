package core

import (
	"errors"
	"math"
)

var ErrSnowballCapacity = errors.New("snowball inventory is full")

func (inventory Inventory) ReceiveSnowball() (Inventory, error) {
	if inventory.Snowballs == math.MaxInt64 {
		return inventory, ErrSnowballCapacity
	}
	inventory.Snowballs++
	return inventory, nil
}

package core

import (
	"errors"
	"math"
)

type SnowballType string

const (
	NormalSnowball SnowballType = "normal"
	YellowSnowball SnowballType = "yellow"
)

var (
	ErrSnowballCapacity           = errors.New("snowball inventory is full")
	ErrSnowballType               = errors.New("unknown snowball type")
	ErrInsufficientYellowSnowball = errors.New("no yellow snowballs available")
)

func (inventory Inventory) ReceiveSnowball() (Inventory, error) {
	if inventory.Snowballs == math.MaxInt64 {
		return inventory, ErrSnowballCapacity
	}
	inventory.Snowballs++
	return inventory, nil
}

func TransferYellowSnowball(sender, recipient GiftRecord) (GiftRecord, GiftRecord, error) {
	if sender.Inventory.YellowSnowballs == 0 {
		return sender, recipient, ErrInsufficientYellowSnowball
	}
	if recipient.Inventory.YellowSnowballs == math.MaxInt64 {
		return sender, recipient, ErrSnowballCapacity
	}
	sender.Inventory.YellowSnowballs--
	recipient.Inventory.YellowSnowballs++
	sender.Aura.Bad = increaseAura(sender.Aura.Bad, 1, 5)
	return sender, recipient, nil
}

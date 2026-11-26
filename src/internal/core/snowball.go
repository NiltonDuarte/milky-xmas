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
	ErrInsufficientSnowball       = errors.New("no snowballs available")
	ErrInsufficientYellowSnowball = errors.New("no yellow snowballs available")
)

// TransferSnowball moves one normal Snowball from the sender to the recipient.
func TransferSnowball(sender, recipient GiftRecord) (GiftRecord, GiftRecord, error) {
	if sender.Inventory.Snowballs == 0 {
		return sender, recipient, ErrInsufficientSnowball
	}
	if recipient.Inventory.Snowballs == math.MaxInt64 {
		return sender, recipient, ErrSnowballCapacity
	}
	sender.Inventory.Snowballs--
	recipient.Inventory.Snowballs++
	return sender, recipient, nil
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
	sender.Aura = sender.Aura.Move(1, -YellowSnowballPoints)
	return sender, recipient, nil
}

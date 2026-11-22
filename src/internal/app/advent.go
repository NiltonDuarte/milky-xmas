package app

import (
	"context"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

type AdventStore interface {
	AdventDoors(context.Context, Member, int) (uint32, error)
	// OpenAdventDoor records an opening once and returns the resulting yearly total.
	OpenAdventDoor(context.Context, Member, int, int) (bool, int, error)
}

type Store interface {
	GiftStore
	AdventStore
}

type AdventOpening struct {
	Status  core.DoorStatus
	Awarded bool
	Total   int
}

func (service Service) ViewAdvent(ctx context.Context, member Member) (core.AdventCalendar, error) {
	now := service.now().UTC()
	opened, err := service.advent.AdventDoors(ctx, member, now.Year())
	if err != nil {
		return core.AdventCalendar{}, err
	}
	return core.AdventView(now, opened), nil
}

func (service Service) OpenAdvent(ctx context.Context, member Member, day int) (AdventOpening, error) {
	now := service.now().UTC()
	status := core.AdventStatus(now, day, false)
	if status != core.DoorAvailable {
		return AdventOpening{Status: status}, nil
	}
	awarded, total, err := service.advent.OpenAdventDoor(ctx, member, now.Year(), day)
	if err != nil {
		return AdventOpening{}, err
	}
	return AdventOpening{Status: core.DoorOpened, Awarded: awarded, Total: total}, nil
}

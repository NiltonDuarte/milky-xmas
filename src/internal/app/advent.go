package app

import (
	"context"
	"fmt"

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
	AuraStore
	StatsStore
}

type AdventOpening struct {
	Status core.DoorStatus
	// Item is the surprise behind an opened door, from 0 to core.AdventDays-1.
	Item int
}

// adventOrder is the player's own shuffle for one purpose, such as button places or contents.
func adventOrder(purpose string, member Member, year int) [core.AdventDays]int {
	return core.AdventOrder(fmt.Sprintf("%s|%s|%s|%d", purpose, member.ServerID, member.UserID, year))
}

func (service Service) ViewAdvent(ctx context.Context, member Member) (core.AdventCalendar, error) {
	now := service.now().UTC()
	year := core.AdventYearKey(now)
	opened, err := service.advent.AdventDoors(ctx, member, year)
	if err != nil {
		return core.AdventCalendar{}, err
	}
	return core.AdventView(now, opened, adventOrder("buttons", member, year)), nil
}

// OpenAdvent opens a door that is available, or shows again one that is already open.
func (service Service) OpenAdvent(ctx context.Context, member Member, day int) (AdventOpening, error) {
	if day < 1 || day > core.AdventDays {
		return AdventOpening{Status: core.DoorInvalid}, nil
	}
	now := service.now().UTC()
	year := core.AdventYearKey(now)
	opened, err := service.advent.AdventDoors(ctx, member, year)
	if err != nil {
		return AdventOpening{}, err
	}
	if opened&(1<<(day-1)) == 0 {
		if status := core.AdventStatus(now, day, false); status != core.DoorAvailable {
			return AdventOpening{Status: status}, nil
		}
		if _, _, err := service.advent.OpenAdventDoor(ctx, member, year, day); err != nil {
			return AdventOpening{}, err
		}
	}
	return AdventOpening{Status: core.DoorOpened, Item: adventOrder("content", member, year)[day-1]}, nil
}

package app

import (
	"errors"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

const HelloCommand = "hello"

var ErrUnknownCommand = errors.New("unknown command")

type Service struct {
	gifts  GiftStore
	advent AdventStore
	now    func() time.Time
	random func(int) int
	tuning core.Tuning
}

// Tuning returns the adjustable numbers this service applies.
func (service Service) Tuning() core.Tuning {
	return service.tuning
}

// WithTuning returns a service that applies the given numbers instead of the defaults.
func (service Service) WithTuning(tuning core.Tuning) Service {
	service.tuning = tuning
	return service
}

func (Service) Reply(command string) (string, error) {
	if command != HelloCommand {
		return "", ErrUnknownCommand
	}
	return core.Greeting(), nil
}

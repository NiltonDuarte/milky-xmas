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
}

func (Service) Reply(command string) (string, error) {
	if command != HelloCommand {
		return "", ErrUnknownCommand
	}
	return core.Greeting(), nil
}

package app

import (
	"errors"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

const HelloCommand = "hello"

var ErrUnknownCommand = errors.New("unknown command")

type Service struct{}

func (Service) Reply(command string) (string, error) {
	if command != HelloCommand {
		return "", ErrUnknownCommand
	}
	return core.Greeting(), nil
}

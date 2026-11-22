package app

import (
	"errors"
	"testing"
)

func TestReply(t *testing.T) {
	for _, tt := range []struct {
		command string
		reply   string
		err     error
	}{
		{command: "hello", reply: "Hello, world!"},
		{command: "hello world", err: ErrUnknownCommand},
		{command: "unknown", err: ErrUnknownCommand},
		{command: "", err: ErrUnknownCommand},
	} {
		t.Run(tt.command, func(t *testing.T) {
			got, err := (Service{}).Reply(tt.command)
			if got != tt.reply || !errors.Is(err, tt.err) {
				t.Errorf("Reply(%q) = %q, %v; want %q, %v", tt.command, got, err, tt.reply, tt.err)
			}
		})
	}
}

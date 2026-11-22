package core

import "testing"

func TestGreeting(t *testing.T) {
	if got := Greeting(); got != "Hello, world!" {
		t.Errorf("Greeting() = %q", got)
	}
}

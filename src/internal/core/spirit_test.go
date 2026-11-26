package core

import "testing"

func TestServerSpirit(t *testing.T) {
	if got := ServerSpirit([]Aura{50, -5, 0, 10, -50}); got != (Spirit{Good: 60, Bad: 55}) {
		t.Errorf("spirit = %+v", got)
	}
	if got := ServerSpirit(nil); got != (Spirit{}) {
		t.Errorf("spirit of an empty server = %+v", got)
	}
}

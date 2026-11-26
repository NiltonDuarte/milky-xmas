package core

// Spirit is the server's Christmas spirit: every player's aura added up by side.
type Spirit struct {
	Good, Bad int
}

// ServerSpirit adds up the good side and the bad side of the players' auras.
func ServerSpirit(auras []Aura) Spirit {
	var spirit Spirit
	for _, aura := range auras {
		if aura > 0 {
			spirit.Good += int(aura)
		} else {
			spirit.Bad -= int(aura)
		}
	}
	return spirit
}

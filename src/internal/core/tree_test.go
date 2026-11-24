package core

import (
	"errors"
	"testing"
	"time"
)

var treeStart = time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)

func TestTreeStartsDryAndDecaysByTheHour(t *testing.T) {
	var tree Tree
	if got := tree.At(treeStart); got.Level != TreeStartLevel || TreeStage(got.Level) != 0 {
		t.Fatalf("new tree = %+v", got)
	}
	watered, err := tree.Water(treeStart, "alice")
	if err != nil || watered.Level != TreeStartLevel+TreeWaterPoints || watered.Waters["alice"] != 1 {
		t.Fatalf("first watering = %+v, %v", watered, err)
	}
	for _, tt := range []struct {
		after time.Duration
		want  int
	}{
		{59 * time.Minute, 12}, {time.Hour, 11}, {5*time.Hour + 30*time.Minute, 7}, {12 * time.Hour, 0}, {100 * time.Hour, 0},
	} {
		if got := watered.At(treeStart.Add(tt.after)).Level; got != tt.want {
			t.Errorf("level after %s = %d, want %d", tt.after, got, tt.want)
		}
	}
	// Decay keeps the part of an hour that has not passed yet.
	later := watered.At(treeStart.Add(90 * time.Minute))
	if again := later.At(treeStart.Add(2 * time.Hour)); again.Level != 10 {
		t.Errorf("level after 2h through two reads = %d, want 10", again.Level)
	}
}

func TestTreeWateringRules(t *testing.T) {
	tree, err := Tree{Level: 98, Since: treeStart}.Water(treeStart, "alice")
	if err != nil || tree.Level != 100 {
		t.Fatalf("watering to the cap = %+v, %v", tree, err)
	}
	if _, err := tree.Water(treeStart.Add(time.Minute), "bob"); !errors.Is(err, ErrTreeFull) {
		t.Fatalf("watering a full tree = %v", err)
	}
	tree = tree.At(treeStart.Add(3 * time.Hour)) // 97%
	if tree.Level != 97 {
		t.Fatalf("level = %d", tree.Level)
	}
	_, err = tree.Water(treeStart.Add(3*time.Hour), "alice") // watered 3 hours ago, so allowed
	if err != nil {
		t.Fatalf("watering after the cooldown = %v", err)
	}
	recent, err := Tree{Level: 50, Since: treeStart}.Water(treeStart, "alice")
	if err != nil {
		t.Fatal(err)
	}
	_, err = recent.Water(treeStart.Add(4*time.Minute), "alice")
	var cooldown TreeCooldownError
	if !errors.Is(err, ErrTreeCooldown) || !errors.As(err, &cooldown) || cooldown.Remaining != 6*time.Minute {
		t.Fatalf("early watering = %v", err)
	}
	if _, err := recent.Water(treeStart.Add(4*time.Minute), "bob"); err != nil {
		t.Fatalf("another player's watering = %v", err)
	}
	if _, err := recent.Water(treeStart.Add(TreeWaterCooldown), "alice"); err != nil {
		t.Fatalf("watering at the cooldown boundary = %v", err)
	}
	if recent.Waterings["alice"] != treeStart || len(recent.Waters) != 1 {
		t.Fatalf("watering changed the original tree: %+v", recent)
	}
}

func TestTreeStagesAndTopWaterers(t *testing.T) {
	for level, want := range map[int]int{0: 0, 25: 0, 26: 1, 50: 1, 51: 2, 75: 2, 76: 3, 99: 3, 100: 4} {
		if got := TreeStage(level); got != want {
			t.Errorf("stage at %d%% = %d, want %d", level, got, want)
		}
	}
	tree := Tree{Waters: map[string]int{"b": 3, "a": 3, "c": 5, "d": 1}}
	top := tree.TopWaterers(3)
	if len(top) != 3 || top[0].Member != "c" || top[1].Member != "a" || top[2].Member != "b" {
		t.Fatalf("top waterers = %+v", top)
	}
	if len((Tree{}).TopWaterers(3)) != 0 {
		t.Error("an unwatered tree has no top waterers")
	}
}

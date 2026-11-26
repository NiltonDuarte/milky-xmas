package tuningfile

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func TestShippedFileEqualsTheDefaults(t *testing.T) {
	got, err := Load(filepath.Join("..", "..", "..", "..", "tuning.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if want := core.DefaultTuning(); !reflect.DeepEqual(got, want) {
		t.Fatalf("tuning.yaml = %+v\nwant defaults %+v", got, want)
	}
}

func TestParseEmptyAndCommentOnlyFilesUseDefaults(t *testing.T) {
	for name, text := range map[string]string{"empty": "", "comments": "# nothing here\n"} {
		got, err := Parse([]byte(text))
		if err != nil || !reflect.DeepEqual(got, core.DefaultTuning()) {
			t.Errorf("%s file = %+v, %v", name, got, err)
		}
	}
}

func TestParsePartialFileKeepsOtherDefaults(t *testing.T) {
	got, err := Parse([]byte("drops:\n  gift_rain:\n    claims_to_end: 3\ngrinch:\n  attack:\n    heavy_damage: 6.5\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := core.DefaultTuning()
	want.Drops.GiftRain.ClaimsToEnd = 3
	want.Grinch.AttackHeavyDamage = 650
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("tuning = %+v\nwant %+v", got, want)
	}
}

func TestParseConvertsUnits(t *testing.T) {
	got, err := Parse([]byte(`
drops:
  lifetime: 90s
grinch:
  fight_duration: 2m30s
  snowballs: {damage: 0.5, boss_hp_after: 90}
  defense: {boss_hp_after: 60.25}
  final_attack: {hit_score: 3}
`))
	if err != nil {
		t.Fatal(err)
	}
	g := got.Grinch
	if got.Drops.Lifetime != 90*time.Second || g.FightDuration != 150*time.Second ||
		g.SnowballDamage != 50 || g.SnowballsBossHP != 9000 || g.DefenseBossHP != 6025 || g.FinalHitScore != 300 {
		t.Fatalf("converted tuning = %+v", got)
	}
}

func TestParseRejectsBadFiles(t *testing.T) {
	for _, tt := range []struct {
		name string
		text string
		want []string
	}{
		{"unknown key", "drops:\n  gift_rain:\n    claims_to_edn: 3\n", []string{"claims_to_edn"}},
		{"unknown section", "dorps: {}\n", []string{"dorps"}},
		{"odds do not add up", "gifts:\n  daily_rarity: {normal: 70, rare: 25, epic: 10}\n", []string{"gifts.daily_rarity", "add up to 100"}},
		{"bare number as duration", "drops:\n  lifetime: 10\n", []string{"duration", "unit"}},
		{"bad duration", "drops:\n  lifetime: soon\n", []string{"soon"}},
		{"text as number", "drops:\n  gift_rain:\n    claims_to_end: many\n", []string{"many"}},
		{"decimal where whole expected", "drops:\n  gift_rain:\n    claims_to_end: 2.5\n", []string{"2.5"}},
		{"three decimals", "grinch:\n  attack:\n    heavy_damage: 5.255\n", []string{"grinch.attack.heavy_damage", "two decimals"}},
		{"milestones out of order", "grinch:\n  defense:\n    boss_hp_after: 80\n", []string{"boss HP milestones must fall from 100 toward 0"}},
		{"tree decay interval as a bare number", "tree:\n  decay_interval: 60\n", []string{"duration", "unit"}},
		{"tree event end that is not a date", "tree:\n  event_end: someday\n", []string{"someday", "date and time"}},
		{"tree water points out of range", "tree:\n  water_points: 0\n", []string{"tree.water_points"}},
		{"several problems", "drops:\n  lifetime: 0s\ngrinch:\n  fight_duration: 0s\n", []string{"drops.lifetime", "grinch.fight_duration"}},
		{"unknown key and bad range together", "bogus: 1\ndrops:\n  lifetime: 0s\n", []string{"bogus", "drops.lifetime"}},
		{"not a mapping", "- one\n- two\n", []string{"cannot unmarshal"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse([]byte(tt.text))
			if err == nil {
				t.Fatal("Parse() = nil error")
			}
			for _, want := range tt.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q does not mention %q", err, want)
				}
			}
		})
	}
}

func TestLoadAndFromEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tuning.yaml")
	if err := os.WriteFile(path, []byte("drops:\n  gift_rain:\n    claims_to_end: 3\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv(EnvVar, "")
	if got, err := FromEnv(); err != nil || !reflect.DeepEqual(got, core.DefaultTuning()) {
		t.Fatalf("unset FromEnv = %+v, %v", got, err)
	}
	t.Setenv(EnvVar, path)
	got, err := FromEnv()
	if err != nil || got.Drops.GiftRain.ClaimsToEnd != 3 {
		t.Fatalf("FromEnv = %+v, %v", got, err)
	}

	missing := filepath.Join(dir, "missing.yaml")
	t.Setenv(EnvVar, missing)
	if _, err := FromEnv(); err == nil || !strings.Contains(err.Error(), missing) {
		t.Fatalf("missing file error = %v, want it to name %s", err, missing)
	}

	bad := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(bad, []byte("drops:\n  lifetime: 0s\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(bad); err == nil || !strings.Contains(err.Error(), bad) || !strings.Contains(err.Error(), "drops.lifetime") {
		t.Fatalf("invalid file error = %v, want the file and the key path", err)
	}
}

func TestParseTreeNumbers(t *testing.T) {
	for name, end := range map[string]string{"plain": "2026-12-25T18:30:00-03:00", "quoted": `"2026-12-25T21:30:00Z"`} {
		got, err := Parse([]byte("tree:\n  start_level: 30\n  water_points: 4\n  water_aura: 2\n  decay_interval: 30m\n  water_cooldown: 0s\n  event_end: " + end + "\n"))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		want := core.TreeRules{
			StartLevel: 30, WaterPoints: 4, WaterAura: 2, DecayInterval: 30 * time.Minute, WaterCooldown: 0,
			EventEnd: time.Date(2026, 12, 25, 21, 30, 0, 0, time.UTC),
		}
		if !reflect.DeepEqual(got.Tree, want) {
			t.Errorf("%s: tree = %+v, want %+v", name, got.Tree, want)
		}
	}
}

func TestParseRegistrationStartWindow(t *testing.T) {
	got, err := Parse([]byte("grinch:\n  registration:\n    start_window: 5m\n    duration: 90s\n"))
	if err != nil {
		t.Fatal(err)
	}
	if got.Grinch.StartWindow != 5*time.Minute || got.Grinch.RegistrationDuration != 90*time.Second || got.Grinch.EarlyStartPlayers != 3 {
		t.Fatalf("registration = %+v", got.Grinch)
	}
	if _, err := Parse([]byte("grinch:\n  registration:\n    start_window: 0s\n")); err == nil || !strings.Contains(err.Error(), "grinch.registration.start_window") {
		t.Fatalf("zero start window error = %v", err)
	}
}

func TestParseSnowballsDuration(t *testing.T) {
	got, err := Parse([]byte("grinch:\n  snowballs:\n    duration: 3m\n"))
	if err != nil || got.Grinch.SnowballsDuration != 3*time.Minute || got.Grinch.SnowballsBossHP != 7500 {
		t.Fatalf("snowballs = %+v, %v", got.Grinch, err)
	}
	if _, err := Parse([]byte("grinch:\n  snowballs:\n    duration: 0s\n")); err == nil || !strings.Contains(err.Error(), "grinch.snowballs.duration") {
		t.Fatalf("zero duration error = %v", err)
	}
}

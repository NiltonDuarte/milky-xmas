// Package tuningfile reads the YAML file that adjusts the documented balance numbers.
//
// Keys left out of the file keep their default. Durations are strings such as "10m".
// Boss HP and damage are percentages of the Grinch's health with up to two decimals.
package tuningfile

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"go.yaml.in/yaml/v3"
)

// EnvVar names the environment variable that points at the tuning file.
const EnvVar = "MILKY_TUNING_CONFIG"

// FromEnv loads the file named by MILKY_TUNING_CONFIG, or returns the defaults when it is unset.
func FromEnv() (core.Tuning, error) {
	path := strings.TrimSpace(os.Getenv(EnvVar))
	if path == "" {
		return core.DefaultTuning(), nil
	}
	log.Printf("Using tuning file %s", path)
	return Load(path)
}

// Load reads and validates the tuning file at path.
func Load(path string) (core.Tuning, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return core.Tuning{}, fmt.Errorf("read tuning file: %w", err)
	}
	tuning, err := Parse(data)
	if err != nil {
		return core.Tuning{}, fmt.Errorf("tuning file %s: %w", path, err)
	}
	return tuning, nil
}

// Parse decodes YAML over the defaults and validates the result. Unknown keys and range
// problems are reported together.
func Parse(data []byte) (core.Tuning, error) {
	doc := defaultsDocument()
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	var problems []error
	if err := decoder.Decode(&doc); err != nil && !errors.Is(err, io.EOF) {
		var typeErrors *yaml.TypeError
		if !errors.As(err, &typeErrors) {
			return core.Tuning{}, err
		}
		problems = append(problems, err)
	}
	tuning, err := doc.tuning()
	if err != nil {
		return core.Tuning{}, errors.Join(append(problems, err)...)
	}
	if err := tuning.Validate(); err != nil {
		problems = append(problems, err)
	}
	if len(problems) > 0 {
		return core.Tuning{}, errors.Join(problems...)
	}
	return tuning, nil
}

// duration accepts only strings like "90s" or "10m". A bare number is ambiguous.
type duration time.Duration

func (d *duration) UnmarshalYAML(node *yaml.Node) error {
	var text string
	if node.Kind != yaml.ScalarNode || node.Tag != "!!str" {
		return fmt.Errorf("line %d: a duration must be a string with a unit, such as 90s or 10m", node.Line)
	}
	if err := node.Decode(&text); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(text)
	if err != nil {
		return fmt.Errorf("line %d: %q is not a duration such as 90s or 10m", node.Line, text)
	}
	*d = duration(parsed)
	return nil
}

// timestamp accepts a date and time such as 2027-01-05T01:00:00Z.
type timestamp time.Time

func (ts *timestamp) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.ScalarNode || (node.Tag != "!!timestamp" && node.Tag != "!!str") {
		return fmt.Errorf("line %d: a date and time must look like 2027-01-05T01:00:00Z", node.Line)
	}
	parsed, err := time.Parse(time.RFC3339, node.Value)
	if err != nil {
		return fmt.Errorf("line %d: %q is not a date and time such as 2027-01-05T01:00:00Z", node.Line, node.Value)
	}
	*ts = timestamp(parsed.UTC())
	return nil
}

// whole accepts only integers. The YAML decoder would otherwise cut 2.5 down to 2.
type whole int64

func (w *whole) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.ScalarNode || node.Tag != "!!int" {
		return fmt.Errorf("line %d: %q must be a whole number", node.Line, node.Value)
	}
	var value int64
	if err := node.Decode(&value); err != nil {
		return err
	}
	*w = whole(value)
	return nil
}

type odds struct {
	Normal whole `yaml:"normal"`
	Rare   whole `yaml:"rare"`
	Epic   whole `yaml:"epic"`
}

type document struct {
	Gifts struct {
		DailyRarity    odds `yaml:"daily_rarity"`
		GoodDeedsBoost struct {
			Threshold whole `yaml:"threshold"`
			Rarity    odds  `yaml:"rarity"`
		} `yaml:"good_deeds_boost"`
	} `yaml:"gifts"`
	Drops struct {
		Lifetime     duration `yaml:"lifetime"`
		RandomActive struct {
			OneIn whole `yaml:"one_in"`
		} `yaml:"random_active"`
		GiftRain struct {
			MinMessages   whole    `yaml:"min_messages"`
			MaxMessages   whole    `yaml:"max_messages"`
			ClaimsToEnd   whole    `yaml:"claims_to_end"`
			ClaimCooldown duration `yaml:"claim_cooldown"`
			CoalPercent   whole    `yaml:"coal_percent"`
		} `yaml:"gift_rain"`
	} `yaml:"drops"`
	Grinch struct {
		Spawn struct {
			CheckInterval      duration `yaml:"check_interval"`
			ChancePercent      whole    `yaml:"chance_percent"`
			ActivityWindow     duration `yaml:"activity_window"`
			CooldownAfterFight duration `yaml:"cooldown_after_fight"`
		} `yaml:"spawn"`
		Registration struct {
			StartWindow       duration `yaml:"start_window"`
			Duration          duration `yaml:"duration"`
			EarlyStartPlayers whole    `yaml:"early_start_players"`
		} `yaml:"registration"`
		FightDuration duration `yaml:"fight_duration"`
		Snowballs     struct {
			Duration    duration `yaml:"duration"`
			Damage      float64  `yaml:"damage"`
			BossHPAfter float64  `yaml:"boss_hp_after"`
		} `yaml:"snowballs"`
		Defense struct {
			BossHPAfter       float64 `yaml:"boss_hp_after"`
			MissHPLossPercent whole   `yaml:"miss_hp_loss_percent"`
			MaxDefenses       whole   `yaml:"max_defenses"`
		} `yaml:"defense"`
		Attack struct {
			Damage      float64 `yaml:"damage"`
			HeavyDamage float64 `yaml:"heavy_damage"`
			BossHPAfter float64 `yaml:"boss_hp_after"`
		} `yaml:"attack"`
		FinalAttack struct {
			HitDamage  float64 `yaml:"hit_damage"`
			HitScore   float64 `yaml:"hit_score"`
			GiftPoints struct {
				Min whole `yaml:"min"`
				Max whole `yaml:"max"`
			} `yaml:"gift_points"`
			SnowballsReward whole `yaml:"snowballs_reward"`
		} `yaml:"final_attack"`
		LossRewards struct {
			Coal                          whole `yaml:"coal"`
			YellowSnowballsPerGroupMember whole `yaml:"yellow_snowballs_per_group_member"`
		} `yaml:"loss_rewards"`
	} `yaml:"grinch"`
	Tree struct {
		StartLevel    whole     `yaml:"start_level"`
		WaterPoints   whole     `yaml:"water_points"`
		WaterAura     whole     `yaml:"water_aura"`
		DecayInterval duration  `yaml:"decay_interval"`
		WaterCooldown duration  `yaml:"water_cooldown"`
		EventEnd      timestamp `yaml:"event_end"`
	} `yaml:"tree"`
}

// percent writes hundredths of HP as a percentage, for the file.
func percent(hundredths int) float64 { return float64(hundredths) / 100 }

func oddsDocument(o core.RarityOdds) odds {
	return odds{Normal: whole(o.Normal), Rare: whole(o.Rare), Epic: whole(o.Epic)}
}

func (o odds) core() core.RarityOdds {
	return core.RarityOdds{Normal: int(o.Normal), Rare: int(o.Rare), Epic: int(o.Epic)}
}

func defaultsDocument() document {
	d := core.DefaultTuning()
	var doc document
	doc.Gifts.DailyRarity = oddsDocument(d.Gifts.DailyRarity)
	doc.Gifts.GoodDeedsBoost.Threshold = whole(d.Gifts.GoodDeedsThreshold)
	doc.Gifts.GoodDeedsBoost.Rarity = oddsDocument(d.Gifts.BoostedRarity)
	doc.Drops.Lifetime = duration(d.Drops.Lifetime)
	doc.Drops.RandomActive.OneIn = whole(d.Drops.RandomActiveOneIn)
	rain := d.Drops.GiftRain
	doc.Drops.GiftRain.MinMessages, doc.Drops.GiftRain.MaxMessages = whole(rain.MinMessages), whole(rain.MaxMessages)
	doc.Drops.GiftRain.ClaimsToEnd, doc.Drops.GiftRain.ClaimCooldown = whole(rain.ClaimsToEnd), duration(rain.ClaimCooldown)
	doc.Drops.GiftRain.CoalPercent = whole(rain.CoalPercent)
	g := d.Grinch
	doc.Grinch.Spawn.CheckInterval, doc.Grinch.Spawn.ChancePercent = duration(g.SpawnCheckInterval), whole(g.SpawnChancePercent)
	doc.Grinch.Spawn.ActivityWindow, doc.Grinch.Spawn.CooldownAfterFight = duration(g.ActivityWindow), duration(g.CooldownAfterFight)
	doc.Grinch.Registration.StartWindow = duration(g.StartWindow)
	doc.Grinch.Registration.Duration, doc.Grinch.Registration.EarlyStartPlayers = duration(g.RegistrationDuration), whole(g.EarlyStartPlayers)
	doc.Grinch.FightDuration = duration(g.FightDuration)
	doc.Grinch.Snowballs.Duration = duration(g.SnowballsDuration)
	doc.Grinch.Snowballs.Damage, doc.Grinch.Snowballs.BossHPAfter = percent(g.SnowballDamage), percent(g.SnowballsBossHP)
	doc.Grinch.Defense.BossHPAfter, doc.Grinch.Defense.MissHPLossPercent = percent(g.DefenseBossHP), whole(g.DefenseMissHPLoss)
	doc.Grinch.Defense.MaxDefenses = whole(g.DefenseMaxDefenses)
	doc.Grinch.Attack.Damage, doc.Grinch.Attack.HeavyDamage = percent(g.AttackDamage), percent(g.AttackHeavyDamage)
	doc.Grinch.Attack.BossHPAfter = percent(g.AttackBossHP)
	doc.Grinch.FinalAttack.HitDamage, doc.Grinch.FinalAttack.HitScore = percent(g.FinalHitDamage), percent(g.FinalHitScore)
	doc.Grinch.FinalAttack.GiftPoints.Min, doc.Grinch.FinalAttack.GiftPoints.Max = whole(g.GiftPointsMin), whole(g.GiftPointsMax)
	doc.Grinch.FinalAttack.SnowballsReward = whole(g.FinalSnowballReward)
	doc.Grinch.LossRewards.Coal, doc.Grinch.LossRewards.YellowSnowballsPerGroupMember = whole(g.LossCoal), whole(g.LossYellowPerGroupMember)
	tree := d.Tree
	doc.Tree.StartLevel, doc.Tree.WaterPoints = whole(tree.StartLevel), whole(tree.WaterPoints)
	doc.Tree.DecayInterval, doc.Tree.WaterCooldown = duration(tree.DecayInterval), duration(tree.WaterCooldown)
	doc.Tree.EventEnd = timestamp(tree.EventEnd)
	doc.Tree.WaterAura = whole(tree.WaterAura)
	return doc
}

// tuning converts the file's units to the core's. It reports every number that cannot
// be converted; range checks belong to core.Tuning.Validate.
func (doc document) tuning() (core.Tuning, error) {
	var problems []error
	hundredths := func(path string, value float64) int {
		scaled := value * 100
		if math.IsNaN(value) || math.Abs(value) > 1e6 || math.Abs(scaled-math.Round(scaled)) > 1e-6 {
			problems = append(problems, fmt.Errorf("%s: %v must be a number with at most two decimals", path, value))
			return 0
		}
		return int(math.Round(scaled))
	}
	var t core.Tuning
	t.Gifts.DailyRarity = doc.Gifts.DailyRarity.core()
	t.Gifts.GoodDeedsThreshold = int(doc.Gifts.GoodDeedsBoost.Threshold)
	t.Gifts.BoostedRarity = doc.Gifts.GoodDeedsBoost.Rarity.core()
	t.Drops.Lifetime = time.Duration(doc.Drops.Lifetime)
	t.Drops.RandomActiveOneIn = int(doc.Drops.RandomActive.OneIn)
	rain := doc.Drops.GiftRain
	t.Drops.GiftRain = core.GiftRainTuning{
		MinMessages: int(rain.MinMessages), MaxMessages: int(rain.MaxMessages), ClaimsToEnd: int(rain.ClaimsToEnd),
		ClaimCooldown: time.Duration(rain.ClaimCooldown), CoalPercent: int(rain.CoalPercent),
	}
	g := doc.Grinch
	t.Grinch = core.GrinchRules{
		SpawnCheckInterval:       time.Duration(g.Spawn.CheckInterval),
		SpawnChancePercent:       int(g.Spawn.ChancePercent),
		ActivityWindow:           time.Duration(g.Spawn.ActivityWindow),
		CooldownAfterFight:       time.Duration(g.Spawn.CooldownAfterFight),
		StartWindow:              time.Duration(g.Registration.StartWindow),
		RegistrationDuration:     time.Duration(g.Registration.Duration),
		EarlyStartPlayers:        int(g.Registration.EarlyStartPlayers),
		FightDuration:            time.Duration(g.FightDuration),
		SnowballsDuration:        time.Duration(g.Snowballs.Duration),
		SnowballDamage:           hundredths("grinch.snowballs.damage", g.Snowballs.Damage),
		SnowballsBossHP:          hundredths("grinch.snowballs.boss_hp_after", g.Snowballs.BossHPAfter),
		DefenseBossHP:            hundredths("grinch.defense.boss_hp_after", g.Defense.BossHPAfter),
		DefenseMissHPLoss:        int(g.Defense.MissHPLossPercent),
		DefenseMaxDefenses:       int(g.Defense.MaxDefenses),
		AttackDamage:             hundredths("grinch.attack.damage", g.Attack.Damage),
		AttackHeavyDamage:        hundredths("grinch.attack.heavy_damage", g.Attack.HeavyDamage),
		AttackBossHP:             hundredths("grinch.attack.boss_hp_after", g.Attack.BossHPAfter),
		FinalHitDamage:           hundredths("grinch.final_attack.hit_damage", g.FinalAttack.HitDamage),
		FinalHitScore:            hundredths("grinch.final_attack.hit_score", g.FinalAttack.HitScore),
		GiftPointsMin:            int(g.FinalAttack.GiftPoints.Min),
		GiftPointsMax:            int(g.FinalAttack.GiftPoints.Max),
		FinalSnowballReward:      int64(g.FinalAttack.SnowballsReward),
		LossCoal:                 int64(g.LossRewards.Coal),
		LossYellowPerGroupMember: int64(g.LossRewards.YellowSnowballsPerGroupMember),
	}
	t.Tree = core.TreeRules{
		StartLevel: int(doc.Tree.StartLevel), WaterPoints: int(doc.Tree.WaterPoints),
		DecayInterval: time.Duration(doc.Tree.DecayInterval), WaterCooldown: time.Duration(doc.Tree.WaterCooldown),
		EventEnd: time.Time(doc.Tree.EventEnd), WaterAura: int(doc.Tree.WaterAura),
	}
	return t, errors.Join(problems...)
}

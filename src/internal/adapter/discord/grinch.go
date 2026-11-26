package discord

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/grinchtext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
)

type grinchBot struct {
	guild        string
	channel      string
	encounter    *app.GrinchEncounter
	lastActivity atomic.Int64
	shown        shownPicture
}

func grinchCommand() *discordgo.ApplicationCommand {
	permission := int64(discordgo.PermissionManageMessages)
	return &discordgo.ApplicationCommand{
		Type: discordgo.ChatApplicationCommand, Name: "grinch", Description: "Manage the Grinch fight",
		DefaultMemberPermissions: &permission,
		Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionSubCommand, Name: "summon", Description: "Summon the Grinch in the encounter channel"},
		},
	}
}

func (b bot) summonGrinch(s *discordgo.Session, i *discordgo.Interaction, options []*discordgo.ApplicationCommandInteractionDataOption) {
	if len(options) != 1 || options[0] == nil || options[0].Name != "summon" ||
		options[0].Type != discordgo.ApplicationCommandOptionSubCommand {
		return
	}
	respondPrivately(s, i, func(ctx context.Context, _ app.Member) string {
		if i.Member.Permissions&(discordgo.PermissionManageMessages|discordgo.PermissionAdministrator) == 0 {
			return "You need Manage Messages permission to summon the Grinch."
		}
		grinch := b.grinches.get(i.GuildID, i.ChannelID)
		if grinch == nil {
			return "Use this command in a configured Grinch channel."
		}
		display := grinchDisplay{s, grinch.channel, grinch.encounter.Rules(), &grinch.shown}
		if _, err := grinch.encounter.Sync(ctx, display); err != nil {
			logInteractionError(i.ID, "reconcile Grinch encounter", err)
		}
		spawned, err := grinch.encounter.Spawn(ctx, time.Time{}, true, display)
		if err != nil {
			logInteractionError(i.ID, "summon the Grinch", err)
			return "Could not confirm the Grinch announcement."
		}
		if !spawned {
			return "A Grinch encounter is already in progress."
		}
		return "The Grinch is coming! Players can register with 🎅."
	})
}

func (b *grinchBot) observe(_ *discordgo.Session, event *discordgo.MessageCreate) {
	if event.GuildID != b.guild || event.ChannelID != b.channel || event.Author == nil ||
		event.Author.Bot || event.WebhookID != "" {
		return
	}
	when := event.Timestamp
	if when.IsZero() {
		when = time.Now()
	}
	next := when.UnixNano()
	for previous := b.lastActivity.Load(); next > previous; previous = b.lastActivity.Load() {
		if b.lastActivity.CompareAndSwap(previous, next) {
			return
		}
	}
}

func (b *grinchBot) run(ctx context.Context, s *discordgo.Session) {
	refresh := time.NewTicker(time.Second)
	defer refresh.Stop()
	roll := time.NewTicker(b.encounter.Rules().SpawnCheckInterval)
	defer roll.Stop()
	display := grinchDisplay{s, b.channel, b.encounter.Rules(), &b.shown}
	initial, cancel := context.WithTimeout(ctx, 3*time.Second)
	_, err := b.encounter.RefreshBoard(initial, display)
	cancel()
	if err != nil && ctx.Err() == nil {
		log.Printf("Restore Grinch board: %v", err)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-refresh.C:
			attempt, cancel := context.WithTimeout(ctx, 3*time.Second)
			_, err := b.encounter.Sync(attempt, display)
			cancel()
			if err != nil && ctx.Err() == nil {
				log.Printf("Update Grinch fight: %v", err)
			}
		case <-roll.C:
			attempt, cancel := context.WithTimeout(ctx, 4*time.Second)
			_, err := b.encounter.Spawn(attempt, time.Unix(0, b.lastActivity.Load()), false, display)
			cancel()
			if err != nil && ctx.Err() == nil {
				log.Printf("Announce Grinch fight: %v", err)
			}
		}
	}
}

func (b *grinchBot) button(s *discordgo.Session, i *discordgo.Interaction) {
	data, ok := i.Data.(discordgo.MessageComponentInteractionData)
	if !ok || i.ChannelID != b.channel || i.Message == nil {
		return
	}
	parts := strings.Split(data.CustomID, ":")
	if (len(parts) != 4 && len(parts) != 5) || parts[0] != "grinch" {
		return
	}
	if i.Member == nil || i.Member.User == nil || i.Member.User.ID == "" {
		return
	}
	at, err := discordgo.SnowflakeTimestamp(i.ID)
	if err != nil {
		at = time.Now()
	}
	switch parts[2] {
	case "start":
		if len(parts) == 4 && parts[3] == "0" {
			b.startButton(s, i, parts[1], i.Message.ID, at)
		}
		return
	case string(core.GrinchSnowballs), string(core.GrinchDefense), string(core.GrinchAttack), string(core.GrinchFinalAttack):
	default:
		return // Not a button of the current fight message.
	}
	slot, err := strconv.Atoi(parts[3])
	if err != nil {
		return
	}
	var round int
	if len(parts) == 5 {
		round, err = strconv.Atoi(parts[4])
		if err != nil {
			return
		}
	}
	// Acknowledge before any work: Discord wants the first response within three seconds of the
	// click, and the store can be busy. The result shows in the fight message, and a failure is a
	// private follow-up.
	if !acknowledgeGrinchButton(s, i, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredMessageUpdate}) {
		return
	}
	playContext, stopPlay := context.WithTimeout(context.Background(), 2*time.Second)
	_, _, err = b.encounter.Play(playContext, parts[1], i.Message.ID, core.GrinchMove{
		Phase: core.GrinchPhase(parts[2]), Slot: slot, DefenseRound: round, FinalRound: round, Member: i.Member.User.ID, RequestID: i.ID, At: at,
	})
	stopPlay()
	if err != nil {
		logInteractionError(i.ID, "play Grinch action", err)
		privateFollowup(s, i, "Could not record your action. Try again.")
	}
}

// acknowledgeGrinchButton sends the first response of an interaction.
func acknowledgeGrinchButton(s *discordgo.Session, i *discordgo.Interaction, response *discordgo.InteractionResponse) bool {
	ackContext, stopAck := context.WithTimeout(context.Background(), 2500*time.Millisecond)
	defer stopAck()
	if err := s.InteractionRespond(i, response, discordgo.WithContext(ackContext)); err != nil {
		logInteractionError(i.ID, "acknowledge Grinch action", err)
		return false
	}
	return true
}

// privateFollowup tells the player something after the interaction was acknowledged.
func privateFollowup(s *discordgo.Session, i *discordgo.Interaction, text string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := s.FollowupMessageCreate(i, true, &discordgo.WebhookParams{Content: text, Flags: discordgo.MessageFlagsEphemeral},
		discordgo.WithContext(ctx)); err != nil {
		logInteractionError(i.ID, "send Grinch follow-up", err)
	}
}

const grinchImageName = "grinch.jpg"
const maxImageName = "maxpee.jpg"

//go:embed grinch.jpg
var grinchImage []byte

// maxImage is shown when the Grinch survives the fight.
//
//go:embed maxpee.jpg
var maxImage []byte

//go:embed grinchdefense.gif
var grinchDefenseImage []byte

//go:embed grinchheart.jpg
var grinchHeartImage []byte

const (
	grinchDefenseImageName = "grinchdefense.gif"
	grinchHeartImageName   = "grinchheart.jpg"
)

type grinchPicture struct {
	name string
	data []byte
}

func (picture grinchPicture) file() *discordgo.File {
	return &discordgo.File{Name: picture.name, ContentType: http.DetectContentType(picture.data), Reader: bytes.NewReader(picture.data)}
}

// phasePicture is the picture of a running fight: the Grinch attacking while the players defend,
// his heart in the last 5%, and his portrait otherwise.
func phasePicture(phase core.GrinchPhase) grinchPicture {
	switch phase {
	case core.GrinchDefense:
		return grinchPicture{grinchDefenseImageName, grinchDefenseImage}
	case core.GrinchFinalAttack:
		return grinchPicture{grinchHeartImageName, grinchHeartImage}
	default:
		return grinchPicture{grinchImageName, grinchImage}
	}
}

// shownPicture remembers which picture each fight message carries, so a picture is uploaded only
// when it changes. After a restart nothing is remembered and the next edit uploads it again.
type shownPicture struct {
	mu            sync.Mutex
	message, name string
}

func (shown *shownPicture) is(message, name string) bool {
	if shown == nil {
		return false
	}
	shown.mu.Lock()
	defer shown.mu.Unlock()
	return shown.message == message && shown.name == name
}

func (shown *shownPicture) set(message, name string) {
	if shown == nil {
		return
	}
	shown.mu.Lock()
	defer shown.mu.Unlock()
	shown.message, shown.name = message, name
}

// pictureUploadTimeout is longer than a text edit needs, so a slow upload is not cut off and
// retried over and over.
const pictureUploadTimeout = 10 * time.Second

type grinchDisplay struct {
	session *discordgo.Session
	channel string
	rules   core.GrinchRules
	shown   *shownPicture
}

func (display grinchDisplay) Announce(ctx context.Context, state core.Grinch) (string, time.Time, error) {
	state.Phase = core.GrinchLobby
	state.StartedAt = time.Now()
	message, err := display.session.ChannelMessageSendComplex(display.channel, &discordgo.MessageSend{
		Components:      grinchView(display.rules, state, []string{grinchImageName}, false, false),
		Flags:           discordgo.MessageFlagsIsComponentsV2,
		Files:           []*discordgo.File{{Name: grinchImageName, ContentType: "image/jpeg", Reader: bytes.NewReader(grinchImage)}},
		AllowedMentions: &discordgo.MessageAllowedMentions{},
	}, discordgo.WithContext(ctx), discordgo.WithRetryOnRatelimit(false), discordgo.WithRestRetries(0))
	if err != nil {
		return "", time.Time{}, err
	}
	start := message.Timestamp
	if start.IsZero() {
		start = state.StartedAt
	}
	display.shown.set(message.ID, grinchImageName)
	return message.ID, start, nil
}

func (display grinchDisplay) Update(ctx context.Context, state core.Grinch) error {
	final := !state.EndedAt.IsZero() && state.Phase != core.GrinchCancelled
	_, overflow := grinchStandings(display.rules, state)
	overflow = final && overflow
	picture := phasePicture(state.Phase)
	if final && state.Phase != core.GrinchWon {
		// The Grinch survived: Max's picture replaces his.
		picture = grinchPicture{maxImageName, maxImage}
	}
	images := []string{picture.name}
	var files []*discordgo.File
	// The final message uploads its picture again, as replacing the attachments also drops stale
	// ones. A running fight uploads only when the picture changes.
	changed := state.Phase != core.GrinchCancelled && !final && !display.shown.is(state.MessageID, picture.name)
	if final || changed {
		files = append(files, picture.file())
		if overflow {
			files = append(files, &discordgo.File{Name: grinchStandingsFile, Reader: strings.NewReader(grinchStandingsText(display.rules, state))})
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.WithoutCancel(ctx), pictureUploadTimeout)
		defer cancel()
	}
	if state.Phase == core.GrinchCancelled {
		images = nil
	}
	edit := func(files []*discordgo.File, images []string, attach, denied bool) *discordgo.MessageEdit {
		components := grinchView(display.rules, state, images, attach, denied)
		edit := &discordgo.MessageEdit{
			ID: state.MessageID, Channel: display.channel, Components: &components, Flags: discordgo.MessageFlagsIsComponentsV2,
			AllowedMentions: &discordgo.MessageAllowedMentions{Parse: []discordgo.AllowedMentionType{}},
		}
		if len(files) > 0 {
			edit.Files, edit.Attachments = files, attachmentsFor(files)
		}
		return edit
	}
	_, err := display.session.ChannelMessageEditComplex(edit(files, images, overflow, false), discordgo.WithContext(ctx))
	var denied *discordgo.RESTError
	if err == nil && len(files) > 0 {
		display.shown.set(state.MessageID, picture.name)
	}
	if (overflow || changed) && errors.As(err, &denied) && denied.Message != nil && denied.Message.Code == 50013 {
		// Without permission to attach files nothing can be uploaded. The announcement's picture is the
		// Grinch's, so it stays only when he lost the fight. When he won, Max's picture cannot be shown
		// and the message has no picture rather than the Grinch's.
		var kept []string
		if state.Phase == core.GrinchWon {
			kept = []string{grinchImageName}
		}
		_, err = display.session.ChannelMessageEditComplex(edit(nil, kept, false, true), discordgo.WithContext(ctx))
	}
	var rest *discordgo.RESTError
	if errors.As(err, &rest) && rest.Message != nil && rest.Message.Code == 10008 {
		return nil // The original message was deleted.
	}
	return err
}

func grinchMessage(rules core.GrinchRules, state core.Grinch) (string, []discordgo.MessageComponent) {
	content := grinchtext.Health(state.Health(rules))
	if state.Phase == core.GrinchCancelled {
		return "The Grinch encounter was cancelled because no one pressed Start.", []discordgo.MessageComponent{}
	}
	// Players press Start to begin the fight. The combat screen has no Join button.
	join := discordgo.ActionsRow{Components: []discordgo.MessageComponent{
		discordgo.Button{Label: "Start", Emoji: &discordgo.ComponentEmoji{Name: "🎅"}, Style: discordgo.SuccessButton,
			CustomID: "grinch:" + state.ID + ":start:0"},
	}}
	if state.Phase == core.GrinchLobby {
		if state.CountdownAt.IsZero() {
			return fmt.Sprintf("%s\n🎄 **The Grinch is coming!**\nPress **Start** to begin the fight. It starts %s after the first press, or right away when %d players press it. **%d/%d to start early**.",
					content, spanText(rules.RegistrationDuration), rules.EarlyStartPlayers, len(state.Joined), rules.EarlyStartPlayers),
				[]discordgo.MessageComponent{join}
		}
		return fmt.Sprintf("%s\n🎄 **The Grinch is coming!**\nThe fight starts <t:%d:R>. Press **Start** to join it. **%d/%d to start early**.",
				content, state.LobbyDeadline(rules).Unix(), len(state.Joined), rules.EarlyStartPlayers),
			[]discordgo.MessageComponent{join}
	}
	if !state.EndedAt.IsZero() {
		if state.Phase == core.GrinchWon {
			return content + "\n🎉 The group defeated the Grinch!", []discordgo.MessageComponent{}
		}
		if state.Phase == core.GrinchDefeat {
			return content + "\n💥 The Grinch defeated all registered players!", []discordgo.MessageComponent{}
		}
		return content + "\n💨 The Grinch escaped!", []discordgo.MessageComponent{}
	}
	if deadline := state.Deadline(rules); !deadline.IsZero() {
		content += fmt.Sprintf("\nEscapes <t:%d:R>", deadline.Unix())
	}
	button := func(slot int, emoji, label string) discordgo.Button {
		return discordgo.Button{
			Label: label, Emoji: &discordgo.ComponentEmoji{Name: emoji}, Style: discordgo.PrimaryButton,
			CustomID: fmt.Sprintf("grinch:%s:%s:%d", state.ID, state.Phase, slot),
		}
	}
	var rows []discordgo.MessageComponent
	switch state.Phase {
	case core.GrinchSnowballs:
		milestone := core.FormatHundredths(rules.SnowballsBossHP)
		content += fmt.Sprintf("\nThrow snowballs to bring the Grinch down to **%s HP** before time runs out, or he wins. Then the %s timer starts.", milestone, timerName(rules.FightDuration))
		rows = []discordgo.MessageComponent{discordgo.ActionsRow{Components: []discordgo.MessageComponent{button(0, "☃️", "Attack")}}}
	case core.GrinchDefense, core.GrinchAttack:
		on, off := "🎄", "❄️"
		content += "\nThe Grinch attacks! Choose a tree to defend."
		if state.Phase == core.GrinchAttack {
			on, off = "🧊", "☃️"
			content = grinchtext.Health(state.Health(rules)) + fmt.Sprintf("\n“Something, something… ouch!”\nEscapes <t:%d:R>", state.Deadline(rules).Unix())
		}
		for row := range 2 {
			columns := make([]discordgo.MessageComponent, 4)
			for column := range columns {
				slot := row*4 + column
				emoji := off
				if state.Board&(1<<slot) != 0 {
					emoji = on
				}
				tile := button(slot, emoji, "")
				if state.Phase == core.GrinchDefense {
					tile.CustomID += ":" + strconv.Itoa(state.DefenseRound)
				}
				columns[column] = tile
			}
			rows = append(rows, discordgo.ActionsRow{Components: columns})
		}
	case core.GrinchFinalAttack:
		content += fmt.Sprintf("\n⚔️ Attack for %s damage, pick 🎁 for %s gift points, or pick ❄️ for %s.",
			core.FormatHundredths(rules.FinalHitDamage), giftPointRange(rules), snowballReward(rules))
		if state.FinalRound == 0 {
			rows = []discordgo.MessageComponent{discordgo.ActionsRow{Components: []discordgo.MessageComponent{button(0, "⚔️", "Attack"), button(core.GrinchSnowballSlot, "❄️", "Snowball")}}}
			break
		}
		tiles := make([]discordgo.MessageComponent, 0, core.GrinchSnowballSlot+1)
		for slot := range core.GrinchSnowballSlot {
			emoji := "⚔️"
			if state.FinalBoard&(1<<slot) != 0 {
				emoji = "🎁"
			}
			tile := button(slot, emoji, "")
			tile.CustomID += ":" + strconv.Itoa(state.FinalRound)
			tiles = append(tiles, tile)
		}
		snowball := button(core.GrinchSnowballSlot, "❄️", "")
		snowball.CustomID += ":" + strconv.Itoa(state.FinalRound)
		tiles = append(tiles, snowball)
		rows = []discordgo.MessageComponent{discordgo.ActionsRow{Components: tiles}}
	default:
		content += "\nThe snowball phase ended."
		rows = []discordgo.MessageComponent{}
	}
	if state.Joined != nil {
		content += fmt.Sprintf("\n🎅 **%d players**", len(state.Joined)) // There is no Join button. Anyone can attack.
	}
	return content, rows
}

// spanText words a duration for a sentence: "1 minute", "90 seconds".
func spanText(d time.Duration) string {
	if d%time.Minute == 0 {
		if minutes := int(d / time.Minute); minutes == 1 {
			return "1 minute"
		} else {
			return fmt.Sprintf("%d minutes", minutes)
		}
	}
	return fmt.Sprintf("%d seconds", int(d.Round(time.Second)/time.Second))
}

// timerName words a duration for the countdown hint: "one-minute", "90-second".
func timerName(d time.Duration) string {
	if d%time.Minute == 0 {
		if minutes := int(d / time.Minute); minutes == 1 {
			return "one-minute"
		} else {
			return fmt.Sprintf("%d-minute", minutes)
		}
	}
	return fmt.Sprintf("%d-second", int(d.Round(time.Second)/time.Second))
}

func giftPointRange(rules core.GrinchRules) string {
	if rules.GiftPointsMin == rules.GiftPointsMax {
		return strconv.Itoa(rules.GiftPointsMin)
	}
	return fmt.Sprintf("%d–%d", rules.GiftPointsMin, rules.GiftPointsMax)
}

func snowballReward(rules core.GrinchRules) string {
	if rules.FinalSnowballReward == 1 {
		return "1 Snowball"
	}
	return fmt.Sprintf("%d Snowballs", rules.FinalSnowballReward)
}

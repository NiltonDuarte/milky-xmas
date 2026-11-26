package discord

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"
)

// Discord cannot put a picture inside text, but it can put an emoji there. The bot keeps one
// application emoji per player, made from their avatar, and writes it before their mention.

const (
	// Discord's limits on the characters of a Components V2 message and of a plain one.
	boxTextLimit     = 4000
	messageTextLimit = 2000

	avatarSize      = "128"
	avatarFileLimit = 1 << 20
)

var mentionPattern = regexp.MustCompile(`<@(\d+)>`)

type avatarEmoji struct{ id, name, hash string }

type avatarRequest struct{ userID, url string }

type avatarEmojis struct {
	session *discordgo.Session
	guild   string

	mu     sync.Mutex
	appID  string
	emojis map[string]avatarEmoji // By user ID.
	asked  map[avatarRequest]bool // What was queued, so a failure is not retried on every message.
	queue  chan avatarRequest
}

func newAvatarEmojis(session *discordgo.Session, guild string) *avatarEmojis {
	return &avatarEmojis{
		session: session, guild: guild, emojis: make(map[string]avatarEmoji), asked: make(map[avatarRequest]bool),
		queue: make(chan avatarRequest, 500),
	}
}

// run reads the emojis the application already has, then makes the ones that are asked for, one at
// a time, until the context ends.
func (a *avatarEmojis) run(ctx context.Context, appID string) {
	if err := a.load(ctx, appID); err != nil {
		log.Printf("Read the avatar emojis: %v", err)
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case request := <-a.queue:
			attempt, cancel := context.WithTimeout(ctx, 20*time.Second)
			if err := a.make(attempt, request); err != nil && ctx.Err() == nil {
				log.Printf("Make an avatar emoji for %s: %v", request.userID, err)
			}
			cancel()
		}
	}
}

func (a *avatarEmojis) load(ctx context.Context, appID string) error {
	existing, err := a.session.ApplicationEmojis(appID, discordgo.WithContext(ctx))
	if err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.appID = appID
	for _, emoji := range existing {
		if userID, hash, ok := parseAvatarName(emoji.Name); ok {
			a.emojis[userID] = avatarEmoji{emoji.ID, emoji.Name, hash}
		}
	}
	return nil
}

// see notes a player's current avatar, so it is ready before they are mentioned.
func (a *avatarEmojis) see(member *discordgo.Member) {
	if a == nil || member == nil || member.User == nil || member.User.Bot || member.User.ID == "" {
		return
	}
	url := stillAvatar(member.AvatarURL(avatarSize))
	a.mu.Lock()
	defer a.mu.Unlock()
	if emoji, ok := a.emojis[member.User.ID]; !ok || emoji.hash != avatarHash(url) {
		a.askLocked(avatarRequest{member.User.ID, url})
	}
}

func (a *avatarEmojis) askLocked(request avatarRequest) {
	if a.asked[request] {
		return
	}
	select {
	case a.queue <- request:
		a.asked[request] = true
	default: // Too many at once. A later message asks again.
	}
}

func (a *avatarEmojis) make(ctx context.Context, request avatarRequest) error {
	url := request.url
	if url == "" {
		member, err := a.member(ctx, request.userID)
		if err != nil {
			return err
		}
		url = stillAvatar(member.AvatarURL(avatarSize))
	}
	hash := avatarHash(url)
	a.mu.Lock()
	old, had := a.emojis[request.userID]
	appID := a.appID
	a.mu.Unlock()
	if had && old.hash == hash {
		return nil
	}
	picture, err := a.picture(ctx, url)
	if err != nil {
		return err
	}
	name, err := avatarName(request.userID, hash)
	if err != nil {
		return err
	}
	created, err := a.session.ApplicationEmojiCreate(appID, &discordgo.EmojiParams{
		Name: name, Image: "data:image/png;base64," + base64.StdEncoding.EncodeToString(picture),
	}, discordgo.WithContext(ctx))
	if err != nil {
		return err
	}
	a.mu.Lock()
	a.emojis[request.userID] = avatarEmoji{created.ID, name, hash}
	a.mu.Unlock()
	if had {
		return a.session.ApplicationEmojiDelete(appID, old.id, discordgo.WithContext(ctx))
	}
	return nil
}

func (a *avatarEmojis) member(ctx context.Context, userID string) (*discordgo.Member, error) {
	if a.guild != "" {
		member, err := a.session.GuildMember(a.guild, userID, discordgo.WithContext(ctx))
		if err == nil && member.User != nil {
			member.GuildID = a.guild
			return member, nil
		}
	}
	// Someone who left the server still has an avatar.
	user, err := a.session.User(userID, discordgo.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return &discordgo.Member{User: user}, nil
}

// picture downloads an avatar and cuts it into a circle.
func (a *avatarEmojis) picture(ctx context.Context, url string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	response, err := a.session.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download avatar: %s", response.Status)
	}
	source, _, err := image.Decode(io.LimitReader(response.Body, avatarFileLimit))
	if err != nil {
		return nil, fmt.Errorf("read avatar: %w", err)
	}
	var out bytes.Buffer
	if err := png.Encode(&out, circle(source)); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// circle keeps the part of the picture inside the largest circle in its middle.
func circle(source image.Image) image.Image {
	bounds := source.Bounds()
	side := min(bounds.Dx(), bounds.Dy())
	left, top := bounds.Min.X+(bounds.Dx()-side)/2, bounds.Min.Y+(bounds.Dy()-side)/2
	round := image.NewNRGBA(image.Rect(0, 0, side, side))
	for y := range side {
		for x := range side {
			// Distances from the centre, doubled to stay in whole numbers.
			dx, dy := 2*x+1-side, 2*y+1-side
			if dx*dx+dy*dy <= side*side {
				round.Set(x, y, color.NRGBAModel.Convert(source.At(left+x, top+y)))
			}
		}
	}
	return round
}

// stillAvatar asks for a still picture in place of an animated one.
func stillAvatar(url string) string { return strings.Replace(url, ".gif", ".png", 1) }

func avatarHash(url string) string {
	sum := sha1.Sum([]byte(url))
	return hex.EncodeToString(sum[:3])
}

// avatarName carries the player and the picture, so the emojis can be matched again after a
// restart. Emoji names hold up to 32 letters, digits, and underscores.
func avatarName(userID, hash string) (string, error) {
	id, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return "", fmt.Errorf("user ID %q: %w", userID, err)
	}
	return "u" + strconv.FormatUint(id, 36) + "_" + hash, nil
}

func parseAvatarName(name string) (string, string, bool) {
	encoded, hash, found := strings.Cut(strings.TrimPrefix(name, "u"), "_")
	id, err := strconv.ParseUint(encoded, 36, 64)
	if !strings.HasPrefix(name, "u") || !found || err != nil || len(hash) != 6 {
		return "", "", false
	}
	return strconv.FormatUint(id, 10), hash, true
}

// decorate writes each mentioned player's avatar before the mention, while the budget of
// characters lasts. A player without an emoji yet gets one for next time.
func (a *avatarEmojis) decorate(text string, budget *int) string {
	if a == nil {
		return text
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return mentionPattern.ReplaceAllStringFunc(text, func(mention string) string {
		userID := mention[2 : len(mention)-1]
		emoji, ok := a.emojis[userID]
		if !ok {
			a.askLocked(avatarRequest{userID: userID})
			return mention
		}
		prefix := "<:" + emoji.name + ":" + emoji.id + "> "
		if len(prefix) > *budget {
			return mention
		}
		*budget -= len(prefix)
		return prefix + mention
	})
}

// message decorates the text of a plain message.
func (a *avatarEmojis) message(text string) string {
	budget := messageTextLimit - utf8.RuneCountInString(text)
	return a.decorate(text, &budget)
}

// blocks decorates the texts of a box, which share one limit.
func (a *avatarEmojis) blocks(blocks []string) []string {
	budget := boxTextLimit
	for _, block := range blocks {
		budget -= utf8.RuneCountInString(block)
	}
	decorated := make([]string, len(blocks))
	for i, block := range blocks {
		decorated[i] = a.decorate(block, &budget)
	}
	return decorated
}

// components decorates every text of a Components V2 message, without changing the given ones.
func (a *avatarEmojis) components(list []discordgo.MessageComponent) []discordgo.MessageComponent {
	if a == nil {
		return list
	}
	budget := boxTextLimit - textLength(list)
	return a.walk(list, &budget)
}

func textLength(list []discordgo.MessageComponent) int {
	total := 0
	for _, component := range list {
		switch component := component.(type) {
		case discordgo.Container:
			total += textLength(component.Components)
		case discordgo.TextDisplay:
			total += utf8.RuneCountInString(component.Content)
		}
	}
	return total
}

func (a *avatarEmojis) walk(list []discordgo.MessageComponent, budget *int) []discordgo.MessageComponent {
	decorated := make([]discordgo.MessageComponent, len(list))
	for i, component := range list {
		switch typed := component.(type) {
		case discordgo.Container:
			typed.Components = a.walk(typed.Components, budget)
			component = typed
		case discordgo.TextDisplay:
			typed.Content = a.decorate(typed.Content, budget)
			component = typed
		}
		decorated[i] = component
	}
	return decorated
}

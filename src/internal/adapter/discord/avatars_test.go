package discord

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestAvatarEmojis(t *testing.T) {
	session, err := discordgo.New("Bot fixture")
	if err != nil {
		t.Fatal(err)
	}
	var square bytes.Buffer
	source := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	for y := range 8 {
		for x := range 8 {
			source.Set(x, y, color.White)
		}
	}
	if err := png.Encode(&square, source); err != nil {
		t.Fatal(err)
	}
	var calls []string
	var uploaded discordgo.EmojiParams
	oldName, _ := avatarName("123456789012345678", "000000")
	session.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		calls = append(calls, r.Method+" "+r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:])
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/emojis"):
			return fakeResponse(http.StatusOK, `{"items":[{"id":"7","name":"`+oldName+`"},{"id":"8","name":"party"}]}`), nil
		case r.Method == http.MethodGet:
			return fakeResponse(http.StatusOK, square.String()), nil
		case r.Method == http.MethodPost:
			body, err := io.ReadAll(r.Body)
			if err != nil || json.Unmarshal(body, &uploaded) != nil {
				t.Fatalf("upload: %s %v", body, err)
			}
			return fakeResponse(http.StatusCreated, `{"id":"9"}`), nil
		default:
			return fakeResponse(http.StatusNoContent, ""), nil
		}
	})
	avatars := newAvatarEmojis(session, "200")
	if err := avatars.load(t.Context(), "100"); err != nil {
		t.Fatal(err)
	}
	mention := "<@123456789012345678>"
	if got := avatars.message(mention); got != "<:"+oldName+":7> "+mention {
		t.Fatalf("loaded emoji = %q", got)
	}

	// A new picture replaces the player's emoji.
	avatars.see(&discordgo.Member{User: &discordgo.User{ID: "123456789012345678", Avatar: "abc"}})
	if err := avatars.make(t.Context(), <-avatars.queue); err != nil {
		t.Fatal(err)
	}
	want := []string{"GET emojis", "GET abc.png", "POST emojis", "DELETE 7"}
	if !slices.Equal(calls, want) || !strings.HasPrefix(uploaded.Image, "data:image/png;base64,") {
		t.Fatalf("calls = %v, want %v; image %.40q", calls, want, uploaded.Image)
	}
	if got := avatars.message(mention); got != "<:"+uploaded.Name+":9> "+mention {
		t.Errorf("new emoji = %q", got)
	}
	if userID, _, ok := parseAvatarName(uploaded.Name); !ok || userID != "123456789012345678" || len(uploaded.Name) > 32 {
		t.Errorf("emoji name %q does not carry the player", uploaded.Name)
	}

	// An unknown player is mentioned as before and gets an emoji later.
	if got := avatars.message("<@42>"); got != "<@42>" || len(avatars.queue) != 1 {
		t.Errorf("unknown player = %q, %d queued", got, len(avatars.queue))
	}
	// Avatars never push a message over Discord's limit.
	long := strings.Repeat("a", boxTextLimit-len(mention)-5) + mention
	if got := avatars.blocks([]string{long}); got[0] != long {
		t.Errorf("a full box grew to %d characters", len(got[0]))
	}
	// Role mentions and a bot without avatars are left alone.
	var none *avatarEmojis
	if avatars.message("<@&5>") != "<@&5>" || none.message(mention) != mention {
		t.Error("changed what it should not")
	}
}

func TestAvatarIsRound(t *testing.T) {
	source := image.NewNRGBA(image.Rect(0, 0, 20, 10))
	for y := range 10 {
		for x := range 20 {
			source.Set(x, y, color.White)
		}
	}
	round := circle(source)
	_, _, _, corner := round.At(0, 0).RGBA()
	_, _, _, middle := round.At(5, 5).RGBA()
	if round.Bounds().Dx() != 10 || round.Bounds().Dy() != 10 || corner != 0 || middle == 0 {
		t.Errorf("circle = %v, corner alpha %d, middle alpha %d", round.Bounds(), corner, middle)
	}
}

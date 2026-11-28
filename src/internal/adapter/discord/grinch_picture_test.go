package discord

import (
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
)

func TestGrinchPictureFollowsThePhaseAndUploadsOnlyOnChange(t *testing.T) {
	session, err := discordgo.New("Bot fixture")
	if err != nil {
		t.Fatal(err)
	}
	var body string
	session.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		body = string(data)
		return fakeResponse(http.StatusOK, `{"id":"1"}`), nil
	})
	display := grinchDisplay{session, "400", core.DefaultTuning().Grinch, &shownPicture{}, nil}
	display.shown.set("m", grinchImageName)
	for _, step := range []struct {
		phase  core.GrinchPhase
		image  string
		upload bool
	}{
		{core.GrinchSnowballs, "grinch.jpg", false},
		{core.GrinchDefense, "grinchdefense.gif", true},
		{core.GrinchDefense, "grinchdefense.gif", false},
		{core.GrinchAttack, "grinch.jpg", true},
		{core.GrinchFinalAttack, "grinchfinal.jpg", true},
		{core.GrinchFinalAttack, "grinchfinal.jpg", false},
	} {
		state := fightWithPlayers(step.phase, 1)
		state.MessageID = "m"
		if err := display.Update(t.Context(), state); err != nil {
			t.Fatal(err)
		}
		var named []string
		for _, name := range []string{"grinch.jpg", "grinchdefense.gif", "grinchfinal.jpg"} {
			if strings.Contains(body, "attachment://"+name) {
				named = append(named, name)
			}
		}
		uploaded := strings.Contains(body, `filename="`+step.image+`"`)
		if !slices.Equal(named, []string{step.image}) || uploaded != step.upload {
			t.Errorf("%s: shows %v, uploaded %v; want %s, upload %v", step.phase, named, uploaded, step.image, step.upload)
		}
	}
}

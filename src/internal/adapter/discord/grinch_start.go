package discord

import (
	"context"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
)

const startErrorText = "Could not register your Start. Try again."

// startButton handles a press of Start on the fight message. The click is acknowledged silently
// first. The first press of a player gets one private confirmation. Pressing Start again, or after
// the fight began, shows nothing.
func (b *grinchBot) startButton(s *discordgo.Session, i *discordgo.Interaction, fightID, messageID string, at time.Time) {
	if !acknowledgeGrinchButton(s, i, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredMessageUpdate}) {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	state, reply, err := b.encounter.Play(ctx, fightID, messageID, core.GrinchMove{
		Phase: core.GrinchLobby, Member: i.Member.User.ID, RequestID: i.ID, At: at,
	})
	if err != nil {
		logInteractionError(i.ID, "register Grinch Start", err)
		privateFollowup(s, i, startErrorText)
		return
	}
	if state.Accepted[i.ID] { // Only this click's registration is accepted under its own request ID.
		privateFollowup(s, i, reply)
	}
}

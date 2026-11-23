package discord

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
)

const grinchRegistrationTitle = "Grinch fight registration"

type grinchRegistrationPanel struct {
	mu     sync.Mutex
	opened *discordgo.Interaction
}

func (b *grinchBot) registrationButton(s *discordgo.Session, i *discordgo.Interaction, fightID, messageID, openerID string, at time.Time) {
	userID := i.Member.User.ID
	value, _ := b.registration.LoadOrStore(userID, &grinchRegistrationPanel{})
	panel := value.(*grinchRegistrationPanel)
	if !panel.mu.TryLock() {
		acknowledgeGrinchButton(s, i, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredMessageUpdate})
		return
	}
	defer panel.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if openerID != "" {
		if panel.opened != nil && panel.opened.ID != openerID {
			clearRegistrationPanel(s, i)
			return
		}
		_, _, err := b.encounter.Play(ctx, fightID, messageID, core.GrinchMove{
			Phase: core.GrinchLobby, Member: userID, RequestID: i.ID, At: at,
		})
		if err != nil {
			registrationError(s, i, err)
			return
		}
		if clearRegistrationPanel(s, i) {
			panel.opened = nil
		}
		return
	}

	allowed, err := b.encounter.CanJoin(ctx, fightID, messageID, userID, at)
	if err != nil {
		registrationError(s, i, err)
		return
	}
	if panel.opened != nil {
		err := s.InteractionResponseDelete(panel.opened, discordgo.WithContext(ctx))
		if err != nil && !registrationGone(err) {
			registrationError(s, i, err)
			return
		}
		panel.opened = nil
	}
	if !allowed {
		acknowledgeGrinchButton(s, i, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredMessageUpdate})
		return
	}
	if acknowledgeGrinchButton(s, i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: grinchRegistrationTitle,
			Flags:   discordgo.MessageFlagsEphemeral,
			Components: []discordgo.MessageComponent{discordgo.ActionsRow{Components: []discordgo.MessageComponent{
				discordgo.Button{Label: "Join", Style: discordgo.SuccessButton, CustomID: "grinch:" + fightID + ":join:" + messageID + ":" + i.ID},
			}}},
		},
	}) {
		panel.opened = i
	}
}

func clearRegistrationPanel(s *discordgo.Session, i *discordgo.Interaction) bool {
	return acknowledgeGrinchButton(s, i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{Content: grinchRegistrationTitle, Components: []discordgo.MessageComponent{}},
	})
}

func registrationGone(err error) bool {
	var rest *discordgo.RESTError
	if !errors.As(err, &rest) || rest.Message == nil {
		return false
	}
	switch rest.Message.Code {
	case 10008, 10015, 50027: // Deleted message, unknown webhook, or expired token.
		return true
	default:
		return false
	}
}

func registrationError(s *discordgo.Session, i *discordgo.Interaction, err error) {
	logInteractionError(i.ID, "update Grinch registration", err)
	acknowledgeGrinchButton(s, i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral, Content: "Could not update your registration. Try again."},
	})
}

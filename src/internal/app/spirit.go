package app

import (
	"context"
	"fmt"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

// AuraStore reads the aura of every player of a server.
type AuraStore interface {
	ServerAuras(context.Context, string) ([]core.Aura, error)
}

// ChristmasSpirit adds up the server's good and bad deeds. Callers check moderator permission.
func (service Service) ChristmasSpirit(ctx context.Context, server string) (core.Spirit, error) {
	if server == "" {
		return core.Spirit{}, ErrInvalidRecipient
	}
	auras, err := service.auras.ServerAuras(ctx, server)
	if err != nil {
		return core.Spirit{}, fmt.Errorf("load server auras: %w", err)
	}
	return core.ServerSpirit(auras), nil
}

package app

import (
	"context"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

// ThrowSnowball moves one snowball of the chosen kind from the sender to the recipient and
// returns how many of that kind the recipient now has.
func (service Service) ThrowSnowball(ctx context.Context, sender Member, recipient Recipient, kind core.SnowballType) (int64, error) {
	switch kind {
	case core.NormalSnowball:
		transfer, err := service.transfer(ctx, sender, recipient, ReasonThrowSnowball, core.TransferSnowball)
		return transfer.recipient.Snowballs, err
	case core.YellowSnowball:
		transfer, err := service.transfer(ctx, sender, recipient, ReasonThrowSnowball, core.TransferYellowSnowball)
		return transfer.recipient.YellowSnowballs, err
	default:
		return 0, core.ErrSnowballType
	}
}

package app

import (
	"context"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

const LookCommand = "look"
const StockingSubcommand = "in-your-stocking"

func (service Service) ViewInventory(ctx context.Context, member Member) (core.Inventory, error) {
	record, err := service.gifts.Load(ctx, member)
	return record.Inventory, err
}

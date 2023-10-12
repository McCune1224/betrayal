package discord

import (
	"github.com/zekrotja/ken"
)

func NotConfessionalError(ctx ken.Context, channelID string) (err error) {
	return ErrorMessage(
		ctx,
		"Not Confessional",
		"This command can only be used in your confessional channel <#"+channelID+">",
	)
}

package discord

import (
	"fmt"
	"strings"

	"github.com/zekrotja/ken"
)

func NotConfessionalError(ctx ken.Context, channelID string) (err error) {
	return ErrorMessage(
		ctx,
		"Not Confessional",
		"This command can only be used in your confessional channel <#"+channelID+">",
	)
}

func NotAuthorizedError(ctx ken.Context) (err error) {
	return ErrorMessage(
		ctx,
		"Not Authorized For Command",
		fmt.Sprintf("Need One Of The Following Roles: %s", strings.Join(AdminRoles, ", ")),
	)
}

func AlexError(ctx ken.Context) (err error) {
	return ErrorMessage(
		ctx,
		"Unable to process command",
		"Alex is a bad programmer. Please yell at him.",
	)
}

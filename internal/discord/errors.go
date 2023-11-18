package discord

import (
	"fmt"
	"strings"

	"github.com/zekrotja/ken"
)

func NotConfessionalError(ctx ken.Context, channelID string) (err error) {
	return ErrorMessage(
		ctx, "Not Confessional", "This command can only be used in your confessional channel <#"+channelID+">")
}

// Returns
func NotAdminError(ctx ken.Context) (err error) {
	return ErrorMessage(
		ctx, "Not Authorized For Command", fmt.Sprintf("Need One Of The Following Roles: %s", strings.Join(AdminRoles, ", ")))
}

func AlexError(ctx ken.Context, title string) (err error) {
	return ErrorMessage(ctx, title, "Alex has horribly failed at something here, please bug "+MentionUser(McKusaID))
}

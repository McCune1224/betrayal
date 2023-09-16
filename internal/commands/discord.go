package commands

const mckusaID = "206268866714796032"

func Mention(userID string) string {
	return "<@" + userID + ">"
}

func Underline(s string) string {
	return "__" + s + "__"
}

func Bold(s string) string {
	return "**" + s + "**"
}

func Italic(s string) string {
	return "*" + s + "*"
}

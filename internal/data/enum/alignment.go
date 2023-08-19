package enum

type AlignmentEnum string

// I do love me a hacky enum
// enables for example:
// Alignment.GOOD = "GOOD"
var Alignment = struct {
	GOOD    AlignmentEnum
	EVIL    AlignmentEnum
	NEUTRAL AlignmentEnum
}{
	GOOD:    "GOOD",
	EVIL:    "EVIL",
	NEUTRAL: "NEUTRAL",
}

package enum

type StatusNameEnum string

var StatusName = struct {
	CURSED      StatusNameEnum
	FROZEN      StatusNameEnum
	PARALYZED   StatusNameEnum
	BURNED      StatusNameEnum
	EMPOWERED   StatusNameEnum
	DRUNK       StatusNameEnum
	RESTRAINED  StatusNameEnum
	DISABLED    StatusNameEnum
	BLACKMAILED StatusNameEnum
	DESPAIRED   StatusNameEnum
	LUCKY       StatusNameEnum
	UNLUCKY     StatusNameEnum
	MADNESS     StatusNameEnum
}{
	CURSED:      "CURSED",
	FROZEN:      "FROZEN",
	PARALYZED:   "PARALYZED",
	BURNED:      "BURNED",
	EMPOWERED:   "EMPOWERED",
	DRUNK:       "DRUNK",
	RESTRAINED:  "RESTRAINED",
	DISABLED:    "DISABLED",
	BLACKMAILED: "BLACKMAILED",
	DESPAIRED:   "DESPAIRED",
	LUCKY:       "LUCKY",
	UNLUCKY:     "UNLUCKY",
	MADNESS:     "MADNESS",
}

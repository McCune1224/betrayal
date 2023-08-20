package enum

type ChangeTypeEnum string

var ChangeType = struct {
	UPGRADE   ChangeTypeEnum
	DOWNGRADE ChangeTypeEnum
}{
	UPGRADE:   "UPGRADE",
	DOWNGRADE: "DOWNGRADE",
}

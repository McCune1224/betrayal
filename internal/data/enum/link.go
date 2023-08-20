package enum

type LinkTypeEnum string

var LinkType = struct {
	INFLICTION LinkTypeEnum
	CURE       LinkTypeEnum
}{
	INFLICTION: "INFLICTION",
	CURE:       "CURE",
}

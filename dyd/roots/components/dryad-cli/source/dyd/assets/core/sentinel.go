
package core

type Sentinel int

const (
	SentinelUnknown Sentinel = iota
	SentinelGarden
	SentinelRoot
	SentinelStem
	SentinelSprout
)

var sentinelToString = map[Sentinel]string{
	SentinelGarden: "garden",
	SentinelRoot: "root",
	SentinelStem: "stem",
	SentinelSprout: "sprout",
	SentinelUnknown: "unknown",
}

var stringToSentinel = map[string]Sentinel{
	"garden": SentinelGarden,
	"root": SentinelRoot,
	"stem": SentinelStem,
	"sprout": SentinelSprout,
	"unknown": SentinelUnknown,
}

func (s Sentinel) String() string {
	if str, ok := sentinelToString[s]; ok {
		return str
	}
	return "unknown"
}

func SentinelFromString(s string) (Sentinel) {
	if sentinel, ok := stringToSentinel[s]; ok {
		return sentinel
	}
	return SentinelUnknown
}
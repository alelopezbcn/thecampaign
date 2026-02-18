package types

type SpyTarget string

const (
	SpyTargetDeck   SpyTarget = "deck"
	SpyTargetPlayer SpyTarget = "player"
)

type SpyInfo struct {
	Target       SpyTarget
	TargetPlayer string // only set when Target == SpyTargetPlayer
}

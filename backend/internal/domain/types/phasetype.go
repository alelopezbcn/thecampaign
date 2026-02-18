package types

type PhaseType string

var (
	PhaseTypeDrawCard  PhaseType = "draw"
	PhaseTypeAttack    PhaseType = "attack"
	PhaseTypeSpySteal  PhaseType = "spy/steal"
	PhaseTypeBuy       PhaseType = "buy"
	PhaseTypeConstruct PhaseType = "construct"
	PhaseTypeEndTurn   PhaseType = "endturn"
)

package types

type Category string

var (
	CategoryAction      Category = "Action"
	CategoryInfo        Category = "Info"
	CategoryError       Category = "Error"
	CategorySkip        Category = "Skip"
	CategoryEndTurn     Category = "EndTurn"
	CategoryTurnExpired Category = "TurnExpired"
	CategoryElimination Category = "Elimination"
)

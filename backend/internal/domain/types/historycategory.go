package types

type Category string

var (
	CategoryAction      Category = "Action"
	CategoryInfo        Category = "Info"
	CategoryError       Category = "Error"
	CategorySkip        Category = "Skip"
	CategoryEndTurn     Category = "EndTurn"
	CategoryElimination Category = "Elimination"
)

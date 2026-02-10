package types

type Category string

var (
	CategoryAction      Category = "Action"
	CategoryInfo        Category = "Info"
	CategoryError       Category = "Error"
	CategorySkip        Category = "Skip"
	CategoryElimination Category = "Elimination"
)

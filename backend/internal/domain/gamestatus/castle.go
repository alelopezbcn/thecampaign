package gamestatus

type Castle struct {
	IsConstructed bool `json:"constructed"`
	ResourceCards int  `json:"resource_cards"`
	Value         int  `json:"value"`
}

func NewCastle(c CastleInput) Castle {
	return Castle{
		IsConstructed: c.IsConstructed,
		ResourceCards: c.ResourceCardsCount,
		Value:         c.Value,
	}
}

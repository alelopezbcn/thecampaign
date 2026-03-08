package gamestatus

type Castle struct {
	IsConstructed  bool `json:"constructed"`
	IsProtected    bool `json:"is_protected"`
	ResourceCards  int  `json:"resource_cards"`
	Value          int  `json:"value"`
	ResourcesToWin int  `json:"resources_to_win"`
}

func NewCastle(c CastleInput) Castle {
	return Castle{
		IsConstructed:  c.IsConstructed,
		IsProtected:    c.IsProtected,
		ResourceCards:  c.ResourceCardsCount,
		Value:          c.Value,
		ResourcesToWin: c.ResourcesToWin,
	}
}

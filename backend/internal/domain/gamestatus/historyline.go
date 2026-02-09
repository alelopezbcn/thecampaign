package gamestatus

import "github.com/alelopezbcn/thecampaign/internal/domain/types"

type HistoryLine struct {
	Msg   string `json:"msg"`
	Color string `json:"color"`
}

func NewHistoryLine(msg string, category types.Category) HistoryLine {
	return HistoryLine{
		Msg:   msg,
		Color: mapColor(category),
	}
}

func mapColor(category types.Category) string {
	switch category {
	case types.CategoryAction:
		return "#12eb70"
	case types.CategoryInfo:
		return "#33C1FF"
	case types.CategoryError:
		return "#FF3333"
	case types.CategorySkip:
		return "#959896"
	case types.CategoryEndTurn:
		return "#ef1a1a"
	case types.CategoryElimination:
		return "#8E44AD"
	default:
		return "#FFFFFF"
	}
}

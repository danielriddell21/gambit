package game

import "github.com/danielriddell21/gambit/pkg/chess"

func ResultText(res chess.Result, reason chess.DrawReason) string {
	switch res {
	case chess.WhiteWins:
		return "White wins — checkmate"
	case chess.BlackWins:
		return "Black wins — checkmate"
	case chess.Draw:
		if reason != chess.NotDraw {
			return "Draw — " + reason.String()
		}
		return "Draw"
	default:
		return "In progress"
	}
}

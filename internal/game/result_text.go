package game

import "github.com/danielriddell21/gambit/pkg/chess"

// ResultText returns a short human-readable description of a finished game, e.g.
// "White wins — checkmate", "Draw — threefold repetition", "Draw — stalemate".
// It is shared by the terminal logger and the GUI banner.
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

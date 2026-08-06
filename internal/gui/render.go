//go:build ebiten

package gui

import (
	"github.com/danielriddell21/gambit/pkg/chess"
)

// view snapshots the UI state the frame composer needs.
func (u *GameUI) view() BoardView {
	return BoardView{
		Snapshot:      u.snapshot,
		Flipped:       u.flipped,
		Selected:      u.selected,
		Legal:         u.game.Board().LegalMoves(),
		WhiteName:     u.game.AgentName(chess.White),
		BlackName:     u.game.AgentName(chess.Black),
		Moves:         len(u.game.Moves()),
		Over:          u.game.Over(),
		Paused:        u.paused,
		AwaitingHuman: u.currentIsHuman(),
		Result:        u.game.Result(),
		DrawReason:    u.game.DrawReason(),
	}
}

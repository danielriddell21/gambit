// Package log provides terminal logging of game progress, formatting each move
// in SAN for human readers.
package log

import (
	"fmt"
	"io"
	"os"

	"github.com/danielriddell21/gambit/internal/game"
	"github.com/danielriddell21/gambit/pkg/chess"
)

// Logger writes move and result lines to an output stream.
type Logger struct {
	w io.Writer
}

// New returns a logger writing to stdout.
func New() *Logger {
	return &Logger{w: os.Stdout}
}

// NewWriter returns a logger writing to w (useful for tests).
func NewWriter(w io.Writer) *Logger {
	return &Logger{w: w}
}

// Move logs a single completed ply, e.g.
//
//  1. White (minimax)  e4        (e2e4)
func (l *Logger) Move(ev game.MoveEvent) {
	moveNo := (ev.Ply + 1) / 2
	dot := "."
	if ev.Mover == chess.Black {
		dot = "..."
	}
	fmt.Fprintf(l.w, "%3d%-3s %-5s (%-8s) %-7s (%s)\n",
		moveNo, dot, ev.Mover, ev.AgentName, ev.SAN, ev.Move)
}

// Result logs the final outcome of a game.
func (l *Logger) Result(res chess.Result, reason chess.DrawReason) {
	switch res {
	case chess.WhiteWins:
		fmt.Fprintf(l.w, "result: %s (white wins by checkmate)\n", res)
	case chess.BlackWins:
		fmt.Fprintf(l.w, "result: %s (black wins by checkmate)\n", res)
	case chess.Draw:
		fmt.Fprintf(l.w, "result: %s (draw by %s)\n", res, reason)
	default:
		fmt.Fprintf(l.w, "result: %s\n", res)
	}
}

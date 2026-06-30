package log

import (
	"fmt"
	"io"
	"os"

	"github.com/danielriddell21/gambit/internal/game"
	"github.com/danielriddell21/gambit/pkg/chess"
)

type Logger struct {
	w io.Writer
}

func New() *Logger {
	return &Logger{w: os.Stdout}
}

func NewWriter(w io.Writer) *Logger {
	return &Logger{w: w}
}

func (l *Logger) Move(ev game.MoveEvent) {
	moveNo := (ev.Ply + 1) / 2
	dot := "."
	if ev.Mover == chess.Black {
		dot = "..."
	}
	_, _ = fmt.Fprintf(l.w, "%3d%-3s %-5s (%-8s) %-7s (%s)\n",
		moveNo, dot, ev.Mover, ev.AgentName, ev.SAN, ev.Move)
}

func (l *Logger) Result(res chess.Result, reason chess.DrawReason) {
	_, _ = fmt.Fprintf(l.w, "result: %s (%s)\n", res, game.ResultText(res, reason))
}

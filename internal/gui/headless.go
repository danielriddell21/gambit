package gui

import (
	"context"
	"fmt"
	"image"

	"github.com/danielriddell21/crucible/canvas"
	"github.com/danielriddell21/crucible/demo"
	"github.com/danielriddell21/crucible/record"

	"github.com/danielriddell21/gambit/pkg/chess"
)

// recordHold keeps the final position on screen for four seconds before a GIF
// loops back to the opening.
const recordHold = 400

// Render records a game to cfg.Rec.Path without opening a window, one frame
// per move. Frames come from the same [DrawFrame] the window uses, so the
// media matches what a player sees — and it needs no display.
//
// The file extension picks the format: .gif or .mp4.
func Render(cfg Config) error {
	barHeight := BarHeight(cfg.SquareSize)
	faces, err := LoadFaces(cfg.SquareSize, barHeight)
	if err != nil {
		return err
	}
	w, h := FrameSize(cfg)
	c := canvas.New(w, h)

	delay := cfg.RecordDelay
	if delay <= 0 {
		delay = 70
	}
	opts := []record.Option{record.WithFrameDelay(delay)}
	if record.IsVideoPath(cfg.Rec.Path) {
		opts = append(opts, record.WithVideo())
	} else {
		// The board is a handful of flat colours, so its own palette quantises
		// cleanly and delta frames keep the file small.
		opts = append(opts, record.WithPalette(demoPalette), record.WithFrameDiff(), record.WithFinalHold(recordHold))
	}
	rec := record.New(cfg.Rec, opts...)

	g := cfg.Game
	view := func() BoardView {
		var snap [64]chess.Piece
		g.Board().Each(func(s chess.Square, p chess.Piece) { snap[s] = p })
		return BoardView{
			Snapshot:   snap,
			WhiteName:  g.AgentName(chess.White),
			BlackName:  g.AgentName(chess.Black),
			Moves:      len(g.Moves()),
			Over:       g.Over(),
			Selected:   chess.NoSquare,
			Result:     g.Result(),
			DrawReason: g.DrawReason(),
		}
	}

	clip := demo.Clip{
		// One frame per move: the clip ends with the game, unless
		// cfg.Rec.Frames caps it sooner.
		Frames:   cfg.Rec.Frames,
		MaxSteps: maxRecordedMoves,
		Stop:     func(int) bool { return g.Over() },
		Step: func(int) error {
			if g.Over() {
				return nil
			}
			ev, ok, err := g.Step(context.Background())
			if err != nil {
				return fmt.Errorf("play move: %w", err)
			}
			if ok && cfg.Logger != nil {
				cfg.Logger.Move(ev)
			}
			return nil
		},
		Frame: func(int) image.Image {
			DrawFrame(c, cfg, faces, view())
			return record.FromRGBA(c.Pixels(), w, h)
		},
	}
	if _, err := clip.Record(rec); err != nil {
		return fmt.Errorf("capture game: %w", err)
	}
	if cfg.Logger != nil {
		cfg.Logger.Result(g.Result(), g.DrawReason())
	}
	if err := rec.Save(cfg.Rec.Path); err != nil {
		return fmt.Errorf("save recording: %w", err)
	}
	fmt.Printf("%s: %d frames\n", cfg.Rec.Path, rec.Len())
	return nil
}

// maxRecordedMoves bounds a recording that is not frame-capped, so a game that
// never resolves still terminates.
const maxRecordedMoves = 400

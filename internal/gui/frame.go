package gui

import (
	"fmt"
	"image/color"

	"golang.org/x/image/font"

	"github.com/danielriddell21/crucible/canvas"
	"github.com/danielriddell21/crucible/keymap"

	"github.com/danielriddell21/gambit/internal/game"
	"github.com/danielriddell21/gambit/pkg/chess"
)

var (
	pieceWhite = color.RGBA{R: 0xf5, G: 0xf5, B: 0xf0, A: 0xff}
	pieceBlack = color.RGBA{R: 0x20, G: 0x20, B: 0x24, A: 0xff}

	selectTint = color.RGBA{R: 0x2e, G: 0x8b, B: 0x57, A: 0x99}
	targetTint = color.RGBA{R: 0x2e, G: 0x8b, B: 0x57, A: 0x55}
)

// hints is the control bar under the board, laid out by crucible/keymap so it
// reads the same as every other app in the family.
var hints = []keymap.Binding{
	{Key: "space", Action: "pause"},
	{Key: "n", Action: "step"},
	{Key: "r", Action: "restart"},
	{Key: "f", Action: "flip"},
	{Key: "+/-", Action: "speed"},
}

// BoardView is everything a frame draws: the position, what the player has
// selected, and how the game stands.
type BoardView struct {
	Snapshot [64]chess.Piece
	Flipped  bool
	Selected chess.Square
	// Legal is the current legal move list, used to light up the squares the
	// selected piece can reach.
	Legal []chess.Move

	WhiteName, BlackName string
	Moves                int
	Over                 bool
	Paused               bool
	AwaitingHuman        bool
	Result               chess.Result
	DrawReason           chess.DrawReason
}

// DrawFrame composes a whole frame onto c: the board, the highlights, the
// pieces, the info bar and — once the game ends — the result banner. Drawing
// through the software canvas rather than the display means the same code
// paints a live window and a headless recording, pixel for pixel.
func DrawFrame(c *canvas.Canvas, cfg Config, f Faces, v BoardView) {
	size := cfg.SquareSize
	boardSize := size * 8

	for rank := range 8 {
		for file := range 8 {
			x, y := squareTopLeft(file, rank, size, v.Flipped)
			col := cfg.LightSquare
			if (file+rank)%2 == 0 {
				col = cfg.DarkSquare
			}
			c.Rect(x, y, size, size, col)
		}
	}
	drawHighlights(c, size, v)
	drawPieces(c, size, f.Piece, v)
	drawInfoBar(c, cfg, boardSize, f.Bar, v)
	if v.Over {
		drawBanner(c, size, boardSize, f.Banner, v)
	}
}

func drawHighlights(c *canvas.Canvas, size int, v BoardView) {
	if v.Selected == chess.NoSquare {
		return
	}
	sx, sy := squareTopLeft(v.Selected.File(), v.Selected.Rank(), size, v.Flipped)
	c.Blend(sx, sy, size, size, selectTint)
	for _, m := range v.Legal {
		if m.From() != v.Selected {
			continue
		}
		tx, ty := squareTopLeft(m.To().File(), m.To().Rank(), size, v.Flipped)
		c.Blend(tx, ty, size, size, targetTint)
	}
}

func drawPieces(c *canvas.Canvas, size int, face font.Face, v BoardView) {
	ascent := face.Metrics().Ascent.Ceil()
	height := face.Metrics().Height.Ceil()
	for s := chess.Square(0); s < 64; s++ {
		p := v.Snapshot[s]
		if p.IsEmpty() {
			continue
		}
		glyph := string(solidGlyph[p.Type()])
		fill, outline := pieceWhite, pieceBlack
		if p.Color() == chess.Black {
			fill, outline = pieceBlack, pieceWhite
		}
		x, y := squareTopLeft(p2f(s), s.Rank(), size, v.Flipped)
		cx := x + (size-canvas.MeasureFace(glyph, face))/2
		cy := y + (size-height)/2 + ascent

		// Outline first, offset in four directions, then the fill on top.
		for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
			c.TextFace(cx+d[0], cy+d[1], glyph, outline, face)
		}
		c.TextFace(cx, cy, glyph, fill, face)
	}
}

func drawInfoBar(c *canvas.Canvas, cfg Config, boardSize int, face font.Face, v BoardView) {
	c.Rect(0, boardSize, boardSize, BarHeight(cfg.SquareSize), cfg.DarkSquare)

	status := fmt.Sprintf("white: %s    black: %s    move %d", v.WhiteName, v.BlackName, (v.Moves+1)/2)
	switch {
	case v.Over:
		status += "    [over]"
	case v.AwaitingHuman:
		status += "    [your move — click a piece]"
	case v.Paused:
		status += "    [paused]"
	}

	pad := int(float64(cfg.SquareSize) * 0.08)
	lineH := face.Metrics().Height.Ceil()
	ascent := face.Metrics().Ascent.Ceil()
	c.TextFace(pad, boardSize+pad+ascent, status, pieceWhite, face)

	// The hints hug the bottom of the bar, wrapping if the board is drawn small
	// enough that they no longer fit on one row.
	_, frameH := FrameSize(cfg)
	hf := keymap.Face{LineHeight: lineH, Measure: func(s string) int { return canvas.MeasureFace(s, face) }}
	for _, line := range keymap.BottomBar(hints, boardSize, frameH, pad, hf) {
		c.TextFace(line.X, line.Y+ascent, line.Text, pieceWhite, face)
	}
}

func drawBanner(c *canvas.Canvas, size, boardSize int, face font.Face, v BoardView) {
	msg := game.ResultText(v.Result, v.DrawReason)
	bandH := int(float64(size) * 1.4)
	bandY := boardSize/2 - bandH/2
	c.Blend(0, bandY, boardSize, bandH, color.RGBA{R: pieceBlack.R, G: pieceBlack.G, B: pieceBlack.B, A: 200})

	th := face.Metrics().Height.Ceil()
	tx := (boardSize - canvas.MeasureFace(msg, face)) / 2
	ty := bandY + (bandH-th)/2 + face.Metrics().Ascent.Ceil()
	c.TextFace(tx, ty, msg, pieceWhite, face)
}

// squareTopLeft is where a board square starts on screen, honouring the flip.
func squareTopLeft(file, rank, size int, flipped bool) (x, y int) {
	col, row := file, 7-rank
	if flipped {
		col, row = 7-file, rank
	}
	return col * size, row * size
}

// p2f is a square's file, named apart from the method so the loop above reads
// clearly alongside its rank.
func p2f(s chess.Square) int { return s.File() }

// FrameSize is the window's pixel size for a given square size.
func FrameSize(cfg Config) (w, h int) {
	return cfg.SquareSize * 8, cfg.SquareSize*8 + BarHeight(cfg.SquareSize)
}

package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/danielriddell21/gambit/pkg/chess"
)

// drawBoard fills the 8x8 grid of light and dark squares.
func (u *GameUI) drawBoard(screen *ebiten.Image) {
	size := float32(u.cfg.SquareSize)
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			x, y := u.squareTopLeft(file, rank)
			c := u.cfg.LightSquare
			if (file+rank)%2 == 0 {
				c = u.cfg.DarkSquare
			}
			vector.FillRect(screen, x, y, size, size, c, false)
		}
	}
}

// drawPieces draws every piece from the cached snapshot.
func (u *GameUI) drawPieces(screen *ebiten.Image) {
	for s := chess.Square(0); s < 64; s++ {
		p := u.snapshot[s]
		if p.IsEmpty() {
			continue
		}
		u.drawPiece(screen, p, s.File(), s.Rank())
	}
}

// drawPiece renders a single piece glyph centered in its square, with an
// outline in the contrasting color so both colors read on any square.
func (u *GameUI) drawPiece(screen *ebiten.Image, p chess.Piece, file, rank int) {
	glyph := string(solidGlyph[p.Type()])

	fill := pieceWhite
	outline := pieceBlack
	if p.Color() == chess.Black {
		fill, outline = pieceBlack, pieceWhite
	}

	x, y := u.squareTopLeft(file, rank)
	gw, gh := text.Measure(glyph, u.face, 0)
	cx := float64(x) + (float64(u.cfg.SquareSize)-gw)/2
	cy := float64(y) + (float64(u.cfg.SquareSize)-gh)/2

	// Outline: draw the glyph offset in four directions.
	for _, d := range [][2]float64{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
		drawGlyph(screen, glyph, u.face, cx+d[0], cy+d[1], outline)
	}
	drawGlyph(screen, glyph, u.face, cx, cy, fill)
}

func drawGlyph(screen *ebiten.Image, glyph string, face text.Face, x, y float64, c color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(c)
	text.Draw(screen, glyph, face, op)
}

// squareTopLeft returns the pixel coordinates of a square's top-left corner.
// Rank 8 is drawn at the top, file a at the left.
func (u *GameUI) squareTopLeft(file, rank int) (x, y float32) {
	size := float32(u.cfg.SquareSize)
	return float32(file) * size, float32(7-rank) * size
}

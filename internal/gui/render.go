//go:build ebiten

package gui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/danielriddell21/gambit/internal/game"
	"github.com/danielriddell21/gambit/pkg/chess"
)

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

func (u *GameUI) drawPieces(screen *ebiten.Image) {
	for s := chess.Square(0); s < 64; s++ {
		p := u.snapshot[s]
		if p.IsEmpty() {
			continue
		}
		u.drawPiece(screen, p, s.File(), s.Rank())
	}
}

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
		drawText(screen, glyph, u.face, cx+d[0], cy+d[1], outline)
	}
	drawText(screen, glyph, u.face, cx, cy, fill)
}

func (u *GameUI) drawInfoBar(screen *ebiten.Image) {
	boardSize := float32(u.boardSize())
	vector.FillRect(screen, 0, boardSize, boardSize, float32(u.barHeight), u.cfg.DarkSquare, false)

	moveNo := (len(u.game.Moves()) + 1) / 2
	status := fmt.Sprintf("white: %s    black: %s    move %d",
		u.game.AgentName(chess.White), u.game.AgentName(chess.Black), moveNo)
	switch {
	case u.game.Over():
		status += "    [over]"
	case u.currentIsHuman():
		status += "    [your move — click a piece]"
	case u.paused:
		status += "    [paused]"
	}
	const hints = "space pause · n step · r restart · f flip · +/- speed"

	pad := float64(u.cfg.SquareSize) * 0.08
	_, lineH := text.Measure("Ay", u.barFace, 0)
	y0 := float64(u.boardSize()) + pad
	drawText(screen, status, u.barFace, pad, y0, pieceWhite)
	drawText(screen, hints, u.barFace, pad, y0+lineH, pieceWhite)
}

func (u *GameUI) drawBanner(screen *ebiten.Image) {
	boardSize := float32(u.boardSize())
	msg := game.ResultText(u.game.Result(), u.game.DrawReason())

	bandH := float32(u.cfg.SquareSize) * 1.4
	bandY := boardSize/2 - bandH/2
	overlay := color.RGBA{R: pieceBlack.R, G: pieceBlack.G, B: pieceBlack.B, A: 200}
	vector.FillRect(screen, 0, bandY, boardSize, bandH, overlay, false)

	tw, th := text.Measure(msg, u.bannerFace, 0)
	tx := (float64(u.boardSize()) - tw) / 2
	ty := float64(bandY) + (float64(bandH)-th)/2
	drawText(screen, msg, u.bannerFace, tx, ty, pieceWhite)
}

func (u *GameUI) drawHighlights(screen *ebiten.Image) {
	if u.selected == chess.NoSquare {
		return
	}
	size := float32(u.cfg.SquareSize)
	selectTint := color.RGBA{R: 0x2e, G: 0x8b, B: 0x57, A: 0x99}
	targetTint := color.RGBA{R: 0x2e, G: 0x8b, B: 0x57, A: 0x55}

	sx, sy := u.squareTopLeft(u.selected.File(), u.selected.Rank())
	vector.FillRect(screen, sx, sy, size, size, selectTint, false)

	for _, m := range u.game.Board().LegalMoves() {
		if m.From() != u.selected {
			continue
		}
		tx, ty := u.squareTopLeft(m.To().File(), m.To().Rank())
		vector.FillRect(screen, tx, ty, size, size, targetTint, false)
	}
}

func drawText(screen *ebiten.Image, s string, face text.Face, x, y float64, c color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(c)
	text.Draw(screen, s, face, op)
}

func (u *GameUI) squareTopLeft(file, rank int) (x, y float32) {
	size := float32(u.cfg.SquareSize)
	col, row := file, 7-rank
	if u.flipped {
		col, row = 7-file, rank
	}
	return float32(col) * size, float32(row) * size
}

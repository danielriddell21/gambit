// Command gifgen renders an agent-vs-agent game to an animated GIF. It is used
// to regenerate the demos under docs/demos. It reuses the chess engine, agents
// and the shared font, and renders frames with the standard library so it needs
// no GUI or external tools.
package main

import (
	"context"
	"flag"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"log"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"github.com/danielriddell21/gambit/internal/agent"
	"github.com/danielriddell21/gambit/internal/assets"
	"github.com/danielriddell21/gambit/internal/game"
	"github.com/danielriddell21/gambit/pkg/chess"
)

var (
	lightSquare = color.RGBA{R: 0xec, G: 0xd9, B: 0xb6, A: 0xff}
	darkSquare  = color.RGBA{R: 0xa9, G: 0x7a, B: 0x55, A: 0xff}
	pieceWhite  = color.RGBA{R: 0xf5, G: 0xf5, B: 0xf0, A: 0xff}
	pieceBlack  = color.RGBA{R: 0x20, G: 0x20, B: 0x24, A: 0xff}
)

var solidGlyph = map[chess.PieceType]rune{
	chess.King:   '♚',
	chess.Queen:  '♛',
	chess.Rook:   '♜',
	chess.Bishop: '♝',
	chess.Knight: '♞',
	chess.Pawn:   '♟',
}

func main() {
	white := flag.String("white", "minimax", "white agent")
	black := flag.String("black", "random", "black agent")
	depth := flag.Int("depth", 4, "search depth")
	seed := flag.Int64("seed", 7, "RNG seed")
	square := flag.Int("square", 56, "square size in pixels")
	delay := flag.Int("delay", 70, "frame delay in 1/100s")
	out := flag.String("out", "docs/demos/game.gif", "output GIF path")
	flag.Parse()

	if err := generate(*white, *black, *depth, *seed, *square, *delay, *out); err != nil {
		log.Fatalf("gifgen: %v", err)
	}
}

func generate(white, black string, depth int, seed int64, square, delay int, out string) error {
	w, err := agent.New(white, agent.Options{Seed: seed, Depth: depth})
	if err != nil {
		return err
	}
	b, err := agent.New(black, agent.Options{Seed: seed + 1, Depth: depth})
	if err != nil {
		return err
	}

	face, err := newFace(float64(square) * 0.82)
	if err != nil {
		return err
	}

	g := game.New(game.Players{White: w, Black: b}, nil)

	var frames []*image.Paletted
	var delays []int
	add := func() {
		frames = append(frames, quantize(renderBoard(g.Board(), square, face)))
		delays = append(delays, delay)
	}

	add() // initial position
	if err := g.Run(context.Background(), func(game.MoveEvent) { add() }); err != nil {
		return err
	}
	if len(delays) > 0 {
		delays[len(delays)-1] = 400 // hold the final position
	}

	f, err := os.Create(out)
	if err != nil {
		return err
	}
	if err := gif.EncodeAll(f, &gif.GIF{Image: frames, Delay: delays}); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

func newFace(size float64) (font.Face, error) {
	ft, err := opentype.Parse(assets.Font)
	if err != nil {
		return nil, err
	}
	return opentype.NewFace(ft, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
}

func renderBoard(b *chess.Board, sq int, face font.Face) *image.RGBA {
	side := sq * 8
	img := image.NewRGBA(image.Rect(0, 0, side, side))
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			x, y := file*sq, (7-rank)*sq
			c := lightSquare
			if (file+rank)%2 == 0 {
				c = darkSquare
			}
			draw.Draw(img, image.Rect(x, y, x+sq, y+sq), &image.Uniform{C: c}, image.Point{}, draw.Src)
		}
	}
	b.Each(func(s chess.Square, p chess.Piece) {
		if p.IsEmpty() {
			return
		}
		fill, outline := pieceWhite, pieceBlack
		if p.Color() == chess.Black {
			fill, outline = pieceBlack, pieceWhite
		}
		drawGlyph(img, face, string(solidGlyph[p.Type()]), s.File()*sq, (7-s.Rank())*sq, sq, fill, outline)
	})
	return img
}

func drawGlyph(img *image.RGBA, face font.Face, glyph string, x, y, sq int, fill, outline color.RGBA) {
	d := &font.Drawer{Dst: img, Face: face}
	adv := d.MeasureString(glyph).Round()
	baseX := x + (sq-adv)/2
	baseY := y + int(float64(sq)*0.80)

	for _, off := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
		d.Src = &image.Uniform{C: outline}
		d.Dot = fixed.P(baseX+off[0], baseY+off[1])
		d.DrawString(glyph)
	}
	d.Src = &image.Uniform{C: fill}
	d.Dot = fixed.P(baseX, baseY)
	d.DrawString(glyph)
}

// demoPalette holds blends between the handful of colors actually used, so
// nearest-color mapping reproduces both the flat squares and the anti-aliased
// glyph edges without dithering artifacts.
var demoPalette = buildPalette()

func buildPalette() color.Palette {
	base := []color.RGBA{lightSquare, darkSquare, pieceWhite, pieceBlack}
	var pal color.Palette
	for _, c := range base {
		pal = append(pal, c)
	}
	const steps = 40
	for i := range base {
		for j := i + 1; j < len(base); j++ {
			for s := 1; s < steps; s++ {
				pal = append(pal, lerp(base[i], base[j], float64(s)/steps))
			}
		}
	}
	return pal
}

func lerp(a, b color.RGBA, t float64) color.RGBA {
	mix := func(x, y uint8) uint8 { return uint8(float64(x)*(1-t) + float64(y)*t) }
	return color.RGBA{R: mix(a.R, b.R), G: mix(a.G, b.G), B: mix(a.B, b.B), A: 0xff}
}

func quantize(rgba *image.RGBA) *image.Paletted {
	pal := image.NewPaletted(rgba.Bounds(), demoPalette)
	draw.Draw(pal, pal.Bounds(), rgba, image.Point{}, draw.Src)
	return pal
}

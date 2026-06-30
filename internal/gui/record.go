//go:build ebiten

package gui

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

// recorder accumulates rendered frames and encodes them as an animated GIF. It
// captures the real Ebiten output via Image.ReadPixels, so the demo matches the
// window exactly.
type recorder struct {
	path   string
	delay  int // per-frame delay in 1/100s
	frames []*image.Paletted
	delays []int
}

func newRecorder(path string, delay int) *recorder {
	return &recorder{path: path, delay: delay}
}

// capture grabs the current screen image as one frame.
func (r *recorder) capture(screen *ebiten.Image) {
	b := screen.Bounds()
	buf := make([]byte, 4*b.Dx()*b.Dy())
	screen.ReadPixels(buf)
	rgba := &image.RGBA{Pix: buf, Stride: 4 * b.Dx(), Rect: image.Rect(0, 0, b.Dx(), b.Dy())}

	pal := image.NewPaletted(rgba.Rect, demoPalette)
	for y := rgba.Rect.Min.Y; y < rgba.Rect.Max.Y; y++ {
		for x := rgba.Rect.Min.X; x < rgba.Rect.Max.X; x++ {
			pal.Set(x, y, rgba.At(x, y))
		}
	}
	r.frames = append(r.frames, pal)
	r.delays = append(r.delays, r.delay)
}

// save writes the accumulated frames to the GIF file, holding the last frame.
func (r *recorder) save() error {
	if len(r.delays) > 0 {
		r.delays[len(r.delays)-1] = 400
	}
	f, err := os.Create(r.path)
	if err != nil {
		return fmt.Errorf("create gif: %w", err)
	}
	if err := gif.EncodeAll(f, &gif.GIF{Image: r.frames, Delay: r.delays}); err != nil {
		_ = f.Close()
		return fmt.Errorf("encode gif: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close gif: %w", err)
	}
	return nil
}

// demoPalette holds blends between the handful of colors the board uses, so
// nearest-color mapping reproduces both the flat squares and the anti-aliased
// glyph edges without dithering artifacts.
var demoPalette = buildPalette()

func buildPalette() color.Palette {
	base := []color.RGBA{
		DefaultConfig().LightSquare,
		DefaultConfig().DarkSquare,
		pieceWhite,
		pieceBlack,
	}
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

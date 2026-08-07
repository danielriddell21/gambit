package gui

import "image/color"

// demoPalette is tuned to the board's own colours — the two square shades and
// the two piece shades, plus blends between them — so the GIF quantises
// cleanly instead of through a generic web palette.
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

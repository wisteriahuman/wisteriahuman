// Package theme は SVG の配色を持つ。ライトは wisteria-io.com のトークンに合わせる。
package theme

import (
	"math"
	"strconv"
)

// Palette は1つの配色。色はすべて #rrggbb。
type Palette struct {
	Name       string
	Bg         string // 画像の背景
	Surface    string // カードの面
	Border     string // 枠線
	Ink        string // 本文
	InkMuted   string // 補足
	Pen        string // リンク・強調
	Accent     string // 藤色（文字にも使う）
	AccentSoft string // 藤色の面・飾り
}

var Light = Palette{
	Name:       "light",
	Bg:         "#fdfdfa",
	Surface:    "#ffffff",
	Border:     "#e3dcef",
	Ink:        "#1d1f24",
	InkMuted:   "#4f5560",
	Pen:        "#24489a",
	Accent:     "#6f529c",
	AccentSoft: "#efe9f7",
}

var Dark = Palette{
	Name:       "dark",
	Bg:         "#0d1117",
	Surface:    "#161b22",
	Border:     "#3a3150",
	Ink:        "#e6e4ec",
	InkMuted:   "#a7abb6",
	Pen:        "#8fb0ff",
	Accent:     "#c3aef0",
	AccentSoft: "#2a2338",
}

// All はライト、ダークの順に返す。
func All() []Palette { return []Palette{Light, Dark} }

// Contrast は WCAG 2.x のコントラスト比を返す。
func Contrast(fg, bg string) float64 {
	a, b := luminance(fg), luminance(bg)
	if a < b {
		a, b = b, a
	}
	return (a + 0.05) / (b + 0.05)
}

func luminance(hex string) float64 {
	ch := func(i int) float64 {
		v, _ := strconv.ParseUint(hex[i:i+2], 16, 8)
		c := float64(v) / 255
		if c <= 0.03928 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}
	return 0.2126*ch(1) + 0.7152*ch(3) + 0.0722*ch(5)
}

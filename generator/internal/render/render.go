// Package render は README に載せる SVG を描く。
package render

import (
	"bytes"
	"embed"
	"encoding/xml"
	"fmt"
	"math"
	"strings"
	"text/template"

	"github.com/wisteriahuman/wisteriahuman/generator/internal/theme"
)

// Font は SVG 内の文字に使うフォントの並び。GitHub は画像内で Web フォントを読めない。
const Font = `Inter, 'Hiragino Sans', 'Noto Sans CJK JP', system-ui, sans-serif`

type Banner struct{ Name, Handle, Affiliation, Tagline string }

type Heading struct{ Title string }

// Card の Body は最大3行。\n で改行する。
type Card struct{ Title, Subtitle, Body, Meta string }

type MiniCard struct{ Title, Meta string }

type LinkCard struct{ Label, Detail string }

type AtCoder struct {
	User                       string
	Current, Highest, Contests int
	Color                      string
	History                    []int
}

type Langs struct{ Items []LangItem }

type LangItem struct {
	Name    string
	Percent float64
	Color   string
}

type Articles struct{ Items []ArticleItem }

// ArticleItem の Date は 2006-01-02。
type ArticleItem struct{ Title, Date string }

//go:embed templates/*.svg.tmpl
var files embed.FS

var tmpl = template.Must(template.New("").Funcs(template.FuncMap{
	"x":     escape,
	"trunc": Truncate,
	"lines": func(s string, max int) []string {
		ls := strings.Split(s, "\n")
		if len(ls) > max {
			ls = ls[:max]
		}
		return ls
	},
	"add": func(a, b float64) float64 { return a + b },
	"mul": func(a, b float64) float64 { return a * b },
	"f":   func(v float64) string { return fmt.Sprintf("%.1f", v) },
}).ParseFS(files, "templates/*.svg.tmpl"))

type view struct {
	P    theme.Palette
	D    any
	L    any // テンプレートで計算しづらい座標
	Font string
}

// Render はテンプレート name を配色 p とデータで描く。
func Render(name string, p theme.Palette, data any) ([]byte, error) {
	v := view{P: p, D: data, Font: Font}
	switch d := data.(type) {
	case Banner:
		v.L = wisteriaClusters()
	case AtCoder:
		v.L = atcoderLayout(d)
	case Langs:
		v.L = langsLayout(d)
	case Articles:
		v.L = articlesLayout(d)
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name+".svg.tmpl", v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Truncate は rune 数が max を超えたら max-1 文字に「…」を足す。
func Truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}

func escape(s string) string {
	var b strings.Builder
	xml.EscapeText(&b, []byte(s))
	return b.String()
}

type petal struct{ X, Y, R, Opacity float64 }

// wisteriaClusters はバナー右側に垂れ下がる藤の房を返す。
func wisteriaClusters() []petal {
	var ps []petal
	for i, c := range []struct{ x, n float64 }{{630, 9}, {700, 12}, {770, 8}} {
		for j := 0.0; j < c.n; j++ {
			// 上ほど広く、下へ細くなる房。段ごとに少し揺らす
			half := 16 * (1 - j/c.n)
			sway := math.Sin(j*0.8+float64(i)) * 4
			r := 6.5 - j*0.3
			op := 0.95 - j*0.06
			ps = append(ps, petal{X: c.x + sway, Y: 22 + j*13, R: r, Opacity: op})
			if half > 5 {
				ps = append(ps,
					petal{X: c.x + sway - half, Y: 22 + j*13 - 3, R: r * 0.85, Opacity: op * 0.8},
					petal{X: c.x + sway + half, Y: 22 + j*13 - 3, R: r * 0.85, Opacity: op * 0.8})
			}
		}
	}
	return ps
}

type atcoderView struct {
	Points string
	LastX  float64
	LastY  float64
	Grid   []float64
}

func atcoderLayout(d AtCoder) atcoderView {
	const x0, x1, y0, y1 = 320.0, 800.0, 40.0, 180.0
	top := math.Ceil(float64(d.Highest+1)/400) * 400
	var v atcoderView
	for g := 0.0; g <= top; g += 400 {
		v.Grid = append(v.Grid, y1-(g/top)*(y1-y0))
	}
	var pts []string
	for i, r := range d.History {
		x := x0
		if len(d.History) > 1 {
			x = x0 + float64(i)/float64(len(d.History)-1)*(x1-x0)
		}
		y := y1 - float64(r)/top*(y1-y0)
		pts = append(pts, fmt.Sprintf("%.1f,%.1f", x, y))
		v.LastX, v.LastY = x, y
	}
	v.Points = strings.Join(pts, " ")
	return v
}

type segment struct {
	X, W  float64
	Color string
}

type legend struct {
	X, Y float64
	Item LangItem
}

type langsView struct {
	Height   float64
	Segments []segment
	Legend   []legend
}

func langsLayout(d Langs) langsView {
	const x0, width = 24.0, 792.0
	var v langsView
	x := x0
	for i, it := range d.Items {
		w := width * it.Percent / 100
		v.Segments = append(v.Segments, segment{X: x, W: w, Color: it.Color})
		x += w
		v.Legend = append(v.Legend, legend{X: x0 + float64(i%4)*198, Y: 76 + float64(i/4)*30, Item: it})
	}
	rows := (len(d.Items) + 3) / 4
	v.Height = 76 + float64(rows-1)*30 + 28
	return v
}

type articleRow struct {
	Y     float64
	Item  ArticleItem
	Title string
}

type articlesView struct {
	Height float64
	Rows   []articleRow
}

func articlesLayout(d Articles) articlesView {
	var v articlesView
	for i, it := range d.Items {
		v.Rows = append(v.Rows, articleRow{Y: 40 + float64(i)*48, Item: it, Title: Truncate(it.Title, 34)})
	}
	v.Height = 40 + float64(len(d.Items)-1)*48 + 28
	return v
}

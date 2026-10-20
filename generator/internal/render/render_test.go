package render

import (
	"bytes"
	"encoding/xml"
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/wisteriahuman/wisteriahuman/generator/internal/theme"
)

var update = flag.Bool("update", false, "ゴールデンファイルを書き換える")

// samples は全テンプレートの代表データ。
func samples() map[string]any {
	return map[string]any{
		"banner":     Banner{Name: "藤井 優気", Handle: "wisteria", Affiliation: "公立はこだて未来大学 情報アーキテクチャ学科 2年", Tagline: "バックエンドと競技プログラミングが好き"},
		"heading":    Heading{Title: "つくったもの"},
		"subheading": Heading{Title: "Neovim プラグイン"},
		"card":       Card{Title: "リフレクト新世界", Subtitle: "P2HACKS2025", Body: "worker を担当：AI 自動投稿の\n確率制御とプロンプト管理", Meta: "TypeScript"},
		"minicard":   MiniCard{Title: "hoverchain.nvim", Meta: "Lua"},
		"linkcard":   LinkCard{Label: "サイト", Detail: "wisteria-io.com"},
		"atcoder":    AtCoder{User: "wisteriahuman", Current: 1351, Highest: 1351, Contests: 47, Color: "#00C0C0", History: []int{100, 400, 800, 1100, 1351}},
		"langs": Langs{Items: []LangItem{
			{Name: "Go", Percent: 40, Color: "#00ADD8"},
			{Name: "TypeScript", Percent: 35, Color: "#3178c6"},
			{Name: "その他", Percent: 25, Color: "#8b8b8b"},
		}},
		"articles": Articles{Items: []ArticleItem{
			{Title: "エラーは値である", Date: "2026-10-01"},
			{Title: "Neovim 入門 その3", Date: "2026-09-20"},
		}},
	}
}

func render(t *testing.T, name string, p theme.Palette, data any) []byte {
	t.Helper()
	out, err := Render(name, p, data)
	if err != nil {
		t.Fatalf("Render(%s, %s): %v", name, p.Name, err)
	}
	return out
}

func wellFormed(t *testing.T, label string, b []byte) {
	t.Helper()
	d := xml.NewDecoder(bytes.NewReader(b))
	for {
		_, err := d.Token()
		if err != nil {
			if err.Error() == "EOF" {
				return
			}
			t.Fatalf("%s: XML として不正: %v\n%s", label, err, b)
		}
	}
}

func TestRenderAllTemplatesBothPalettes(t *testing.T) {
	for name, data := range samples() {
		for _, p := range theme.All() {
			out := render(t, name, p, data)
			if !bytes.HasPrefix(bytes.TrimSpace(out), []byte("<svg")) {
				t.Errorf("%s/%s: <svg で始まらない", name, p.Name)
			}
			wellFormed(t, name+"/"+p.Name, out)
		}
	}
}

func TestNoExternalReferences(t *testing.T) {
	xmlns := regexp.MustCompile(`xmlns(:\w+)?="[^"]*"`)
	for name, data := range samples() {
		for _, p := range theme.All() {
			s := xmlns.ReplaceAllString(string(render(t, name, p, data)), "")
			for _, bad := range []string{"http", "@import", "<script", "<image"} {
				if strings.Contains(s, bad) {
					t.Errorf("%s/%s: %q を含む", name, p.Name, bad)
				}
			}
		}
	}
}

func TestEscapesSpecialChars(t *testing.T) {
	out := render(t, "card", theme.Light, Card{Title: `A & B <C> "D"`, Subtitle: "s", Body: "b", Meta: "m"})
	wellFormed(t, "card", out)
	if !bytes.Contains(out, []byte("&amp;")) {
		t.Errorf("& がエスケープされていない:\n%s", out)
	}
}

func TestTruncate(t *testing.T) {
	if got := Truncate("あいうえおかきくけこ", 5); got != "あいうえ…" {
		t.Errorf("Truncate = %q, want あいうえ…", got)
	}
	if got := Truncate("あいう", 5); got != "あいう" {
		t.Errorf("Truncate = %q, want あいう", got)
	}
}

func TestArticleTitleIsTruncated(t *testing.T) {
	long := strings.Repeat("長", 60)
	out := string(render(t, "articles", theme.Light, Articles{Items: []ArticleItem{{Title: long, Date: "2026-10-01"}}}))
	m := regexp.MustCompile(`長+…?`).FindString(out)
	if n := utf8.RuneCountInString(m); n == 0 || n > 34 {
		t.Errorf("タイトルが %d rune（34 以内のはず）", n)
	}
}

func TestGolden(t *testing.T) {
	for name, data := range samples() {
		for _, p := range theme.All() {
			out := render(t, name, p, data)
			path := filepath.Join("testdata", name+"-"+p.Name+".golden.svg")
			if *update {
				if err := os.WriteFile(path, out, 0o644); err != nil {
					t.Fatal(err)
				}
				continue
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("%s: %v（-update で作成）", path, err)
			}
			if !bytes.Equal(out, want) {
				t.Errorf("%s とずれている（意図した変更なら -update）", path)
			}
		}
	}
}

func TestSubheadingRenders(t *testing.T) {
	for _, p := range theme.All() {
		out := render(t, "subheading", p, Heading{Title: "Neovim プラグイン"})
		wellFormed(t, "subheading/"+p.Name, out)
		if !bytes.Contains(out, []byte("Neovim プラグイン")) {
			t.Error("見出しの文字が無い")
		}
	}
}

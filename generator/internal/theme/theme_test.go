package theme

import (
	"math"
	"testing"
)

func TestLightUsesSiteTokens(t *testing.T) {
	want := map[string][2]string{
		"Bg":       {Light.Bg, "#fdfdfa"},
		"Ink":      {Light.Ink, "#1d1f24"},
		"InkMuted": {Light.InkMuted, "#4f5560"},
		"Pen":      {Light.Pen, "#24489a"},
	}
	for name, v := range want {
		if v[0] != v[1] {
			t.Errorf("Light.%s = %q, want %q", name, v[0], v[1])
		}
	}
}

func TestContrastKnownValues(t *testing.T) {
	if got := Contrast("#000000", "#ffffff"); math.Abs(got-21) > 0.01 {
		t.Errorf("black on white = %v, want 21", got)
	}
	if got := Contrast("#7b5ea7", "#7b5ea7"); math.Abs(got-1) > 0.0001 {
		t.Errorf("same color = %v, want 1", got)
	}
}

func TestTextColorsMeetAA(t *testing.T) {
	for _, p := range All() {
		texts := map[string]string{"Ink": p.Ink, "InkMuted": p.InkMuted, "Pen": p.Pen, "Accent": p.Accent}
		for _, bg := range []string{p.Bg, p.Surface} {
			for name, fg := range texts {
				if c := Contrast(fg, bg); c < 4.5 {
					t.Errorf("%s: %s %s on %s = %.2f, want >= 4.5", p.Name, name, fg, bg, c)
				}
			}
		}
	}
}

package feed

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const sample = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"><channel><title>wisteria-io</title>
<item><title>B</title><link>https://wisteria-io.com/blog/b/</link><pubDate>Tue, 02 Sep 2025 00:00:00 GMT</pubDate></item>
<item><title>D</title><link>https://wisteria-io.com/blog/d/</link><pubDate>Thu, 02 Oct 2025 00:00:00 +0900</pubDate></item>
<item><title>A</title><link>https://wisteria-io.com/blog/a/</link><pubDate>Fri, 01 Aug 2025 00:00:00 GMT</pubDate></item>
<item><title>C</title><link>https://wisteria-io.com/blog/c/</link><pubDate>Mon, 15 Sep 2025 00:00:00 GMT</pubDate></item>
</channel></rss>`

func TestParseSortsAndLimits(t *testing.T) {
	got, err := Parse(strings.NewReader(sample), 3)
	if err != nil {
		t.Fatal(err)
	}
	var titles []string
	for _, it := range got {
		titles = append(titles, it.Title)
	}
	if strings.Join(titles, ",") != "D,C,B" {
		t.Errorf("titles = %v, want D,C,B", titles)
	}
	if got[0].Link != "https://wisteria-io.com/blog/d/" {
		t.Errorf("link = %s", got[0].Link)
	}
}

func TestParseEntities(t *testing.T) {
	in := `<rss><channel><item><title>A &amp; B</title><link>x</link><pubDate>Fri, 01 Aug 2025 00:00:00 GMT</pubDate></item></channel></rss>`
	got, err := Parse(strings.NewReader(in), 3)
	if err != nil || got[0].Title != "A & B" {
		t.Errorf("got %+v, %v", got, err)
	}
}

func TestParseEmpty(t *testing.T) {
	if _, err := Parse(strings.NewReader(`<rss><channel></channel></rss>`), 3); err == nil {
		t.Error("0件はエラーになるはず")
	}
}

func TestParseBadDate(t *testing.T) {
	in := `<rss><channel><item><title>A</title><link>x</link><pubDate>昨日</pubDate></item></channel></rss>`
	if _, err := Parse(strings.NewReader(in), 3); err == nil {
		t.Error("読めない日付はエラーになるはず")
	}
}

func TestFetch404(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()
	if _, err := Fetch(context.Background(), srv.Client(), srv.URL+"/rss.xml", 3); err == nil {
		t.Error("404 はエラーになるはず")
	}
}

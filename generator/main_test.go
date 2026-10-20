package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wisteriahuman/wisteriahuman/generator/internal/atcoder"
	"github.com/wisteriahuman/wisteriahuman/generator/internal/feed"
)

var staticNames = []string{
	"banner", "h-works", "h-tools", "h-atcoder", "h-langs", "h-articles", "h-links",
	"card-reflect", "card-kunekune", "card-wisteria-io", "card-nix-config",
	"h-nvim", "mini-colab", "mini-yajiuma", "mini-hoverchain", "mini-yt",
	"link-site", "link-x", "link-atcoder",
}

func okDeps() Deps {
	return Deps{
		AtCoder: func(context.Context) (atcoder.Summary, error) {
			return atcoder.Summary{Current: 1351, Highest: 1351, Contests: 47, History: []int{5, 1351}}, nil
		},
		Langs: func(context.Context) (map[string]int64, error) {
			return map[string]int64{"Go": 100, "Lua": 50}, nil
		},
		Feed: func(context.Context) ([]feed.Item, error) {
			return []feed.Item{{Title: "エラーは値である", Link: "x", Published: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)}}, nil
		},
	}
}

func exists(t *testing.T, dir, name string) bool {
	t.Helper()
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}

func count(t *testing.T, dir string) int {
	t.Helper()
	es, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	return len(es)
}

func TestOfflineWritesStaticOnly(t *testing.T) {
	dir := t.TempDir()
	if err := run(context.Background(), Config{Out: dir, Offline: true}, okDeps()); err != nil {
		t.Fatal(err)
	}
	for _, n := range staticNames {
		for _, v := range []string{"light", "dark"} {
			if !exists(t, dir, n+"-"+v+".svg") {
				t.Errorf("%s-%s.svg が無い", n, v)
			}
		}
	}
	if got := count(t, dir); got != len(staticNames)*2 {
		t.Errorf("ファイル数 = %d, want %d", got, len(staticNames)*2)
	}
}

func TestFailedFetchKeepsPreviousFile(t *testing.T) {
	dir := t.TempDir()
	old := filepath.Join(dir, "atcoder-light.svg")
	os.WriteFile(old, []byte("OLD"), 0o644)
	deps := okDeps()
	deps.AtCoder = func(context.Context) (atcoder.Summary, error) { return atcoder.Summary{}, errors.New("down") }
	err := run(context.Background(), Config{Out: dir}, deps)
	if err == nil {
		t.Error("取得失敗はエラーを返すはず")
	}
	if b, _ := os.ReadFile(old); string(b) != "OLD" {
		t.Errorf("atcoder-light.svg が上書きされた: %q", b)
	}
	if !exists(t, dir, "langs-light.svg") {
		t.Error("成功した langs は書かれるはず")
	}
}

func TestEmptyLangsIsError(t *testing.T) {
	deps := okDeps()
	deps.Langs = func(context.Context) (map[string]int64, error) { return map[string]int64{"HTML": 10}, nil }
	dir := t.TempDir()
	if err := run(context.Background(), Config{Out: dir}, deps); err == nil {
		t.Error("集計できる言語が無いときはエラーのはず")
	}
	if exists(t, dir, "langs-light.svg") {
		t.Error("空の langs は書かないはず")
	}
}

func TestNoFeedSkipsArticles(t *testing.T) {
	dir := t.TempDir()
	deps := okDeps()
	deps.Feed = nil
	if err := run(context.Background(), Config{Out: dir}, deps); err != nil {
		t.Fatal(err)
	}
	if exists(t, dir, "articles-light.svg") {
		t.Error("feed が無いときは articles を作らないはず")
	}
}

func TestAllFetchesSucceed(t *testing.T) {
	dir := t.TempDir()
	if err := run(context.Background(), Config{Out: dir}, okDeps()); err != nil {
		t.Fatal(err)
	}
	if got := count(t, dir); got != 44 {
		t.Errorf("ファイル数 = %d, want 44", got)
	}
}

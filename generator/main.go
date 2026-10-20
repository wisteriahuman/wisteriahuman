// Command generator はプロフィール README の SVG を assets/ に書き出す。
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/wisteriahuman/wisteriahuman/generator/internal/atcoder"
	"github.com/wisteriahuman/wisteriahuman/generator/internal/feed"
	"github.com/wisteriahuman/wisteriahuman/generator/internal/langs"
	"github.com/wisteriahuman/wisteriahuman/generator/internal/render"
	"github.com/wisteriahuman/wisteriahuman/generator/internal/theme"
)

type Config struct {
	Out     string
	Offline bool   // 静的な画像だけ書く
	AtCoder string // カードに出す AtCoder のユーザー名
}

// Deps はデータの取得。Feed が nil なら記事は作らない（RSS 公開前）。
type Deps struct {
	AtCoder func(context.Context) (atcoder.Summary, error)
	Langs   func(context.Context) (map[string]int64, error)
	Feed    func(context.Context) ([]feed.Item, error)
}

func main() {
	out := flag.String("out", "../assets", "書き出し先")
	offline := flag.Bool("offline", false, "静的な画像だけ書く")
	feedURL := flag.String("feed", "", "記事の RSS（空なら記事を作らない）")
	team := flag.String("team", "team-repos.json", "チーム作品の設定")
	user := flag.String("user", "wisteriahuman", "GitHub ユーザー")
	atcoderUser := flag.String("atcoder", "wisteriahuman", "AtCoder ユーザー")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	hc := &http.Client{Timeout: 30 * time.Second}
	gh := langs.Client{HTTP: hc, BaseURL: "https://api.github.com", Token: os.Getenv("GITHUB_TOKEN")}

	deps := Deps{
		AtCoder: func(ctx context.Context) (atcoder.Summary, error) {
			return atcoder.Fetch(ctx, hc, "https://atcoder.jp", *atcoderUser)
		},
		Langs: func(ctx context.Context) (map[string]int64, error) {
			total, err := gh.OwnRepoBytes(ctx, *user)
			if err != nil {
				return nil, err
			}
			repos, err := langs.LoadTeamRepos(*team)
			if err != nil {
				return nil, err
			}
			for _, r := range repos {
				b, err := gh.TeamRepoBytes(ctx, r)
				if err != nil {
					return nil, err
				}
				for l, n := range b {
					total[l] += n
				}
			}
			return total, nil
		},
	}
	if *feedURL != "" {
		deps.Feed = func(ctx context.Context) ([]feed.Item, error) { return feed.Fetch(ctx, hc, *feedURL, 3) }
	}

	if err := run(ctx, Config{Out: *out, Offline: *offline, AtCoder: *atcoderUser}, deps); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run は画像を書き出す。取得に失敗した画像は前回のまま残し、エラーをまとめて返す。
func run(ctx context.Context, cfg Config, deps Deps) error {
	if err := os.MkdirAll(cfg.Out, 0o755); err != nil {
		return err
	}
	var errs []error
	for _, img := range staticImages {
		errs = append(errs, write(cfg.Out, img))
	}
	if cfg.Offline {
		return errors.Join(errs...)
	}

	if s, err := deps.AtCoder(ctx); err != nil {
		errs = append(errs, err)
	} else {
		errs = append(errs, write(cfg.Out, image{"atcoder", "atcoder", render.AtCoder{
			User: cfg.AtCoder, Current: s.Current, Highest: s.Highest, Contests: s.Contests,
			Color: atcoder.Color(s.Current), History: s.History,
		}}))
	}

	if total, err := deps.Langs(ctx); err != nil {
		errs = append(errs, err)
	} else if shares := langs.Shares(total, 8); len(shares) == 0 {
		errs = append(errs, langs.ErrEmpty)
	} else {
		var items []render.LangItem
		for _, s := range shares {
			items = append(items, render.LangItem{Name: s.Name, Percent: s.Percent, Color: langs.ColorOf(s.Name)})
		}
		errs = append(errs, write(cfg.Out, image{"langs", "langs", render.Langs{Items: items}}))
	}

	if deps.Feed != nil {
		if items, err := deps.Feed(ctx); err != nil {
			errs = append(errs, err)
		} else {
			var rows []render.ArticleItem
			for _, it := range items {
				rows = append(rows, render.ArticleItem{Title: it.Title, Date: it.Published.Format(time.DateOnly)})
			}
			errs = append(errs, write(cfg.Out, image{"articles", "articles", render.Articles{Items: rows}}))
		}
	}
	return errors.Join(errs...)
}

// write はライトとダークを両方描けたときだけ、一時ファイル経由で書き出す。
func write(dir string, img image) error {
	outs := map[string][]byte{}
	for _, p := range theme.All() {
		b, err := render.Render(img.Tmpl, p, img.Data)
		if err != nil {
			return fmt.Errorf("%s: %w", img.Name, err)
		}
		outs[filepath.Join(dir, img.Name+"-"+p.Name+".svg")] = b
	}
	for _, path := range slices.Sorted(maps.Keys(outs)) {
		tmp := path + ".tmp"
		if err := os.WriteFile(tmp, outs[path], 0o644); err != nil {
			return err
		}
		if err := os.Rename(tmp, path); err != nil {
			return err
		}
	}
	return nil
}

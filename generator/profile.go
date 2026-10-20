package main

import "github.com/wisteriahuman/wisteriahuman/generator/internal/render"

// image は書き出す1枚分。ファイル名は Name-light.svg / Name-dark.svg になる。
type image struct {
	Name string
	Tmpl string
	Data any
}

// staticImages は README に載せる、取得の要らない画像。
var staticImages = []image{
	{"banner", "banner", render.Banner{
		Name: "藤井 優気", Handle: "wisteria",
		Affiliation: "公立はこだて未来大学 情報アーキテクチャ学科 2年",
		Tagline:     "バックエンドと競技プログラミングが好き",
	}},
	{"h-works", "heading", render.Heading{Title: "つくったもの"}},
	{"h-tools", "heading", render.Heading{Title: "道具づくり"}},
	{"h-atcoder", "heading", render.Heading{Title: "競技プログラミング"}},
	{"h-langs", "heading", render.Heading{Title: "よく書く言語"}},
	{"h-articles", "heading", render.Heading{Title: "最近書いた記事"}},
	{"h-links", "heading", render.Heading{Title: "リンク"}},
	{"card-reflect", "card", render.Card{
		Title: "リフレクト新世界", Subtitle: "P2HACKS2025", Meta: "TypeScript",
		Body: "worker を担当：AI 自動投稿の\n確率制御とプロンプト管理、\n日次・週次ジョブ、自動デプロイ",
	}},
	{"card-kunekune", "card", render.Card{
		Title: "パキッと クネクネタスク", Subtitle: "JPHACKS", Meta: "TypeScript",
		Body: "タスク API のユースケース・\nリポジトリ層と、AI タスク\n分割 API の仕様を担当",
	}},
	{"card-wisteria-io", "card", render.Card{
		Title: "wisteria-io", Subtitle: "ポートフォリオとブログ", Meta: "Astro",
		Body: "自宅の Mac mini から\nCloudflare Tunnel で配信。\n判断は ADR に記録",
	}},
	{"card-nix-config", "card", render.Card{
		Title: "nix-config", Subtitle: "開発環境", Meta: "Nix",
		Body: "home-manager で役割ごとに\n環境を管理。curl 一発で\n新しいマシンを構築",
	}},
	{"h-nvim", "subheading", render.Heading{Title: "Neovim プラグイン"}},
	{"mini-colab", "minicard", render.MiniCard{Title: "colab.nvim", Meta: "Python"}},
	{"mini-yajiuma", "minicard", render.MiniCard{Title: "yajiuma.nvim", Meta: "Lua"}},
	{"mini-hoverchain", "minicard", render.MiniCard{Title: "hoverchain.nvim", Meta: "Lua"}},
	{"mini-yt", "minicard", render.MiniCard{Title: "yt.nvim", Meta: "Lua"}},
	{"link-site", "linkcard", render.LinkCard{Label: "サイト", Detail: "wisteria-io.com"}},
	{"link-x", "linkcard", render.LinkCard{Label: "X", Detail: "@wisteria_man"}},
	{"link-atcoder", "linkcard", render.LinkCard{Label: "AtCoder", Detail: "wisteriahuman"}},
}

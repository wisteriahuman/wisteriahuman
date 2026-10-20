# プロフィールREADME 刷新 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** README の画像をすべて自作 SVG（ライト/ダーク）にし、Go の生成器と毎日のワークフローで AtCoder・言語比率・ブログ新着を最新に保つ。

**Architecture:** `generator/` は Go の単一モジュール。データ取得（`atcoder` / `langs` / `feed`）と描画（`render`）を分け、`main.go` が両者をつないで `assets/<name>-{light,dark}.svg` を書く。静的な SVG（バナー・見出し・カード）はネットなしで生成でき、動的な SVG は取得に成功したときだけ上書きする。README は `<picture>` で `assets/` を参照するだけ。

**Tech Stack:** Go 1.26（標準ライブラリのみ。`text/template`, `encoding/xml`, `encoding/json`, `net/http`）、GitHub Actions。

**Spec:** `docs/superpowers/specs/2026-10-21-profile-readme-design.md`

## Global Constraints

- 文言は日本語のみ。名前は「藤井 優気 / wisteria」、所属は「公立はこだて未来大学 情報アーキテクチャ学科 2年」
- 土台色（ライト）: 紙 `#fdfdfa`、インク `#1d1f24`、インク弱 `#4f5560`、ペン `#24489a`。差し色は藤色
- すべての SVG はライト版とダーク版の2枚。README では `<picture>` + `prefers-color-scheme` で切り替える
- SVG は外部リソース（Web フォント・外部画像・`<script>`）を一切参照しない
- `font-family="Inter, 'Hiragino Sans', 'Noto Sans CJK JP', system-ui, sans-serif"`。文字のパス化はしない
- 文字と背景のコントラストは WCAG AA（4.5:1 以上）
- 言語比率から除外: `Jupyter Notebook`, `HTML`, `TeX`, `CSS`
- 取得に失敗した SVG は上書きしない（前回のまま）。いずれかが失敗したらプロセスは終了コード 1
- 外部依存パッケージは追加しない（`go.mod` に require なし）
- このリポジトリでは commit / push はユーザーが行う。各タスクの最後は `git add` までとし、コミットメッセージ案を提示する

## Review Focus

- AtCoder の JSON が空配列・HTML（メンテナンス画面）・rated な参加が0件 → エラーを返し、前回の SVG を残す（Task 3 でテスト）
- GitHub API が 100件を超えるリポジトリを返す・200以外（レート制限 403）を返す → ページングして全件、非200はエラー（Task 4 でテスト）
- チーム作品の Trees API が `truncated: true`、または指定ディレクトリに該当ファイルが0件 → エラー（設定の打ち間違いに気づけるように）（Task 4 でテスト）
- 記事タイトルや説明に `&` `<` `"` が含まれる → SVG が壊れずエスケープされる（Task 2 でテスト）
- 長い日本語タイトル → 規定の文字数で「…」に切り詰め、カード幅からはみ出さない（Task 2 でテスト）

---

## ファイル構成

```
generator/
  go.mod                      module github.com/wisteriahuman/wisteriahuman/generator, go 1.26
  main.go                     フラグ解析・取得・描画・書き出しのオーケストレーション
  main_test.go
  team-repos.json             チーム作品のリポジトリと担当ディレクトリ
  internal/theme/theme.go     パレット（Light/Dark）、コントラスト計算
  internal/render/            SVG 描画（text/template を go:embed）
    render.go                 共通: Render(name, palette, data) と文字列ヘルパ
    templates/*.svg.tmpl
    testdata/*.golden.svg
  internal/atcoder/atcoder.go 履歴 JSON の取得・解析・要約
  internal/langs/langs.go     自分のリポジトリ + チーム作品の言語集計
  internal/feed/feed.go       RSS の取得・解析
assets/                       生成物（*-light.svg / *-dark.svg）
.github/workflows/update-assets.yml
README.md
```

---

### Task 1: テーマとプロジェクトの土台

**Files:**
- Create: `generator/go.mod`, `generator/internal/theme/theme.go`
- Test: `generator/internal/theme/theme_test.go`

**Interfaces:**
- Produces:
  - `type Palette struct { Name string; Bg, Surface, Border, Ink, InkMuted, Pen, Accent, AccentSoft string }`（色は `#rrggbb`。`Name` は `"light"` / `"dark"`）
  - `var Light, Dark Palette`、`func All() []Palette`（Light, Dark の順）
  - `func Contrast(fg, bg string) float64`（WCAG 2.x の相対輝度による比）

- [ ] **Step 1: 失敗するテストを書く**

```go
func TestLightUsesSiteTokens(t *testing.T) {
	// Bg=#fdfdfa, Ink=#1d1f24, InkMuted=#4f5560, Pen=#24489a
}
func TestContrastKnownValues(t *testing.T) {
	// Contrast("#000000", "#ffffff") ≈ 21.0（誤差 0.01）、同色は 1.0
}
func TestTextColorsMeetAA(t *testing.T) {
	// 全 Palette で Ink / InkMuted / Pen / Accent の Bg と Surface に対する Contrast >= 4.5
}
```

- [ ] **Step 2: `cd generator && go test ./...` を実行し、未定義で FAIL することを確認**

- [ ] **Step 3: `theme.go` を実装**

Light の Surface / Border / Accent（藤色）/ AccentSoft と Dark 全体の値は、`TestTextColorsMeetAA` を満たすように実装者が選ぶ。目安: Accent は `#7b5ea7` 前後、Dark の Bg は GitHub ダーク（`#0d1117`）になじむ濃さ。

- [ ] **Step 4: `go test ./...` が PASS**

- [ ] **Step 5: ステージしてコミット案を提示**

`git add generator/` / 案: `feat(generator): add theme palettes with contrast checks`

---

### Task 2: SVG 描画（静的な画像一式 + 動的な画像のテンプレート）

**Files:**
- Create: `generator/internal/render/render.go`, `generator/internal/render/templates/*.svg.tmpl`
- Test: `generator/internal/render/render_test.go`, `generator/internal/render/testdata/*.golden.svg`

**Interfaces:**
- Consumes: `theme.Palette`
- Produces:
  - `func Render(tmpl string, p theme.Palette, data any) ([]byte, error)`（テンプレート名は下表）
  - `func Truncate(s string, max int) string`（rune 数で数え、超えたら `max-1` 文字 + `…`）
  - データ型（`render` パッケージ内）:
    - `Banner{ Name, Handle, Affiliation, Tagline string }`
    - `Heading{ Title string }`
    - `Card{ Title, Subtitle, Body, Meta string }`（Body は最大3行。`\n` で改行）
    - `MiniCard{ Title, Meta string }`
    - `LinkCard{ Label, Detail string }`
    - `AtCoder{ User string; Current, Highest, Contests int; Color string; History []int }`
    - `Langs{ Items []LangItem }`、`LangItem{ Name string; Percent float64; Color string }`
    - `Articles{ Items []ArticleItem }`、`ArticleItem{ Title, Date string }`（Date は `2006-01-02`）

| テンプレート | 幅×高さ(px) | 用途 |
|---|---|---|
| `banner` | 840×200 | ヘッダー |
| `heading` | 840×56 | セクション見出し（藤色の飾り付き、左寄せ） |
| `card` | 410×170 | つくったもの・道具づくり |
| `minicard` | 200×72 | Neovim プラグイン |
| `linkcard` | 270×80 | リンク |
| `atcoder` | 840×220 | レートカード（現在値・最高値・参加回数・レート色の帯・折れ線） |
| `langs` | 840×可変 | 積み上げ横棒 + 凡例（上位8言語、残りは「その他」） |
| `articles` | 840×可変 | 新着3件（日付 + タイトル） |

- [ ] **Step 1: 失敗するテストを書く**

```go
func TestRenderAllTemplatesBothPalettes(t *testing.T)  // 全テンプレート × theme.All() で err == nil、出力が "<svg" で始まり xml.Unmarshal で壊れていない
func TestNoExternalReferences(t *testing.T)            // 全出力に "http" "@import" "<script" "<image" が含まれない（xmlns 属性は除く）
func TestEscapesSpecialChars(t *testing.T)             // Card{Title: `A & B <C> "D"`} → 出力が XML として妥当、"&amp;" を含む
func TestTruncate(t *testing.T)                        // Truncate("あいうえおかきくけこ", 5) == "あいうえ…"、Truncate("あいう", 5) == "あいう"
func TestArticleTitleIsTruncated(t *testing.T)         // 60文字のタイトル → 出力のタイトルは 34 rune 以内
func TestGolden(t *testing.T)                          // 各テンプレート × light/dark を testdata/<tmpl>-<palette>.golden.svg と比較。-update フラグで書き換え
```

- [ ] **Step 2: `go test ./internal/render/` が FAIL することを確認**

- [ ] **Step 3: `render.go` とテンプレートを実装**

テンプレートは `//go:embed templates/*.svg.tmpl` で埋め込む。`text/template` に XML エスケープ関数（`encoding/xml.EscapeText` を包む）を `x` として登録し、文字列はすべて `{{x .Title}}` で出す。フォントと色は Global Constraints のとおり。角丸・余白は Meltype に寄せる（rx 12〜16、内側の余白 24）。記事タイトルは 34 rune、カード本文は1行 24 rune を上限にする。

- [ ] **Step 4: `go test ./internal/render/ -update` でゴールデンを作り、出力を目視確認してから `go test ./internal/render/` が PASS**

目視確認: 生成した SVG をブラウザで開き、ライト/ダークとも文字がはみ出さないこと。

- [ ] **Step 5: ステージしてコミット案を提示**

案: `feat(generator): render SVG templates for banner, cards and stats`

---

### Task 3: AtCoder の取得と要約

**Files:**
- Create: `generator/internal/atcoder/atcoder.go`
- Test: `generator/internal/atcoder/atcoder_test.go`, `generator/internal/atcoder/testdata/history.json`（実データの一部）

**Interfaces:**
- Produces:
  - `type Summary struct { Current, Highest, Contests int; History []int }`（History は rated の NewRating を古い順）
  - `func Parse(r io.Reader) (Summary, error)`
  - `func Fetch(ctx context.Context, c *http.Client, baseURL, user string) (Summary, error)`（`baseURL + "/users/" + user + "/history/json"`。本番の baseURL は `https://atcoder.jp`）
  - `func Color(rating int) string`

- [ ] **Step 1: 失敗するテストを書く**

```go
func TestParseSummary(t *testing.T)        // testdata: IsRated=false を含む3件 → Contests は rated の数、Current は最後の NewRating、Highest は最大
func TestParseNoRated(t *testing.T)        // [] と rated 0件 → err != nil
func TestParseHTML(t *testing.T)           // "<html>..." → err != nil
func TestColor(t *testing.T)               // 399→#808080, 400→#804000, 800→#008000, 1351→#00C0C0, 1600→#0000FF, 2000→#C0C000, 2400→#FF8000, 2800→#FF0000
func TestFetchNon200(t *testing.T)         // httptest で 503 → err != nil
```

- [ ] **Step 2: FAIL を確認** — `go test ./internal/atcoder/`

- [ ] **Step 3: 実装**（JSON のフィールドは `IsRated bool`, `NewRating int`）

- [ ] **Step 4: PASS を確認**

- [ ] **Step 5: ステージしてコミット案を提示** — 案: `feat(generator): fetch and summarize AtCoder rating history`

---

### Task 4: 言語の集計

**Files:**
- Create: `generator/internal/langs/langs.go`, `generator/team-repos.json`
- Test: `generator/internal/langs/langs_test.go`

**Interfaces:**
- Produces:
  - `type TeamRepo struct { Repo string `json:"repo"`; Paths []string `json:"paths"` }`（Repo は `owner/name`、Paths は前方一致するディレクトリ）
  - `func LoadTeamRepos(path string) ([]TeamRepo, error)`
  - `type Client struct { HTTP *http.Client; BaseURL, Token string }`（本番の BaseURL は `https://api.github.com`。Token が空なら Authorization を付けない）
  - `func (c Client) OwnRepoBytes(ctx context.Context, user string) (map[string]int64, error)` — `GET /users/{user}/repos?type=owner&per_page=100` を `Link: rel="next"` で全ページ取得。`fork` と `private` を除き、各リポジトリの `GET /repos/{o}/{r}/languages` を合算
  - `func (c Client) TeamRepoBytes(ctx context.Context, t TeamRepo) (map[string]int64, error)` — `GET /repos/{o}/{r}/git/trees/HEAD?recursive=1` の blob を Paths で絞り、拡張子から言語を判定して size を合算
  - `func LanguageOf(path string) (string, bool)`
  - `type Share struct { Name string; Bytes int64; Percent float64 }`
  - `func Shares(total map[string]int64, top int) []Share` — 除外言語を落とし、降順で上位 `top`、残りは `Name: "その他"` にまとめる。Percent の合計は 100
  - `func ColorOf(lang string) string`（linguist の色。未知は `#8b8b8b`）

`team-repos.json` の初期値:

```json
[
  { "repo": "p2hacks2025/pre-03", "paths": ["apps/worker/", "apps/api/"] },
  { "repo": "jphacks/sp_2502", "paths": ["web/src/server/"] }
]
```

拡張子の対応: `.go` Go / `.ts` `.tsx` `.mts` `.cts` TypeScript / `.js` `.jsx` `.mjs` `.cjs` JavaScript / `.py` Python / `.swift` Swift / `.lua` Lua / `.rs` Rust / `.rb` Ruby / `.sh` `.bash` `.zsh` Shell / `.nix` Nix / `.java` Java / `.kt` Kotlin / `.c` `.h` C / `.cc` `.cpp` `.hpp` C++ / `.svelte` Svelte / `.astro` Astro / `.sql` SQL。それ以外（png・md・json など）は数えない。

- [ ] **Step 1: 失敗するテストを書く**（GitHub API は `httptest.Server` で偽装）

```go
func TestOwnRepoBytesPaginatesAndSkipsForks(t *testing.T) // 2ページ（Link ヘッダ）、fork 1件・private 1件を含む → 残りの languages だけ合算
func TestOwnRepoBytesNon200(t *testing.T)                 // 403 → err != nil
func TestTeamRepoBytesFiltersPaths(t *testing.T)          // apps/worker/a.ts(100), apps/worker/x.png(9999), apps/native/b.swift(50) → {"TypeScript":100}
func TestTeamRepoBytesTruncated(t *testing.T)             // truncated:true → err != nil
func TestTeamRepoBytesNoMatch(t *testing.T)               // Paths に一致する対象ファイル0件 → err != nil
func TestSharesExcludesAndGroups(t *testing.T)            // HTML/CSS/TeX/Jupyter Notebook を除外、top=2 で3言語目以降は「その他」、Percent 合計 100±0.01
func TestLoadTeamRepos(t *testing.T)                      // 実ファイル team-repos.json を読めて2件
```

- [ ] **Step 2: FAIL を確認** — `go test ./internal/langs/`
- [ ] **Step 3: 実装**
- [ ] **Step 4: PASS を確認**
- [ ] **Step 5: ステージしてコミット案を提示** — 案: `feat(generator): aggregate language bytes from own and team repos`

---

### Task 5: RSS の取得

**Files:**
- Create: `generator/internal/feed/feed.go`
- Test: `generator/internal/feed/feed_test.go`

**Interfaces:**
- Produces:
  - `type Item struct { Title, Link string; Published time.Time }`
  - `func Parse(r io.Reader, n int) ([]Item, error)` — RSS 2.0 の `channel/item`（`title` / `link` / `pubDate` は RFC1123Z と RFC1123 の両方を受ける）を新しい順に最大 n 件。0件はエラー
  - `func Fetch(ctx context.Context, c *http.Client, url string, n int) ([]Item, error)`

- [ ] **Step 1: 失敗するテストを書く**

```go
func TestParseSortsAndLimits(t *testing.T)  // 順不同の4件 → 新しい順の3件
func TestParseEntities(t *testing.T)        // <title>A &amp; B</title> → "A & B"
func TestParseEmpty(t *testing.T)           // item 0件 → err != nil
func TestFetch404(t *testing.T)             // 404 → err != nil
```

- [ ] **Step 2–4: FAIL → 実装 → PASS**（`go test ./internal/feed/`）
- [ ] **Step 5: ステージしてコミット案を提示** — 案: `feat(generator): parse RSS feed for latest articles`

---

### Task 6: main（オーケストレーションと書き出し）

**Files:**
- Create: `generator/main.go`, `generator/profile.go`（README に載せる静的データ）
- Test: `generator/main_test.go`

**Interfaces:**
- Consumes: Task 1〜5 のすべて
- Produces（CLI）: `go run . [-out ../assets] [-offline] [-feed URL] [-team team-repos.json] [-user wisteriahuman] [-atcoder wisteriahuman]`。環境変数 `GITHUB_TOKEN` を Client.Token に使う
  - `-offline`: 静的な画像だけ書く
  - `-feed` が空なら記事は生成しない（RSS 公開前の状態）
- 出力ファイル名（すべて `-light.svg` / `-dark.svg` の2枚）:
  `banner`, `h-works`, `h-tools`, `h-atcoder`, `h-langs`, `h-articles`, `h-links`, `card-reflect`, `card-kunekune`, `card-wisteria-io`, `card-nix-config`, `mini-colab`, `mini-yajiuma`, `mini-hoverchain`, `mini-yt`, `link-site`, `link-x`, `link-atcoder`, `atcoder`, `langs`, `articles`
- `func run(ctx context.Context, cfg Config, deps Deps) error` を main から呼ぶ（テストで差し替えるため）。`Deps` は AtCoder / langs / feed の取得関数を持つ

`profile.go` の文言（このまま使う。Body は1行24 rune 以内に収めて改行する）:

| 名前 | Title / Subtitle | Body | Meta |
|---|---|---|---|
| banner | 藤井 優気 / wisteria | — | 公立はこだて未来大学 情報アーキテクチャ学科 2年 ／ バックエンドと競技プログラミングが好き |
| card-reflect | リフレクト新世界 / P2HACKS2025 | worker を担当：AI 自動投稿の確率制御とプロンプト管理、日次・週次ジョブ、自動デプロイ | TypeScript |
| card-kunekune | パキッと クネクネタスク / JPHACKS | タスク API のユースケース・リポジトリ層と、AI タスク分割 API の仕様を担当 | TypeScript |
| card-wisteria-io | wisteria-io / ポートフォリオとブログ | 自宅の Mac mini から Cloudflare Tunnel で配信。判断は ADR に記録 | Astro |
| card-nix-config | nix-config / 開発環境 | home-manager で役割ごとに環境を管理。curl 一発で新しいマシンを構築 | Nix |
| mini-* | colab.nvim / yajiuma.nvim / hoverchain.nvim / yt.nvim | — | Python / Lua / Lua / Lua |
| link-* | サイト / X / AtCoder | wisteria-io.com / @wisteria_man / wisteriahuman | — |

見出し: つくったもの / 道具づくり / 競技プログラミング / よく書く言語 / 最近書いた記事 / リンク。

書き出しは一時ファイル → rename。取得に失敗した画像は書かず、エラーをまとめて最後に返す（他の画像は書く）。

- [ ] **Step 1: 失敗するテストを書く**

```go
func TestOfflineWritesStaticOnly(t *testing.T)         // t.TempDir に -offline → banner-light.svg 等の静的 18種×2 があり、atcoder-*.svg / langs-*.svg / articles-*.svg が無い
func TestFailedFetchKeepsPreviousFile(t *testing.T)    // 既存 atcoder-light.svg に "OLD" → AtCoder 取得がエラー → ファイルは "OLD" のまま、langs は新しく書かれ、run は err を返す
func TestNoFeedSkipsArticles(t *testing.T)             // Feed 空 → articles-*.svg を作らず、err == nil
func TestAllFetchesSucceed(t *testing.T)               // 偽の Deps 全成功 → 21種×2 = 42 ファイル、err == nil
```

- [ ] **Step 2–4: FAIL → 実装 → PASS**（`go test ./...`）

- [ ] **Step 5: 実データで生成して確認**

Run: `cd generator && GITHUB_TOKEN=$(gh auth token) go run . -out ../assets`
Expected: 終了コード 0、`assets/` に 40 ファイル（articles 以外）。AtCoder カードに 1351、言語比率に Go・Lua・TypeScript などが並び、HTML / TeX / Jupyter が無いこと。

- [ ] **Step 6: ステージしてコミット案を提示** — `git add generator/ assets/` / 案: `feat(generator): generate profile SVG assets`

---

### Task 7: README の書き換え

**Files:**
- Modify: `README.md`（全面書き換え）

各画像は次の形。見出しは `width="100%"`、カードは2枚並び（`width="49%"`）、ミニカードは4枚並び（`width="24%"`）、リンクカードは3枚並び（`width="32%"`）。すべて `<p align="center">` で中央寄せ。

```html
<a href="https://github.com/p2hacks2025/pre-03">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="assets/card-reflect-dark.svg">
    <img src="assets/card-reflect-light.svg" alt="リフレクト新世界（P2HACKS2025）：worker を担当" width="49%">
  </picture>
</a>
```

リンク先: card-reflect → `https://github.com/p2hacks2025/pre-03`、card-kunekune → `https://github.com/jphacks/sp_2502`、card-wisteria-io → `https://wisteria-io.com`、card-nix-config → `https://github.com/wisteriahuman/nix-config`、mini-* → `https://github.com/wisteriahuman/<name>.nvim`、link-site → `https://wisteria-io.com`、link-x → `https://x.com/wisteria_man`、link-atcoder と atcoder → `https://atcoder.jp/users/wisteriahuman`。

構成順: banner → つくったもの → 道具づくり → 競技プログラミング → よく書く言語 → リンク。**記事セクションは入れない**（RSS 公開後に Task 9 で追加）。旧 github-readme-stats の `<img>` は削除。`alt` は画像内の文字を読み上げられる日本語にする。

- [ ] **Step 1: 書き換える**
- [ ] **Step 2: 検証** — `grep -c 'vercel.app' README.md` が 0。README が参照する `assets/*.svg` がすべて存在する:

```bash
grep -oE 'assets/[a-z0-9-]+\.svg' README.md | sort -u | xargs -I{} test -f {} && echo OK
```

Expected: `OK`

- [ ] **Step 3: 見た目の確認** — `gh` ではプレビューできないので、ユーザーに push 後の https://github.com/wisteriahuman で、ライト/ダーク両方の表示を確かめてもらう
- [ ] **Step 4: ステージしてコミット案を提示** — 案: `Redesign profile README with generated SVG cards`

---

### Task 8: 自動更新ワークフロー

**Files:**
- Create: `.github/workflows/update-assets.yml`

内容:
- トリガ: `schedule: cron "0 21 * * *"`（毎日 06:00 JST）、`workflow_dispatch`、`push`（`generator/**` の変更時、main のみ）
- `permissions: contents: write`
- 手順: `actions/checkout@v4` → `actions/setup-go@v5`（`go-version-file: generator/go.mod`）→ `go test ./...` → `go run . -out ../assets`（`GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}`。`-feed` は付けない）→ `assets/` に差分があるときだけ `github-actions[bot]` として `chore: update generated assets` をコミットして push
- 生成ステップが失敗しても、書けた分は commit する（`if: always()` を commit ステップに付け、最後にジョブを失敗させる）
- `concurrency: update-assets` で同時実行を防ぐ

- [ ] **Step 1: 書く**
- [ ] **Step 2: 検証** — `actionlint` があれば `actionlint .github/workflows/update-assets.yml` がエラーなし。なければ `python3 -c "import yaml,sys;yaml.safe_load(open(sys.argv[1]))" .github/workflows/update-assets.yml`
- [ ] **Step 3: ステージしてコミット案を提示** — 案: `ci: regenerate profile assets daily`
- [ ] **Step 4: push 後の確認（ユーザー）** — Actions タブから手動実行し、成功すること・差分がなければコミットされないことを確認

---

### Task 9: 記事セクションの有効化（wisteria-io の RSS 公開後）

前提: wisteria-io 側で `https://wisteria-io.com/rss.xml` が公開されていること（wisteria-io は任せるプロジェクトなので、そちらのフローで別途実施。約束は spec の「付随作業」）。

**Files:**
- Modify: `.github/workflows/update-assets.yml`（`-feed https://wisteria-io.com/rss.xml` を追加）
- Modify: `README.md`（よく書く言語 と リンク の間に h-articles + articles を追加。リンク先は `https://wisteria-io.com/blog/`）

- [ ] **Step 1: `curl -s -o /dev/null -w '%{http_code}' https://wisteria-io.com/rss.xml` が 200**
- [ ] **Step 2: `cd generator && go run . -out ../assets -feed https://wisteria-io.com/rss.xml` で articles-*.svg が生成され、新着3件が並ぶ**
- [ ] **Step 3: README とワークフローを更新し、Task 7 Step 2 の検証コマンドが OK**
- [ ] **Step 4: ステージしてコミット案を提示** — 案: `Show latest blog posts on profile README`

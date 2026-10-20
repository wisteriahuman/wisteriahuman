# プロフィールREADME 刷新 設計

## 意図

GitHub プロフィール（`wisteriahuman/wisteriahuman` の README.md）を、インターン・選考の担当者が30秒で「何をした人か」を掴める状態にする。従の読者はエンジニア仲間・競プロ勢。

現状の問題:

- 中身が自己紹介1行とリンク3つで、「バックエンドと競プロが好き」を裏付ける実物がない
- 言語比率画像（自前ホストの github-readme-stats）が `Something went wrong!` を返し、壊れて見えている

見た目の参考は [yksr-melt/Meltype の README](https://github.com/yksr-melt/Meltype)。見出し・カード・バナーを自作 SVG にし、色を揃え、中央寄せと余白で整える作り方を取り入れる。

## ゴール

- 上から読むだけで「作ったもの → それを支える道具 → 地力（競プロ）→ 継続（記事）」が伝わる
- 外部サービスの停止で画像が壊れることが二度とない
- AtCoder・言語比率・ブログ新着が人手なしで最新に保たれる
- ライト/ダーク両方で自然に見える

## 非ゴール

- 英語版
- ピン留めの変更（README の直下に表示されるので、README は文脈の説明に徹し、リポジトリの羅列はピン留めに任せる）
- 受賞歴の記載
- 草・コミット数などの活動グラフ（プロフィール標準の草と重複）

## アンチゴール（動いていても失敗）

- 取得失敗時にエラー画像・空の画像・壊れた画像がコミットされる
- README の見た目が閲覧時に外部サービスへ依存する
- ダークモードで白い板が浮く
- 文字が環境依存フォントで大きく崩れ、はみ出す

## 決定事項（ユーザー合意）

### 文言・名前

- 日本語のみ。「藤井 優気 / wisteria」と本名とハンドルを併記
- 所属: 公立はこだて未来大学 情報アーキテクチャ学科 2年

### 配色

- 土台は wisteria-io.com のトークン: 紙 `#fdfdfa`、インク `#1d1f24`、インク弱 `#4f5560`、ペン `#24489a`
- 差し色に藤色
- すべての SVG にライト版とダーク版を作り、`<picture>` + `prefers-color-scheme` で切り替える

### 構成（上から順）

1. **ヘッダーバナー**: 名前（本名 / ハンドル）、所属、一言（バックエンドと競プロが好き）
2. **つくったもの**（カード）
   - P2HACKS2025「リフレクト新世界」（p2hacks2025/pre-03、TypeScript）: バックエンドの worker 担当。AI 自動投稿（投稿頻度に応じた確率制御・プロンプト管理）、日次/週次ジョブ、worker の自動デプロイ、タイムラインへの AI 投稿注入 API
   - JPHACKS「パキッと クネクネタスク」（jphacks/sp_2502、https://sp-2502.vercel.app ）: タスク API のユースケース層・リポジトリ層、AI タスク分割 API の仕様書
   - wisteria-io（https://wisteria-io.com ）: Astro 製ポートフォリオ兼ブログ。自宅 Mac mini から Cloudflare Tunnel で配信、self-hosted runner でデプロイ、判断を ADR で記録、Lighthouse CI で検証
3. **道具づくり**（カード）
   - nix-config: home-manager で役割（role）単位に環境を管理、curl 一発で新マシンを構築
   - Neovim プラグイン群を1枠: colab.nvim / yajiuma.nvim / hoverchain.nvim / yt.nvim
4. **競プロ**: AtCoder レートカード（現在値・最高値・レート色・推移の簡易グラフ）。ユーザー `wisteriahuman`
5. **よく書く言語**: 言語比率（今の壊れた画像の置き換え）
6. **最近書いた記事**: wisteria-io の新着3件（タイトル・日付、記事へリンク）
7. **リンク**: サイト（wisteria-io.com）、X（@wisteria_man）、AtCoder をカードで

カードはそれぞれ該当リポジトリ・URL へリンクする（`<a>` で SVG 画像を包む）。

Neovim プラグイン群の枠（私の判断）: 1枚の画像につけられるリンクは1つだけなので、見出し付きの枠の中に4つの小さなカード SVG を並べ、それぞれのリポジトリへリンクする。

### 自動生成

- このリポジトリに `generator/`（Go、標準ライブラリ中心）を置く
- `.github/workflows/` で毎日1回 + 手動実行（`workflow_dispatch`）。生成した SVG を `assets/` にコミットし、差分がなければコミットしない
- README が参照するのはリポジトリ内の `assets/*.svg` だけ。閲覧時に外部へ依存しない
- 言語比率の集計対象（Jupyter Notebook / HTML / TeX / CSS は除外）:
  - 自分の非 fork 公開リポジトリ: API から自動で拾い、リポジトリごとの言語バイト数を合算
  - チーム作品: `repositoriesContributedTo` は直近の貢献しか返さず取りこぼす・余計なものが混ざるため使わない。設定ファイルに「リポジトリ + 自分が担当したディレクトリ」を列挙し、生成時にそのディレクトリだけを集計する。初期値は `p2hacks2025/pre-03` の `apps/worker`（と `apps/api`）、`jphacks/sp_2502` の `web/src/server`。ハッカソンに出るたびに1行足す運用
  - 2つの単位をそろえるため、チーム作品の集計もバイト数で行い、言語の判定は拡張子で行う（GitHub の linguist とずれる分は許容する）

### 付随作業（別リポジトリ）

- wisteria-io に `@astrojs/rss` で RSS を追加する。wisteria-io は「任せるプロジェクト」なので、そちらのフローで別途進める
- この設計が wisteria-io に求める約束: `https://wisteria-io.com/rss.xml` で、`draft: false` の記事を `pubDate` の新しい順に、`title` / `link` / `pubDate` / `description` 付きで返す（スキーマは `src/content/config.ts` の blog コレクション）

### 運用

- このリポジトリは任せるプロジェクトではない。ユーザーの手での commit/push は、ユーザーが行う。ワークフローによる `assets/` の自動コミットは合意済み

## 制約

- GitHub は README の画像を `<img>` として描画するため、SVG 内から Web フォントや外部リソースを読み込めない。文字は閲覧者の環境のフォントで描かれる
- GitHub は画像をカメラプロキシ経由でキャッシュする。同じパスでも更新が反映されるまで少し遅れることがある
- ワークフローは `GITHUB_TOKEN` で動かす（公開情報だけを扱う）。`contents: write` が必要
- AtCoder の履歴は `https://atcoder.jp/users/wisteriahuman/history/json`（生成時にだけ取得）

## 実装者に任せること（私の判断として記す）

- **藤色とダーク版の具体的な色値**: 土台トークンから導き、文字と背景のコントラストは WCAG AA を満たす
- **SVG の文字**: `font-family` は Inter / Hiragino Sans / Noto Sans CJK JP / system-ui の順に並べる。どのフォールバックでもはみ出さない幅に余裕を持たせる。文字のパス化はしない
- **静的な SVG も生成器で作る**: バナー・見出し・カードもテンプレートとデータ（Go の構造体か `generator/` 内のデータファイル）から生成し、色や寸法を一箇所で管理する。データ取得が要らない静的な SVG はネット接続なしで生成できるようにする
- **取得失敗時**: 該当する SVG は上書きせず前回のまま残し、ワークフローは失敗として終了する（気づけるように）。部分的な失敗で他の SVG の更新を止めるかは実装者の判断
- **RSS がまだ無い間**: 記事セクションの SVG を作らず、README からも外しておく。RSS が公開されたら有効にする（実装順は RSS → 記事セクション）
- **チーム作品のディレクトリの集計方法**: 浅いクローン、sparse-checkout、Trees API のどれを使うか
- **テスト**: パース（AtCoder JSON・RSS・言語集計）は単体テストを、SVG はゴールデンファイルでの比較を想定
- **ファイル名・ディレクトリ構成・スケジュールの時刻**

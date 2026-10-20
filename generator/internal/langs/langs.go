// Package langs は自分のリポジトリとチーム作品から、よく書く言語の比率を集計する。
package langs

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
)

// excluded は比率に含めない言語。
var excluded = map[string]bool{"Jupyter Notebook": true, "HTML": true, "TeX": true, "CSS": true}

// TeamRepo はチーム作品のうち自分が担当したディレクトリ。Paths は前方一致。
type TeamRepo struct {
	Repo  string   `json:"repo"`
	Paths []string `json:"paths"`
}

func LoadTeamRepos(file string) ([]TeamRepo, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var ts []TeamRepo
	if err := json.Unmarshal(b, &ts); err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	return ts, nil
}

// Client は GitHub API（本番の BaseURL は https://api.github.com）を呼ぶ。
type Client struct {
	HTTP    *http.Client
	BaseURL string
	Token   string
}

func (c Client) get(ctx context.Context, url string, v any) (http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, res.Status)
	}
	return res.Header, json.NewDecoder(res.Body).Decode(v)
}

var nextLink = regexp.MustCompile(`<([^>]+)>;\s*rel="next"`)

// OwnRepoBytes は user の非 fork 公開リポジトリの言語バイト数を合算する。
func (c Client) OwnRepoBytes(ctx context.Context, user string) (map[string]int64, error) {
	total := map[string]int64{}
	url := c.BaseURL + "/users/" + user + "/repos?type=owner&per_page=100"
	for url != "" {
		var repos []struct {
			FullName string `json:"full_name"`
			Fork     bool   `json:"fork"`
			Private  bool   `json:"private"`
		}
		h, err := c.get(ctx, url, &repos)
		if err != nil {
			return nil, err
		}
		for _, r := range repos {
			if r.Fork || r.Private {
				continue
			}
			var ls map[string]int64
			if _, err := c.get(ctx, c.BaseURL+"/repos/"+r.FullName+"/languages", &ls); err != nil {
				return nil, err
			}
			for l, n := range ls {
				total[l] += n
			}
		}
		url = ""
		if m := nextLink.FindStringSubmatch(h.Get("Link")); m != nil {
			url = m[1]
		}
	}
	return total, nil
}

// TeamRepoBytes は t.Paths 配下のファイルを拡張子で言語に分け、バイト数を合算する。
func (c Client) TeamRepoBytes(ctx context.Context, t TeamRepo) (map[string]int64, error) {
	var tree struct {
		Truncated bool `json:"truncated"`
		Tree      []struct {
			Path string `json:"path"`
			Type string `json:"type"`
			Size int64  `json:"size"`
		} `json:"tree"`
	}
	if _, err := c.get(ctx, c.BaseURL+"/repos/"+t.Repo+"/git/trees/HEAD?recursive=1", &tree); err != nil {
		return nil, err
	}
	if tree.Truncated {
		return nil, fmt.Errorf("%s: ツリーが大きすぎて途中で切れている", t.Repo)
	}
	total := map[string]int64{}
	for _, p := range t.Paths {
		found := false
		for _, e := range tree.Tree {
			if e.Type != "blob" || !strings.HasPrefix(e.Path, p) {
				continue
			}
			if l, ok := LanguageOf(e.Path); ok {
				total[l] += e.Size
				found = true
			}
		}
		// パスごとに確かめ、1つの打ち間違いも見逃さない
		if !found {
			return nil, fmt.Errorf("%s: %s に集計できるファイルが無い（設定を確認）", t.Repo, p)
		}
	}
	return total, nil
}

var extensions = map[string]string{
	".go": "Go",
	".ts": "TypeScript", ".tsx": "TypeScript", ".mts": "TypeScript", ".cts": "TypeScript",
	".js": "JavaScript", ".jsx": "JavaScript", ".mjs": "JavaScript", ".cjs": "JavaScript",
	".py": "Python", ".swift": "Swift", ".lua": "Lua", ".rs": "Rust", ".rb": "Ruby",
	".sh": "Shell", ".bash": "Shell", ".zsh": "Shell",
	".nix": "Nix", ".java": "Java", ".kt": "Kotlin",
	".c": "C", ".h": "C", ".cc": "C++", ".cpp": "C++", ".hpp": "C++",
	".svelte": "Svelte", ".astro": "Astro", ".sql": "SQL",
}

func LanguageOf(p string) (string, bool) {
	l, ok := extensions[strings.ToLower(path.Ext(p))]
	return l, ok
}

type Share struct {
	Name    string
	Bytes   int64
	Percent float64
}

// Shares は除外言語を落とし、多い順に上位 top 件、残りを「その他」にまとめる。
func Shares(total map[string]int64, top int) []Share {
	var ss []Share
	var sum int64
	for l, n := range total {
		if excluded[l] || n <= 0 {
			continue
		}
		ss = append(ss, Share{Name: l, Bytes: n})
		sum += n
	}
	slices.SortFunc(ss, func(a, b Share) int {
		return cmp.Or(cmp.Compare(b.Bytes, a.Bytes), cmp.Compare(a.Name, b.Name))
	})
	if len(ss) > top {
		other := Share{Name: "その他"}
		for _, s := range ss[top:] {
			other.Bytes += s.Bytes
		}
		ss = append(ss[:top], other)
	}
	for i := range ss {
		ss[i].Percent = float64(ss[i].Bytes) / float64(sum) * 100
	}
	return ss
}

// ErrEmpty は集計できる言語が1つも無いとき。
var ErrEmpty = errors.New("langs: 集計できる言語が無い")

// colors は GitHub linguist の色。Lua の #000080 はダーク背景で見えないため明るくしている。
var colors = map[string]string{
	"Go": "#00ADD8", "TypeScript": "#3178c6", "JavaScript": "#f1e05a", "Python": "#3572A5",
	"Swift": "#F05138", "Lua": "#4f5fd0", "Rust": "#dea584", "Ruby": "#701516",
	"Shell": "#89e051", "Nix": "#7e7eff", "Java": "#b07219", "Kotlin": "#A97BFF",
	"C": "#555555", "C++": "#f34b7d", "Svelte": "#ff3e00", "Astro": "#ff5a03", "SQL": "#e38c00",
	"Dockerfile": "#384d54", "Vim Script": "#199f4b", "Makefile": "#427819",
}

func ColorOf(lang string) string {
	if c, ok := colors[lang]; ok {
		return c
	}
	return "#8b8b8b"
}

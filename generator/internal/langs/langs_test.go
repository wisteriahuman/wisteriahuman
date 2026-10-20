package langs

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func serve(t *testing.T, routes map[string]func(w http.ResponseWriter, r *http.Request)) Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Path
		if r.URL.RawQuery != "" {
			key += "?" + r.URL.RawQuery
		}
		if h, ok := routes[key]; ok {
			h(w, r)
			return
		}
		t.Logf("unexpected request %s", key)
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	return Client{HTTP: srv.Client(), BaseURL: srv.URL, Token: "tok"}
}

func text(s string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, s) }
}

func TestOwnRepoBytesPaginatesAndSkipsForks(t *testing.T) {
	var c Client
	c = serve(t, map[string]func(http.ResponseWriter, *http.Request){
		"/users/me/repos?type=owner&per_page=100": func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer tok" {
				t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
			}
			w.Header().Set("Link", fmt.Sprintf(`<%s/users/me/repos?type=owner&per_page=100&page=2>; rel="next"`, c.BaseURL))
			fmt.Fprint(w, `[{"full_name":"me/a","fork":false,"private":false},{"full_name":"me/forked","fork":true,"private":false}]`)
		},
		"/users/me/repos?type=owner&per_page=100&page=2": text(`[{"full_name":"me/b","fork":false,"private":false},{"full_name":"me/secret","fork":false,"private":true}]`),
		"/repos/me/a/languages":                          text(`{"Go":100,"HTML":50}`),
		"/repos/me/b/languages":                          text(`{"Go":10,"Lua":30}`),
	})
	got, err := c.OwnRepoBytes(context.Background(), "me")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]int64{"Go": 110, "HTML": 50, "Lua": 30}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestOwnRepoBytesNon200(t *testing.T) {
	c := serve(t, map[string]func(http.ResponseWriter, *http.Request){
		"/users/me/repos?type=owner&per_page=100": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, `{"message":"API rate limit exceeded"}`)
		},
	})
	if _, err := c.OwnRepoBytes(context.Background(), "me"); err == nil {
		t.Error("403 はエラーになるはず")
	}
}

const treePath = "/repos/org/team/git/trees/HEAD?recursive=1"

func TestTeamRepoBytesFiltersPaths(t *testing.T) {
	c := serve(t, map[string]func(http.ResponseWriter, *http.Request){
		treePath: text(`{"truncated":false,"tree":[
			{"path":"apps/worker","type":"tree"},
			{"path":"apps/worker/a.ts","type":"blob","size":100},
			{"path":"apps/worker/x.png","type":"blob","size":9999},
			{"path":"apps/workers/c.ts","type":"blob","size":7},
			{"path":"apps/native/b.swift","type":"blob","size":50}]}`),
	})
	got, err := c.TeamRepoBytes(context.Background(), TeamRepo{Repo: "org/team", Paths: []string{"apps/worker/"}})
	if err != nil {
		t.Fatal(err)
	}
	if want := map[string]int64{"TypeScript": 100}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestTeamRepoBytesTruncated(t *testing.T) {
	c := serve(t, map[string]func(http.ResponseWriter, *http.Request){
		treePath: text(`{"truncated":true,"tree":[{"path":"apps/worker/a.ts","type":"blob","size":100}]}`),
	})
	if _, err := c.TeamRepoBytes(context.Background(), TeamRepo{Repo: "org/team", Paths: []string{"apps/worker/"}}); err == nil {
		t.Error("truncated はエラーになるはず")
	}
}

func TestTeamRepoBytesNoMatch(t *testing.T) {
	c := serve(t, map[string]func(http.ResponseWriter, *http.Request){
		treePath: text(`{"truncated":false,"tree":[{"path":"apps/worker/a.png","type":"blob","size":100}]}`),
	})
	if _, err := c.TeamRepoBytes(context.Background(), TeamRepo{Repo: "org/team", Paths: []string{"apps/worker/"}}); err == nil {
		t.Error("対象ファイル0件はエラーになるはず")
	}
}

func TestLanguageOf(t *testing.T) {
	cases := map[string]string{"a/b.go": "Go", "x.tsx": "TypeScript", "x.mjs": "JavaScript", "flake.nix": "Nix", "init.lua": "Lua", "a.cpp": "C++"}
	for p, want := range cases {
		if got, ok := LanguageOf(p); !ok || got != want {
			t.Errorf("LanguageOf(%s) = %s, %v", p, got, ok)
		}
	}
	for _, p := range []string{"logo.png", "README.md", "package.json", "Makefile"} {
		if _, ok := LanguageOf(p); ok {
			t.Errorf("LanguageOf(%s) は対象外のはず", p)
		}
	}
}

func TestSharesExcludesAndGroups(t *testing.T) {
	got := Shares(map[string]int64{"Go": 600, "Lua": 300, "Swift": 60, "Rust": 40, "HTML": 9999, "CSS": 1, "TeX": 1, "Jupyter Notebook": 1}, 2)
	names := []string{}
	sum := 0.0
	for _, s := range got {
		names = append(names, s.Name)
		sum += s.Percent
	}
	if want := []string{"Go", "Lua", "その他"}; !reflect.DeepEqual(names, want) {
		t.Errorf("names = %v, want %v", names, want)
	}
	if math.Abs(sum-100) > 0.01 {
		t.Errorf("Percent 合計 = %v", sum)
	}
	if got[2].Bytes != 100 {
		t.Errorf("その他 = %d, want 100", got[2].Bytes)
	}
}

func TestSharesFewLanguagesHasNoOther(t *testing.T) {
	got := Shares(map[string]int64{"Go": 1, "Lua": 1}, 8)
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
}

func TestLoadTeamRepos(t *testing.T) {
	got, err := LoadTeamRepos("../../team-repos.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Repo != "p2hacks2025/pre-03" {
		t.Errorf("got %+v", got)
	}
}

func TestColorOf(t *testing.T) {
	if ColorOf("Go") != "#00ADD8" || ColorOf("謎の言語") != "#8b8b8b" {
		t.Errorf("Go=%s unknown=%s", ColorOf("Go"), ColorOf("謎の言語"))
	}
}

func TestTeamRepoBytesOnePathNoMatch(t *testing.T) {
	c := serve(t, map[string]func(http.ResponseWriter, *http.Request){
		treePath: text(`{"truncated":false,"tree":[{"path":"apps/worker/a.ts","type":"blob","size":100}]}`),
	})
	_, err := c.TeamRepoBytes(context.Background(), TeamRepo{Repo: "org/team", Paths: []string{"apps/worker/", "apps/apii/"}})
	if err == nil || !strings.Contains(err.Error(), "apps/apii/") {
		t.Errorf("一致しないパスを名指しするエラーのはず: %v", err)
	}
}

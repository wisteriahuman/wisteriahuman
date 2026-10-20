package atcoder

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestParseSummary(t *testing.T) {
	f, err := os.Open("testdata/history.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	got, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	want := Summary{Current: 1300, Highest: 1351, Contests: 3, History: []int{5, 1351, 1300}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Parse = %+v, want %+v", got, want)
	}
}

func TestParseNoRated(t *testing.T) {
	for _, in := range []string{`[]`, `[{"IsRated": false, "NewRating": 0}]`} {
		if _, err := Parse(strings.NewReader(in)); err == nil {
			t.Errorf("Parse(%s): エラーになるはず", in)
		}
	}
}

func TestParseHTML(t *testing.T) {
	if _, err := Parse(strings.NewReader("<html><body>メンテナンス中</body></html>")); err == nil {
		t.Error("HTML はエラーになるはず")
	}
}

func TestColor(t *testing.T) {
	cases := map[int]string{0: "#808080", 399: "#808080", 400: "#804000", 800: "#008000", 1351: "#00C0C0", 1600: "#0000FF", 2000: "#C0C000", 2400: "#FF8000", 2800: "#FF0000", 3500: "#FF0000"}
	for r, want := range cases {
		if got := Color(r); got != want {
			t.Errorf("Color(%d) = %s, want %s", r, got, want)
		}
	}
}

func TestFetchUsesHistoryPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/wisteriahuman/history/json" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(`[{"IsRated": true, "NewRating": 42}]`))
	}))
	defer srv.Close()
	got, err := Fetch(context.Background(), srv.Client(), srv.URL, "wisteriahuman")
	if err != nil || got.Current != 42 {
		t.Errorf("Fetch = %+v, %v", got, err)
	}
}

func TestFetchNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`[{"IsRated": true, "NewRating": 42}]`))
	}))
	defer srv.Close()
	if _, err := Fetch(context.Background(), srv.Client(), srv.URL, "wisteriahuman"); err == nil {
		t.Error("503 はエラーになるはず")
	}
}

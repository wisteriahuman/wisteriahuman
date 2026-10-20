// Package atcoder は AtCoder のコンテスト履歴を取得して要約する。
package atcoder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Summary の History は rated な参加の NewRating を古い順に並べたもの。
type Summary struct {
	Current, Highest, Contests int
	History                    []int
}

type entry struct {
	IsRated   bool
	NewRating int
}

// Parse は /users/{user}/history/json の本文を要約する。rated な参加が無ければエラー。
func Parse(r io.Reader) (Summary, error) {
	var es []entry
	if err := json.NewDecoder(r).Decode(&es); err != nil {
		return Summary{}, fmt.Errorf("atcoder: 履歴を読めない: %w", err)
	}
	var s Summary
	for _, e := range es {
		if !e.IsRated {
			continue
		}
		s.History = append(s.History, e.NewRating)
		s.Current = e.NewRating
		s.Highest = max(s.Highest, e.NewRating)
	}
	s.Contests = len(s.History)
	if s.Contests == 0 {
		return Summary{}, errors.New("atcoder: rated な参加が無い")
	}
	return s, nil
}

// Fetch は baseURL（本番は https://atcoder.jp）から user の履歴を取得する。
func Fetch(ctx context.Context, c *http.Client, baseURL, user string) (Summary, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/users/"+user+"/history/json", nil)
	if err != nil {
		return Summary{}, err
	}
	res, err := c.Do(req)
	if err != nil {
		return Summary{}, fmt.Errorf("atcoder: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return Summary{}, fmt.Errorf("atcoder: %s", res.Status)
	}
	return Parse(res.Body)
}

// Color はレートに対応する AtCoder の色を返す。
func Color(rating int) string {
	colors := []string{"#808080", "#804000", "#008000", "#00C0C0", "#0000FF", "#C0C000", "#FF8000"}
	if i := rating / 400; i < len(colors) {
		return colors[i]
	}
	return "#FF0000"
}

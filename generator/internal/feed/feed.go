// Package feed は RSS 2.0 から新着記事を読む。
package feed

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"
)

type Item struct {
	Title, Link string
	Published   time.Time
}

// Parse は channel/item を新しい順に最大 n 件返す。0件はエラー。
func Parse(r io.Reader, n int) ([]Item, error) {
	var doc struct {
		Items []struct {
			Title   string `xml:"title"`
			Link    string `xml:"link"`
			PubDate string `xml:"pubDate"`
		} `xml:"channel>item"`
	}
	if err := xml.NewDecoder(r).Decode(&doc); err != nil {
		return nil, fmt.Errorf("feed: %w", err)
	}
	var items []Item
	for _, it := range doc.Items {
		t, err := parseDate(strings.TrimSpace(it.PubDate))
		if err != nil {
			return nil, fmt.Errorf("feed: %q の日付: %w", it.Title, err)
		}
		items = append(items, Item{Title: strings.TrimSpace(it.Title), Link: strings.TrimSpace(it.Link), Published: t})
	}
	if len(items) == 0 {
		return nil, errors.New("feed: 記事が無い")
	}
	slices.SortStableFunc(items, func(a, b Item) int { return b.Published.Compare(a.Published) })
	return items[:min(n, len(items))], nil
}

func parseDate(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC1123Z, s)
	if err != nil {
		return time.Parse(time.RFC1123, s)
	}
	return t, nil
}

func Fetch(ctx context.Context, c *http.Client, url string, n int) ([]Item, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("feed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("feed: %s: %s", url, res.Status)
	}
	return Parse(res.Body, n)
}

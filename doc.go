package main

import (
	"unicode/utf8"
)

// Document defined common document
type Document struct {
	ID        string            `json:"uri"`
	Title     string            `json:"title,omitempty"`
	Summary   string            `json:"summary,omitempty"`
	Content   string            `json:"content,omitempty"`
	Meta      map[string]string `json:"tags,omitempty"`
	CreatedAt int64             `json:"created_at,omitempty"`
}

type hitResult struct {
	ID        string              `json:"uri"`
	Fragments map[string][]string `json:"fragments"`
	Score     float64             `json:"score"`
}

func hasDocument(id string) bool {
	var doc, err = docIndex.Document(id)
	if err != nil {
		return false
	}
	if doc == nil {
		return false
	}
	return true
}

func filterUtf8(old string) string {
	n := old[:]
	for !utf8.ValidString(n) {
		if len(n) <= 0 {
			break
		}
		n = n[:len(n)-1]
	}
	if len(n) < len(old) {
		n = n + "…"
	}

	return n
}

func filterFragment(in []string) []string {
	out := make([]string, len(in))
	for i, f := range in {
		out[i] = filterUtf8(f)
	}
	return out
}

func filterFragments(in map[string][]string) map[string][]string {
	out := make(map[string][]string)
	for k, v := range in {
		out[k] = filterFragment(v)
	}
	return out
}

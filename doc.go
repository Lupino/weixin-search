package main

import (
	"github.com/mholt/binding"
	"net/http"
)

// Document defined common document
type Document struct {
	ID        string `json:"id"`
	Title     string `json:"title,omitempty"`
	Summary   string `json:"summary,omitempty"`
	Content   string `json:"content,omitempty"`
	Tags      string `json:"tags,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"`
}

// FieldMap defined the interface for binding form
func (doc *Document) FieldMap(_ *http.Request) binding.FieldMap {
	return binding.FieldMap{
		&doc.ID:        binding.Field{Form: "id", Required: true},
		&doc.Title:     binding.Field{Form: "title", Required: true},
		&doc.Summary:   binding.Field{Form: "summary", Required: false},
		&doc.Content:   binding.Field{Form: "content", Required: false},
		&doc.Tags:      binding.Field{Form: "tags", Required: false},
		&doc.CreatedAt: binding.Field{Form: "created_at", Required: false},
	}
}

// ResultID defined result document id
type ResultID struct {
	ID    string  `json:"id"`
	Score float64 `json:"score"`
}

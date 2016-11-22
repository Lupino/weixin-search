package main

import (
	"encoding/json"
	"github.com/blevesearch/bleve/document"
	"github.com/mholt/binding"
	"net/http"
)

// Document defined common document
type Document struct {
	ID        string   `json:"uri"`
	Title     string   `json:"title,omitempty"`
	Content   string   `json:"content,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	CreatedAt int64    `json:"created_at,omitempty"`
}

// FieldMap defined the interface for binding form
func (doc *Document) FieldMap(_ *http.Request) binding.FieldMap {
	return binding.FieldMap{
		&doc.ID:        binding.Field{Form: "uri", Required: true},
		&doc.Title:     binding.Field{Form: "title", Required: true},
		&doc.Content:   binding.Field{Form: "content", Required: false},
		&doc.Tags:      binding.Field{Form: "tags", Required: false},
		&doc.CreatedAt: binding.Field{Form: "created_at", Required: false},
	}
}

type hitResult struct {
	ID        string              `json:"uri"`
	Fragments map[string][]string `json:"fragments"`
	Score     float64             `json:"score"`
}

func getDocument(id string) (*Document, error) {
	var doc, err = docIndex.Document(id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}

	var realDoc = new(Document)
	realDoc.ID = doc.ID
	for _, field := range doc.Fields {
		switch field.Name() {
		case "title":
			realDoc.Title = string(field.Value())
			break
		case "content":
			realDoc.Content = string(field.Value())
			break
		case "tags":
			var payload = field.Value()
			json.Unmarshal(payload, &realDoc.Tags)
			break
		case "created_at":
			v, _ := field.(*document.NumericField).Number()
			realDoc.CreatedAt = int64(v)
			break
		}
	}
	return realDoc, nil
}

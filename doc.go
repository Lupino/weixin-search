package main

import (
	"encoding/json"
	// "github.com/blevesearch/bleve/document"
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
	ID      string            `json:"uri"`
	Summary string            `json:"summary"`
	Meta    map[string]string `json:"meta"`
	Score   float64           `json:"score"`
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
		// case "title":
		//     realDoc.Title = string(field.Value())
		//     break
		// case "content":
		//     realDoc.Content = string(field.Value())
		//     break
		case "summary":
			realDoc.Summary = string(field.Value())
			break
		case "meta":
			var payload = field.Value()
			json.Unmarshal(payload, &realDoc.Meta)
			break
			// case "created_at":
			//     v, _ := field.(*document.NumericField).Number()
			//     realDoc.CreatedAt = int64(v)
			//     break
		}
	}
	return realDoc, nil
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

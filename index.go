package main

import (
	"flag"
	"github.com/Lupino/tokenizer"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzers/custom_analyzer"
	"github.com/blevesearch/bleve/analysis/char_filters/html_char_filter"
	"github.com/blevesearch/bleve/index/store/goleveldb"
)

var tokenizerHost = flag.String("tokenizer", "localhost:3000", "tokenizer server host.")

func createMapping() *bleve.IndexMapping {
	mapping, err := newIndexMapping()
	if err != nil {
		panic(err)
	}
	return mapping
}

func openIndex(path string) (index bleve.Index, err error) {
	if index, err = bleve.Open(path); err != nil {
		mapping := createMapping()
		if index, err = bleve.NewUsing(path, mapping, bleve.Config.DefaultIndexType, goleveldb.Name, nil); err != nil {
			return
		}
	}
	return
}

func newIndexMapping() (*bleve.IndexMapping, error) {
	var (
		mapping *bleve.IndexMapping
		err     error
	)
	mapping = bleve.NewIndexMapping()

	if err = mapping.AddCustomTokenizer("sego",
		map[string]interface{}{
			"host": *tokenizerHost,
			"type": tokenizer.Name,
		},
	); err != nil {
		return nil, err
	}

	// create a custom analyzer
	if err = mapping.AddCustomAnalyzer("sego",
		map[string]interface{}{
			"type": custom_analyzer.Name,
			"char_filters": []string{
				html_char_filter.Name,
			},
			"tokenizer": "sego",
			"token_filters": []string{
				"possessive_en",
				"to_lower",
				"stop_en",
			},
		},
	); err != nil {
		return nil, err
	}

	mapping.DefaultAnalyzer = tokenizer.Name
	return mapping, nil
}

package main

import (
	"encoding/json"
	"flag"
	"github.com/Lupino/go-periodic"
)

var (
	funcName = flag.String("funcname", "search-index", "search index funcname.")
)

func submitDoc(doc Document) error {
	var data, _ = json.Marshal(doc)
	return pclient.SubmitJob(*funcName, doc.ID,
		map[string]string{"args": string(data)})
}

func indexDocHandle(job periodic.Job) {
	var doc Document
	if err := json.Unmarshal([]byte(job.Args), &doc); err != nil {
		job.Done()
		return
	}
	if err := docIndex.Index(doc.ID, doc); err != nil {
		job.Fail()
		return
	}
	job.Done()
}

package main

import (
	"encoding/json"
	"github.com/Lupino/go-periodic"
)

var (
	funcName = "patent-search-index"
)

func submitDoc(doc Document) error {
	var data, _ = json.Marshal(doc)
	return pclient.SubmitJob(funcName, doc.ID,
		map[string]string{"args": string(data)})
}

func indexDocHandle(job periodic.Job) {
	var doc Document
	if err := json.Unmarshal([]byte(job.Args), &doc); err != nil {
		job.Done()
	}
	if err := docIndex.Index(doc.ID, doc); err != nil {
		job.Fail()
	}
}

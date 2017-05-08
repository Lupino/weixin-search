package main

import (
	"github.com/Lupino/go-periodic"
)

var funcName = "-crawl-link"

func submitCrawlLink(link string) error {
	return pclient.SubmitJob(name+funcName, link, nil)
}

func crawlLinkHandle(job periodic.Job) {
	var (
		err error
		doc Document
	)

	if err = docIndex.Index(doc.ID, doc); err != nil {
		job.Fail()
		return
	}
	job.Done()
}

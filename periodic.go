package main

import (
	"fmt"
	"github.com/Lupino/go-periodic"
	"strconv"
)

var funcName = "-crawl-link"

func submitCrawlLink(link string) error {
	return pclient.SubmitJob(name+funcName, link, nil)
}

func retryJob(job periodic.Job) {
	if job.Raw.Counter > 20 {
		job.Done()
	} else {
		job.SchedLater(int(job.Raw.Counter)*10, 1)
	}
}

func crawlLinkHandle(job periodic.Job) {
	var (
		err  error
		doc  Document
		meta map[string]string
		art  Article
	)

	if meta, err = doCrawl(job.Name); err != nil {
		fmt.Printf("doCrawl() failed (%s)", err)
		return
	}
	if art, err = createArticle(meta); err != nil {
		fmt.Printf("createArticle() failed (%s)", err)
		retryJob(job)
		return
	}
	if err = createTimeline(meta["biz"], art); err != nil {
		fmt.Printf("createTimeline() failed (%s)", err)
		retryJob(job)
		return
	}
	if err = updateCover(art, meta["cover"]); err != nil {
		fmt.Printf("updateCover() failed (%s)", err)
		retryJob(job)
		return
	}
	meta["id"] = strconv.Itoa(art.ID)
	doc = metaDoc(meta)
	if err = docIndex.Index(doc.ID, doc); err != nil {
		retryJob(job)
		return
	}
	job.Done()
}

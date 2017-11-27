package main

import (
	"github.com/Lupino/go-periodic"
	"log"
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
		log.Printf("doCrawl() failed (%s)\n", err)
		return
	}
	if !validMeta(meta) {
		// log.Printf("Invalid meta (%s)\n", meta)
		retryJob(job)
		return
	}
	if art, err = createArticle(meta); err != nil {
		log.Printf("createArticle() failed (%s)\n", err)
		retryJob(job)
		return
	}
	if err = createTimeline(meta["biz"], art); err != nil {
		log.Printf("createTimeline() failed (%s)\n", err)
		retryJob(job)
		return
	}
	if err = updateCover(art, meta["cover"]); err != nil {
		log.Printf("updateCover() failed (%s)\n", err)
	}
	meta["id"] = strconv.Itoa(art.ID)
	doc = metaDoc(meta)
	if err = docIndex.Index(doc.ID, doc); err != nil {
		retryJob(job)
		return
	}
	job.Done()
}

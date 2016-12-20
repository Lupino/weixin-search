package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Lupino/go-periodic"
	"log"
	"net/http"
	"net/url"
	"strings"
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

func submitHotLink(link string) error {
	return pclient.SubmitJob("hot-"+*funcName, link, nil)
}

func indexHotHandle(job periodic.Job) {
	var (
		form = url.Values{}
		url  string
		req  *http.Request
		rsp  *http.Response
		err  error
		doc  Document
	)

	form.Add("data", job.Name)
	url = fmt.Sprintf("http://%s/api/extract/", extractHost)

	req, _ = http.NewRequest("POST", url, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	if rsp, err = http.DefaultClient.Do(req); err != nil {
		log.Printf("http.DefaultClient.Do() failed (%s)\n", err)
		job.Done()
		return
	}
	defer rsp.Body.Close()

	if int(rsp.StatusCode/100) != 2 {
		job.Done()
		return
	}

	decoder := json.NewDecoder(rsp.Body)
	if err = decoder.Decode(&doc); err != nil {
		log.Printf("json.NewDecoder().Decode() failed (%s)", err)
		job.Done()
		return
	}

	if err := docIndex.Index(doc.ID, doc); err != nil {
		job.Fail()
		return
	}
	job.Done()
}

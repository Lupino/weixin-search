package main

import (
	"github.com/PuerkitoBio/goquery"
	"regexp"
	"strings"
)

func extractData(data string) string {
	var words = strings.Split(data, "||")
	for _, word := range words {
		word = strings.Trim(word, "\" ")
		if len(word) > 0 {
			return word
		}
	}
	return data
}

func doCrawl(uri string) (meta map[string]string, err error) {
	var (
		doc    *goquery.Document
		reMeta = regexp.MustCompile("var (biz|sn|mid|msg_title|msg_desc|msg_cdn_url) = ([^;]+);")
		text   string
		match  []string
		html   string
	)
	meta = make(map[string]string)

	if doc, err = goquery.NewDocument(uri); err != nil {
		return
	}
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		text = s.Text()
		if reMeta.MatchString(text) {
			match = reMeta.FindStringSubmatch(text)
			meta[match[1]] = extractData(match[2])
		}
	})

	contentElement := doc.Find("#js_content")

	contentElement.Find("img").Each(func(i int, s *goquery.Selection) {
		var (
			imgUrl string
			ok     bool
		)

		if imgUrl, ok = s.Attr("data-src"); !ok {
			imgUrl, _ = s.Attr("src")
		}
		s.SetAttr("src", imgUrl)
		s.RemoveAttr("data-src")
		s.RemoveAttr("data-s")
		s.RemoveAttr("data-type")
		s.RemoveAttr("data-ratio")
		s.RemoveAttr("data-w")
	})

	if html, err = contentElement.Html(); err != nil {
		return
	}
	meta["msg_content"] = strings.Trim(html, " \n\r\t")
	return
}

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
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
		}
	})

	if cover, ok := meta["msg_cdn_url"]; ok {
		if file, err := uploadImage(cover, "jpg"); err == nil {
			meta["msg_cdn_url"] = fileUrl(file)
		}
	}

	contentElement := doc.Find("#js_content")

	contentElement.Find("img").Each(func(i int, s *goquery.Selection) {
		var (
			imgUrl string
			ok     bool
			file   File
			tp     string
		)

		if imgUrl, ok = s.Attr("data-src"); !ok {
			imgUrl, _ = s.Attr("src")
		}
		tp = s.AttrOr("data-type", "jpg")
		if file, err = uploadImage(imgUrl, tp); err == nil {
			imgUrl = fileUrl(file)
		}
		s.SetAttr("src", imgUrl)
		s.RemoveAttr("data-src")
		s.RemoveAttr("data-s")
		s.RemoveAttr("data-ratio")
		s.RemoveAttr("data-w")
	})

	if html, err = contentElement.Html(); err != nil {
		return
	}
	meta["msg_content"] = strings.Trim(html, " \n\r\t")
	return
}

// {"extra":{"height":296,"size":23903,"width":640,"name":"abc.jpg","type":"image/jpeg","ext":"jpg"},"bucket":"data/33c6615d7aaf88ff2ad1","key":"261645BACE6E14F328DEEE46B6057DECE95482F7","created_at":1494222465,"id":2}
type Extra struct {
	Ext string `json:"ext"`
}

type File struct {
	Key   string `json:"key"`
	Extra Extra  `json:"extra"`
}

func fileUrl(file File) string {
	return articleHost + "/" + file.Key + "." + file.Extra.Ext + "?key=" + articleKey
}

func uploadImage(imgUrl, tp string) (file File, err error) {
	var (
		req    *http.Request
		rsp    *http.Response
		imgRsp *http.Response
		raw    []byte
		url    = articleHost + "/api/upload/?fileName=abc." + tp
	)

	if imgRsp, err = http.Get(imgUrl); err != nil {
		return
	}
	if imgRsp.StatusCode != 200 {
		err = errors.New("Request image failed.")
		return
	}
	if raw, err = ioutil.ReadAll(imgRsp.Body); err != nil {
		return
	}
	defer imgRsp.Body.Close()
	if req, err = http.NewRequest("PUT", url, bytes.NewReader(raw)); err != nil {
		return
	}
	if err = filledRequestHeaderWithRaw(req); err != nil {
		return
	}
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		return
	}
	if rsp.StatusCode != 200 {
		err = errors.New("Upload image failed.")
		return
	}
	decoder := json.NewDecoder(rsp.Body)
	if err = decoder.Decode(&file); err != nil {
		return
	}
	return
}

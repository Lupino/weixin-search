package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	urlLib "net/url"
	"regexp"
	"strconv"
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
		reMeta = regexp.MustCompile("var (biz|sn|mid|msg_title|msg_desc|msg_cdn_url|svr_time) = ([^;]+);")
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
		lines := strings.Split(text, "\n")
		for _, line := range lines {
			if reMeta.MatchString(line) {
				match = reMeta.FindStringSubmatch(line)
				meta[match[1]] = extractData(match[2])
			}
		}
	})

	if cover, ok := meta["msg_cdn_url"]; ok {
		if file, err := uploadImage(cover, "jpg"); err == nil {
			meta["msg_cdn_url"] = fileUrl(file)
			meta["cover"] = strconv.Itoa(file.ID)
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
	meta["msg_text"] = strings.Trim(contentElement.Text(), " \n\r\t")
	return
}

// {"extra":{"height":296,"size":23903,"width":640,"name":"abc.jpg","type":"image/jpeg","ext":"jpg"},"bucket":"data/33c6615d7aaf88ff2ad1","key":"261645BACE6E14F328DEEE46B6057DECE95482F7","created_at":1494222465,"id":2}
type Extra struct {
	Ext string `json:"ext"`
}

type File struct {
	ID    int    `json:"id"`
	Key   string `json:"key"`
	Extra Extra  `json:"extra"`
}

func fileUrl(file File) string {
	return articleHost + "/" + file.Key + "." + file.Extra.Ext + "?key=" + articleKey
}

func metaUrl(meta map[string]string) string {
	return fmt.Sprintf("https://mp.weixin.qq.com/s?__biz=%s&mid=%s&idx=%s&sn=%s", meta["biz"], meta["mid"], meta["idx"], meta["sn"])
}

func metaCreatedAt(meta map[string]string) string {
	if ct, ok := meta["svr_time"]; ok {
		words := strings.Split(ct, "*")
		return strings.Trim(words[0], "\" ")
	}
	return "0"
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
	defer rsp.Body.Close()
	return
}

type Article struct {
	ID int `json:"id"`
}

type ArticleResult struct {
	Article Article `json:"article"`
}

func createArticle(meta map[string]string) (art Article, err error) {
	var (
		form = urlLib.Values{}
		url  = articleHost + "/api/articles/"
		req  *http.Request
		rsp  *http.Response
	)
	form.Add("title", meta["msg_title"])
	form.Add("summary", meta["msg_desc"])
	form.Add("content", meta["msg_content"])
	form.Add("from_url", metaUrl(meta))
	form.Add("created_at", metaCreatedAt(meta))
	if req, err = http.NewRequest("POST", url, strings.NewReader(form.Encode())); err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	filledRequestHeader(req, form)
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		return
	}
	if rsp.StatusCode != 200 {
		err = errors.New("Create article failed.")
		return
	}
	decoder := json.NewDecoder(rsp.Body)
	var ret ArticleResult
	if err = decoder.Decode(&ret); err != nil {
		return
	}
	defer rsp.Body.Close()
	art = ret.Article
	return
}

func createTimeline(timeline string, art Article) (err error) {
	var (
		form = urlLib.Values{}
		url  = articleHost + "/api/timeline/" + timeline + "/"
		req  *http.Request
		rsp  *http.Response
	)
	form.Add("art_id", strconv.Itoa(art.ID))
	if req, err = http.NewRequest("POST", url, strings.NewReader(form.Encode())); err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	filledRequestHeader(req, form)
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		return
	}
	if rsp.StatusCode != 200 {
		raw, _ := ioutil.ReadAll(rsp.Body)
		err = fmt.Errorf("Create timeline failed (%s)", raw)
		return
	}
	defer rsp.Body.Close()
	return
}

func removeTimeline(timeline string, art Article) (err error) {
	var (
		url = articleHost + "/api/timeline/" + timeline + "/" + strconv.Itoa(art.ID) + "/"
		req *http.Request
		rsp *http.Response
	)
	if req, err = http.NewRequest("DELETE", url, nil); err != nil {
		return
	}
	filledRequestHeader(req, urlLib.Values{})
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		return
	}
	if rsp.StatusCode != 200 {
		err = errors.New("Create timeline failed.")
		return
	}
	defer rsp.Body.Close()
	return
}

func updateCover(art Article, fileId string) (err error) {
	var (
		form = urlLib.Values{}
		url  = articleHost + "/api/articles/" + strconv.Itoa(art.ID) + "/cover"
		req  *http.Request
		rsp  *http.Response
	)
	form.Add("file_id", fileId)
	if req, err = http.NewRequest("POST", url, strings.NewReader(form.Encode())); err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	filledRequestHeader(req, form)
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		return
	}
	if rsp.StatusCode != 200 {
		err = errors.New("Update cover failed.")
		return
	}
	defer rsp.Body.Close()
	return
}

func metaDoc(meta map[string]string) (doc Document) {
	ct, _ := strconv.ParseInt(metaCreatedAt(meta), 10, 0)
	return Document{
		ID:      metaUrl(meta),
		Title:   meta["msg_title"],
		Summary: meta["msg_desc"],
		Content: meta["msg_text"],
		Meta: map[string]string{
			"id":    meta["id"],
			"cover": meta["msg_cdn_url"],
		},
		CreatedAt: ct,
	}
}

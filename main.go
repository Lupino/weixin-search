package main

import (
	"encoding/json"
	"flag"
	"github.com/Lupino/go-periodic"
	"github.com/Lupino/tokenizer"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var (
	path          string
	host          string
	segoAddr      string
	name          string
	periodicHost  string
	docIndex      bleve.Index
	pclient       = periodic.NewClient()
	pworker       = periodic.NewWorker(2)
	r             = render.New()
	articleKey    string
	articleSecret string
	articleHost   string
	sharefsKey    string
	sharefsSecret string
	sharefsHost   string
	reUrl         = regexp.MustCompile("^https?://mp.weixin.qq.com/s")
)

func sendJSONResponse(w http.ResponseWriter, status int, key string, data interface{}) {
	if key == "" {
		r.JSON(w, status, data)
	} else {
		r.JSON(w, status, map[string]interface{}{key: data})
	}
}

func init() {
	flag.StringVar(&host, "host", "localhost:3030", "The search server host.")
	flag.StringVar(&path, "db", "simple-search.db", "The database path.")
	flag.StringVar(&periodicHost, "periodic", "unix:///tmp/periodic.sock", "The periodic server address")
	flag.StringVar(&name, "name", "weixin-search", "The search server name.")
	flag.StringVar(&articleKey, "article-key", "", "The article service key.")
	flag.StringVar(&articleSecret, "article-secret", "", "The article service secret.")
	flag.StringVar(&articleHost, "article-host", "https://gw.huabot.com", "The article service host.")
	flag.StringVar(&sharefsKey, "sharefs-key", "", "The sharefs service key.")
	flag.StringVar(&sharefsSecret, "sharefs-secret", "", "The sharefs service secret.")
	flag.StringVar(&sharefsHost, "sharefs-host", "https://gw.huabot.com", "The sharefs service host.")
	flag.StringVar(&segoAddr, "tokenizer", "localhost:3000", "tokenizer server host.")

	flag.Parse()
}

func main() {

	pclient.Connect(periodicHost)
	pworker.Connect(periodicHost)
	pworker.AddFunc(name+funcName, crawlLinkHandle)

	var router = mux.NewRouter()

	tokenizer.SegoTokenizerHost = segoAddr
	docIndex, _ = openIndex(path)

	go pworker.Work()

	router.HandleFunc("/api/docs/", func(w http.ResponseWriter, req *http.Request) {
		var (
			err error
		)
		req.ParseForm()
		uri := req.Form.Get("uri")
		if !reUrl.MatchString(uri) {
			sendJSONResponse(w, http.StatusBadRequest, "err", "Invalid weixin url.")
			return
		}
		if hasDocument(uri) {
			sendJSONResponse(w, http.StatusOK, "result", "OK")
			return
		}
		if err = submitCrawlLink(uri); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		sendJSONResponse(w, http.StatusOK, "result", "OK")
	}).Methods("POST")

	router.HandleFunc("/api/docs/count", func(w http.ResponseWriter, req *http.Request) {
		var docCount, err = docIndex.DocCount()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		sendJSONResponse(w, http.StatusOK, "", docCount)
	})

	router.HandleFunc("/api/docs/", func(w http.ResponseWriter, req *http.Request) {
		var (
			qs  = req.URL.Query()
			uri = qs.Get("uri")
			err error
		)
		if !reUrl.MatchString(uri) {
			sendJSONResponse(w, http.StatusBadRequest, "err", "Invalid weixin url.")
			return
		}
		if err = docIndex.Delete(uri); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		sendJSONResponse(w, http.StatusOK, "result", "OK")
	}).Methods("DELETE")

	router.HandleFunc("/api/search/", func(w http.ResponseWriter, req *http.Request) {
		var (
			qs    = req.URL.Query()
			err   error
			from  int
			size  int
			total uint64
			order []string
			q     = qs.Get("q")
		)
		if from, err = strconv.Atoi(qs.Get("from")); err != nil {
			from = 0
		}

		if size, err = strconv.Atoi(qs.Get("size")); err != nil {
			size = 10
		}

		if size > 100 {
			size = 100
		}

		if q == "" {
			sendJSONResponse(w, http.StatusBadRequest, "err", "q is required.")
			return
		}

		_query, err := query.ParseQuery([]byte(q))
		if err != nil {
			_query = bleve.NewQueryStringQuery(q)
		}

		searchRequest := bleve.NewSearchRequestOptions(_query, size, from, false)

		for _, word := range strings.Split(qs.Get("order"), " ") {
			word = strings.Trim(word, " \"'")
			if len(word) > 0 {
				order = append(order, word)
			}
		}

		if len(order) > 0 {
			searchRequest.SortBy(order)
		}

		searchResult, err := docIndex.Search(searchRequest)
		if err != nil {
			log.Printf("bleve.Index.Search() failed(%s)", err)
			sendJSONResponse(w, http.StatusBadRequest, "err", err)
			return
		}

		var hits = make([]hitResult, len(searchResult.Hits))
		for i, hit := range searchResult.Hits {
			doc, err := getDocument(hit.ID)
			if err != nil {
				continue
			}
			hits[i] = hitResult{
				ID:        hit.ID,
				Title:     doc.Title,
				Summary:   doc.Summary,
				Meta:      nil,
				Score:     hit.Score,
				CreatedAt: doc.CreatedAt,
			}
			json.Unmarshal([]byte(doc.Meta), &hits[i].Meta)
		}

		total = searchResult.Total

		sendJSONResponse(w, http.StatusOK, "", map[string]interface{}{
			"total": total,
			"from":  from,
			"size":  size,
			"hits":  hits,
		})
	}).Methods("GET")

	n := negroni.New(negroni.NewRecovery(), negroni.NewLogger())
	n.UseHandler(router)
	n.Run(host)
}

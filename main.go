package main

import (
	"flag"
	"github.com/Lupino/go-periodic"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/mholt/binding"
	"github.com/unrolled/render"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

var r = render.New()

func sendJSONResponse(w http.ResponseWriter, status int, key string, data interface{}) {
	if key == "" {
		r.JSON(w, status, data)
	} else {
		r.JSON(w, status, map[string]interface{}{key: data})
	}
}

var (
	path         string
	host         string
	verifyRegexp string
	periodicAddr string
	extractHost  string
	docIndex     bleve.Index
	pclient      = periodic.NewClient()
	pworker      = periodic.NewWorker(2)
)

func init() {
	flag.StringVar(&host, "host", "localhost:3030", "The search server host.")
	flag.StringVar(&path, "db", "simple-search.db", "The database path.")
	flag.StringVar(&periodicAddr, "periodic", "unix:///tmp/periodic.sock", "The periodic server address")
	flag.StringVar(&verifyRegexp, "regexp", ".*", "The valid host regexp.")
	flag.StringVar(&extractHost, "extract", "localhost:3031", "The content extract host.")
	flag.Parse()
}

func isValidHost(link string) bool {
	var (
		err     error
		u       *url.URL
		matched bool
	)

	if u, err = url.Parse(link); err != nil {
		return false
	}
	if matched, err = regexp.MatchString(verifyRegexp, u.Host); err != nil {
		return false
	}
	return matched
}

func main() {
	pclient.Connect(periodicAddr)
	pworker.Connect(periodicAddr)
	pworker.AddFunc(*funcName, indexDocHandle)
	pworker.AddFunc("hot-"+*funcName, indexHotHandle)

	var router = mux.NewRouter()
	docIndex, _ = openIndex(path)

	go pworker.Work()

	router.HandleFunc("/api/docs/", func(w http.ResponseWriter, req *http.Request) {
		doc := new(Document)
		errs := binding.Bind(req, doc)
		if errs.Handle(w) {
			return
		}
		if err := submitDoc(*doc); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		sendJSONResponse(w, http.StatusOK, "result", "OK")
	}).Methods("POST")

	// auto index on simple crawl
	router.HandleFunc("/api/docs/hot/", func(w http.ResponseWriter, req *http.Request) {
		var qs = req.URL.Query()
		uri := qs.Get("uri")
		if isValidHost(uri) {
			sendJSONResponse(w, http.StatusBadRequest, "err", "Invalid host.")
			return
		}
		if hasDocument(uri) {
			sendJSONResponse(w, http.StatusOK, "result", "OK")
			return
		}
		if err := submitHotLink(uri); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		sendJSONResponse(w, http.StatusOK, "result", "OK")
	}).Methods("GET")

	router.HandleFunc("/api/docs/", func(w http.ResponseWriter, req *http.Request) {
		var qs = req.URL.Query()
		uri := qs.Get("uri")
		var doc, err = getDocument(uri)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if doc == nil {
			sendJSONResponse(w, http.StatusNotFound, "err", "doc["+uri+"] not exists.")
			return
		}
		sendJSONResponse(w, http.StatusOK, "", doc)
	}).Methods("GET")

	router.HandleFunc("/api/docs/", func(w http.ResponseWriter, req *http.Request) {
		var qs = req.URL.Query()
		uri := qs.Get("uri")
		if err := docIndex.Delete(uri); err != nil {
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
		searchRequest.Highlight = bleve.NewHighlightWithStyle("html")
		searchRequest.Highlight.AddField("content")
		searchRequest.Highlight.AddField("title")
		searchRequest.Highlight.AddField("tags")
		searchResult, err := docIndex.Search(searchRequest)
		if err != nil {
			log.Printf("bleve.Index.Search() failed(%s)", err)
			sendJSONResponse(w, http.StatusBadRequest, "err", err)
			return
		}

		var hits = make([]hitResult, len(searchResult.Hits))
		for i, hit := range searchResult.Hits {
			hits[i] = hitResult{
				ID:        hit.ID,
				Fragments: filterFragments(hit.Fragments),
				Score:     hit.Score,
			}
		}

		total = searchResult.Total

		sendJSONResponse(w, http.StatusOK, "", map[string]interface{}{
			"total": total,
			"from":  from,
			"size":  size,
			"q":     q,
			"hits":  hits,
		})
	}).Methods("GET")

	n := negroni.New(negroni.NewRecovery(), negroni.NewLogger())
	n.UseHandler(router)
	n.Run(host)
}

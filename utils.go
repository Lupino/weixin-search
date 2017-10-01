package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// hmacSHA256
func hmacSHA256(slot string, params map[string]string) string {
	mac := hmac.New(sha256.New, []byte(slot))
	var keys []string
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		mac.Write([]byte(key))
		mac.Write([]byte(params[key]))
	}
	sum := mac.Sum(nil)
	return strings.ToUpper(hex.EncodeToString(sum))
}

func filledRequestHeader(req *http.Request, params url.Values) {
	var (
		signParams = make(map[string]string)
		timestamp  = strconv.FormatInt(time.Now().Unix(), 10)
		sign       string
	)
	signParams["pathname"] = req.URL.Path
	signParams["key"] = articleKey
	signParams["timestamp"] = timestamp
	for key := range params {
		signParams[key] = params.Get(key)
	}
	sign = hmacSHA256(articleSecret, signParams)

	req.Header.Add("X-REQUEST-KEY", articleKey)
	req.Header.Add("X-REQUEST-TIME", timestamp)
	req.Header.Add("X-REQUEST-SIGNATURE", sign)
}

func filledRequestHeaderWithRaw(req *http.Request) error {
	var (
		signParams = make(map[string]string)
		timestamp  = strconv.FormatInt(time.Now().Unix(), 10)
		reader     io.ReadCloser
		err        error
		sign       string
		raw        []byte
	)

	signParams["pathname"] = req.URL.Path
	signParams["key"] = articleKey
	signParams["timestamp"] = timestamp

	if reader, err = req.GetBody(); err != nil {
		return err
	}
	defer reader.Close()
	if raw, err = ioutil.ReadAll(reader); err != nil {
		return err
	}

	signParams["raw"] = string(raw)
	sign = hmacSHA256(articleSecret, signParams)

	req.Header.Add("X-REQUEST-KEY", articleKey)
	req.Header.Add("X-REQUEST-TIME", timestamp)
	req.Header.Add("X-REQUEST-SIGNATURE", sign)
	return nil
}

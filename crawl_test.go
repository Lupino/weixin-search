package main

import (
	"fmt"
	"testing"
)

func TestDoCrawl(t *testing.T) {
	var (
		uri = "https://mp.weixin.qq.com/s/0xRVDtKH9DCIffyMh4ha3w"
	)
	meta, err := doCrawl(uri)
	fmt.Printf("meta=%s\n", meta)
	fmt.Printf("err=%v\n", err)
}

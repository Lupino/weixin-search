package main

import (
	"fmt"
	"testing"
)

func TestDoCrawl(t *testing.T) {
	var (
		uri = "https://mp.weixin.qq.com/s/0xRVDtKH9DCIffyMh4ha3w"
	)
	articleHost = "http://127.0.0.1:3300"
	articleKey = "33c6615d7aaf88ff2ad1"
	articleSecret = "4de6cb3779eedbde19f794697d519c122894357a73be5a481467be5769948656"
	meta, err := doCrawl(uri)
	fmt.Printf("meta=%s\n", meta)
	fmt.Printf("err=%v\n", err)
}

func TestUploadFile(t *testing.T) {
	var (
		file   File
		err    error
		imgUrl = "http://mmbiz.qpic.cn/mmbiz_png/w9Cccd1M0afktEUibWQQoSU4UONkPMMUHIHOJMibe2ibhxFEn2KpjgGh9350GW0rREwHjcicdPrQskuJN7kqGydH9A/640?wx_fmt=png"
	)
	articleHost = "http://127.0.0.1:3300"
	articleKey = "33c6615d7aaf88ff2ad1"
	articleSecret = "4de6cb3779eedbde19f794697d519c122894357a73be5a481467be5769948656"
	if file, err = uploadImage(imgUrl, "jpg"); err != nil {
		t.Fatalf("uploadImage() failed (%s)", err)
	}
	fmt.Printf("image: %s\n", fileUrl(file))
}

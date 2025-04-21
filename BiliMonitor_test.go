package main

import (
	"net/url"
	"testing"
)

func TestGetBvid3And4(t *testing.T) {
	bvid3, bvid4, er := GetBvid3And4()
	if er != nil {
		t.Error(er)
	}
	t.Log(bvid3)
	t.Log(bvid4)
}
func TestGetBiliUserVideoList(t *testing.T) {
	VideoList, err := getBiliUserVideoList("1101360089", 1, 30)
	if err != nil {
		t.Error(err)
	}
	for i, list := range VideoList {
		t.Log(i, list.bvid, list.title)
	}
}
func TestMixinKeyGet(t *testing.T) {
	imgUrl, subUrl, _ := getBiliTicket()
	postData := url.Values{}
	postData.Add("foo", "114")
	postData.Add("bar", "514")
	postData.Add("zab", "1919810")
	t.Log(postData)
	mixinKeyGet(imgUrl, subUrl, postData)
	t.Log(postData)
}
func TestGetUnixTimeStamp(t *testing.T) {
	TimeStamp, err := getUnixTimeStamp()
	if err != nil {
		t.Error(err)
	}
	t.Log(TimeStamp)
}
func TestGetBiliTickey(t *testing.T) {
	imgUrl, subUrl, err := getBiliTicket()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(len(imgUrl), len(subUrl))
}

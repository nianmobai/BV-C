package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/url"
	"strings"
	"testing"
)

// 测试删除
func TestDelUp(t *testing.T) {
	var dsn = dsnRaw
	dsn = strings.Replace(dsn, "username", usrName, -1)
	dsn = strings.Replace(dsn, "password", psd, -1)
	dsn = strings.Replace(dsn, "protocol", "tcp", -1)
	dsn = strings.Replace(dsn, "address", addr, -1)
	dsn = strings.Replace(dsn, "dbname", dbName, -1)
	db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	delUp(&[]uint{1955897084}, db)
}

// 获取视频时间变化信息测试
func TestGetVideoStats(t *testing.T) {
	var dsn = dsnRaw
	dsn = strings.Replace(dsn, "username", usrName, -1)
	dsn = strings.Replace(dsn, "password", psd, -1)
	dsn = strings.Replace(dsn, "protocol", "tcp", -1)
	dsn = strings.Replace(dsn, "address", addr, -1)
	dsn = strings.Replace(dsn, "dbname", dbName, -1)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Log(err)
	}
	out := getVideoStats("BV1mpjhzdEuP", db)
	t.Log(out)
}

// 全局更新测试
func TestUpdate(t *testing.T) {
	var dsn = dsnRaw
	dsn = strings.Replace(dsn, "username", usrName, -1)
	dsn = strings.Replace(dsn, "password", psd, -1)
	dsn = strings.Replace(dsn, "protocol", "tcp", -1)
	dsn = strings.Replace(dsn, "address", addr, -1)
	dsn = strings.Replace(dsn, "dbname", dbName, -1)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Log(err)
	}
	uplist, _ := getUPList(db)
	t.Log(uplist)
	updateBvlist(&uplist, db)
	bvlist, _ := getAllBV(db)
	t.Log(bvlist)
	updateBVStat(&bvlist, db)
}

// 远程数据库读写测试
func TestWriteDBAndReadDB(t *testing.T) {
	//登录数据库
	var dsn = dsnRaw
	dsn = strings.Replace(dsn, "username", usrName, -1)
	dsn = strings.Replace(dsn, "password", psd, -1)
	dsn = strings.Replace(dsn, "protocol", "tcp", -1)
	dsn = strings.Replace(dsn, "address", addr, -1)
	dsn = strings.Replace(dsn, "dbname", dbName, -1)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Log(err)
	}
	//初始化数据
	var list []UpInfo
	list = append(list, UpInfo{
		Mid:  1955897084,
		Name: "鸣潮",
	})
	//保存UP表
	errW := saveUPList(&list, db)
	if errW != nil {
		t.Log(errW)
		return
	}
	//阅读UP表
	Read, errR := getUPList(db)
	if errR != nil {
		t.Log(errR)
		return
	}
	t.Log(Read)
}

// 初始化数据库测试
func TestCheckDB(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Log(err)
		return
	}
	t.Log("result")
}

// 根据视频BV爬取信息
func TestGetDetail(t *testing.T) {
	bvid := "BV12HJBzHERG"
	result, _, err := crawVideoDetail(bvid)
	if err != nil {
		t.Log("err", err.Error())
	}
	t.Log("播放量", result.Views)
	t.Log("投币", result.Coins)
	t.Log("点赞", result.Likes)
	t.Log("回复数", result.Reply)
	t.Log("在线观看人数", result.Online)
}

// 爬取bvid3与bvid4键测试
func TestGetBvid3And4(t *testing.T) {
	bvid3, bvid4, er := GetBvid3And4()
	if er != nil {
		t.Error(er)
	}
	t.Log(bvid3)
	t.Log(bvid4)
}

// 由用户获取视频列表测试
func TestGetBiliUserVideoList(t *testing.T) {
	VideoList, err := crawBiliUserVideoList(1955897084, 1, 30)
	if err != nil {
		t.Error(err)
	}
	for i, list := range VideoList {
		t.Log(i, list.Bvid, list.Title)
	}
}

// 计算mixingKey的算法测试
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

// 爬取时间戳测试
func TestGetUnixTimeStamp(t *testing.T) {
	TimeStamp, err := getUnixTimeStamp()
	if err != nil {
		t.Error(err)
	}
	t.Log(TimeStamp)
}

// 爬取tickey测试
func TestGetBiliTickey(t *testing.T) {
	imgUrl, subUrl, err := getBiliTicket()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(len(imgUrl), len(subUrl))
}

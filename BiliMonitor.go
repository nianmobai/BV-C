package main

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type BvList struct {
	Mid     uint
	Aid     int64
	Bvid    string `gorm:"type:varchar(255);primaryKey"` //主键
	Ctime   int64  //创建时间
	Pubdate int64  //发布时间
	Title   string //标题
	UpName  string
}
type UpInfo struct {
	Mid  uint `gorm:"primaryKey"` //主键
	Name string
}

type VideoStat struct {
	Bvid     string // BV号
	StatTime int64  // 统计时间
	Views    int    // 播放量
	Likes    int    // 点赞数
	Coins    int    // 硬币数
	Reply    int    //回复
	Online   int64  //在线观看人数
}

var usrName = "bili"
var psd = "ttlIEEE"
var addr = "localhost:3306"
var dbName = "biliMonitor"
var dsnRaw = "username:password@protocol(address)/dbname?charset=utf8mb4&parseTime=True"

var day = 10
var OneDay = 86400
var maxTimeInterval = day * OneDay

// 从数据库中获取视频的简要信息
func getVideoInfo(bv string, db *gorm.DB) (BvList, error) {
	var result BvList
	err := db.Where("bvid = ?", bv).First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return result, errors.New("not found")
	}
	return result, nil
}

// 全局更新数据
func globalUpdate() {
	log.Println("time cron")
	var dsn = dsnRaw
	dsn = strings.Replace(dsn, "username", usrName, -1)
	dsn = strings.Replace(dsn, "password", psd, -1)
	dsn = strings.Replace(dsn, "protocol", "tcp", -1)
	dsn = strings.Replace(dsn, "address", addr, -1)
	dsn = strings.Replace(dsn, "dbname", dbName, -1)
	db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	up, _ := getUPList(db)
	updateBvlist(&up, db)
	bvlist := getBvListInTenDays(db)
	updateBVStat(&bvlist, db)
}

// 删除视频
func delBv(list *[]string, db *gorm.DB) {
	var bvlist []BvList
	for _, item := range *list {
		bvlist = append(bvlist, BvList{
			Bvid: item,
		})
	}
	db.Delete(&bvlist)
}

// 删除Up主
func delUp(list *[]uint, db *gorm.DB) {
	var uplist []UpInfo
	for _, item := range *list {
		uplist = append(uplist, UpInfo{
			Mid: item,
		})
	}
	db.Delete(&uplist)
}

// 从数据库统计视频状态信息，返回的信息按时间升序排列
func getVideoStats(bv string, db *gorm.DB) []VideoStat {
	var result []VideoStat
	db.Where("bvid = ?", bv).Order("stat_time").Find(&result)
	return result
}

func getBvListInTenDays(db *gorm.DB) []BvList {
	var result []BvList
	Interval := time.Now().Unix() - int64(maxTimeInterval)
	db.Where("pubdate > ?", Interval).Find(&result)
	return result
}

// 更新视频列表
func updateBvlist(list *[]UpInfo, db *gorm.DB) {
	var result []BvList
	for _, item := range *list {
		BvItem, _ := crawBiliUserVideoList(item.Mid, 1, 10)
		for i := 0; i < len(BvItem); i++ {
			BvItem[i].UpName = item.Name
		}
		result = append(result, BvItem...)
	}
	log.Println(result)
	db.Save(&result)
}

// 更新视频信息
func updateBVStat(list *[]BvList, db *gorm.DB) {
	var result []VideoStat
	for _, item := range *list {
		//根据BV号获取视频信息存入数据库当中
		stat, _, err := crawVideoDetail(item.Bvid)
		if err != nil {
			continue
		}
		result = append(result, stat)
		time.Sleep(1 * time.Second)
	}
	db.Save(&result)
}

// 爬取视频当前时刻的信息
func crawVideoDetail(bv string) (VideoStat, BvList, error) {
	result := VideoStat{}
	info := BvList{}
	baseUrl := "https://api.bilibili.com/x/web-interface/view?bvid=bvReplace"
	baseUrl = strings.Replace(baseUrl, "bvReplace", bv, -1)
	req, err := http.NewRequest("GET", baseUrl, nil)
	if err != nil {
		return result, info, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36")
	client := &http.Client{}
	resp, errR := client.Do(req)
	if errR != nil {
		return result, info, errR
	}
	defer resp.Body.Close()
	rawRes, _ := io.ReadAll(resp.Body)
	var data map[string]interface{}
	json.Unmarshal(rawRes, &data)
	if data["code"].(float64) != 0 {
		return result, info, errors.New("get info:response error")
	}
	result.Bvid = bv                                                                                              //BV号
	result.StatTime = time.Now().Unix()                                                                           //时刻
	result.Views = int(data["data"].(map[string]interface{})["stat"].(map[string]interface{})["view"].(float64))  //播放量
	result.Likes = int(data["data"].(map[string]interface{})["stat"].(map[string]interface{})["like"].(float64))  //点赞数
	result.Coins = int(data["data"].(map[string]interface{})["stat"].(map[string]interface{})["coin"].(float64))  //投币数
	result.Reply = int(data["data"].(map[string]interface{})["stat"].(map[string]interface{})["reply"].(float64)) //回复量

	cid := int64(data["data"].(map[string]interface{})["cid"].(float64))
	onlineUrl := "https://api.bilibili.com/x/player/online/total?bvid={BVID}&cid={CID}"
	onlineUrl = strings.Replace(onlineUrl, "{BVID}", bv, -1)
	onlineUrl = strings.Replace(onlineUrl, "{CID}", strconv.FormatInt(cid, 10), -1)
	reqOnline, errOnline := http.NewRequest("GET", onlineUrl, nil)
	if errOnline != nil {
		return result, info, errOnline
	}
	reqOnline.Header.Set("Content-Type", "application/json; charset=utf-8")
	reqOnline.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36")
	respOnline, errOnline := client.Do(reqOnline)
	if errOnline != nil {
		return result, info, errOnline
	}
	defer respOnline.Body.Close()
	rawResOnline, _ := io.ReadAll(respOnline.Body)
	var total map[string]interface{}
	json.Unmarshal(rawResOnline, &total)
	if total["code"].(float64) != 0 {
		return result, info, errors.New("get total:response error")
	}
	result.Online, _ = strconv.ParseInt(total["data"].(map[string]interface{})["count"].(string), 10, 64)

	info.UpName = data["data"].(map[string]interface{})["owner"].(map[string]interface{})["name"].(string)
	info.Mid = uint(data["data"].(map[string]interface{})["owner"].(map[string]interface{})["mid"].(float64))
	info.Bvid = result.Bvid
	info.Title = data["data"].(map[string]interface{})["title"].(string)
	info.Pubdate = int64(data["data"].(map[string]interface{})["pubdate"].(float64))
	info.Ctime = int64(data["data"].(map[string]interface{})["ctime"].(float64))
	info.Aid = int64(data["data"].(map[string]interface{})["aid"].(float64))

	return result, info, nil
}

// 获取UP视频信息
func crawBiliUserVideoList(mid uint, pn int, ps int) ([]BvList, error) {
	baseUrl := "https://api.bilibili.com/x/series/recArchivesByKeywords"
	u, _ := url.Parse(baseUrl)
	postData := u.Query()
	postData.Add("mid", strconv.FormatUint(uint64(mid), 10))
	postData.Add("pn", strconv.Itoa(pn))
	postData.Add("ps", strconv.Itoa(ps))
	postData.Add("orderby", "pubdate")
	postData.Add("keywords", "")
	u.RawQuery = postData.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36")
	client := http.Client{}
	resp, errResp := client.Do(req)
	if errResp != nil {
		return nil, errResp
	}
	defer resp.Body.Close()
	result, _ := io.ReadAll(resp.Body)
	responseSave := make(map[string]interface{})
	json.Unmarshal(result, &responseSave)
	if responseSave["code"].(float64) != 0 {
		return nil, errors.New("upate list error")
	}
	list := responseSave["data"].(map[string]interface{})["archives"].([]interface{})
	var videolist = make([]BvList, len(list))
	for i := 0; i < len(list); i++ {
		videolist[i].Mid = mid
		videolist[i].Aid = int64(list[i].(map[string]interface{})["aid"].(float64))
		videolist[i].Bvid = list[i].(map[string]interface{})["bvid"].(string)
		videolist[i].Ctime = int64(list[i].(map[string]interface{})["ctime"].(float64))
		videolist[i].Pubdate = int64(list[i].(map[string]interface{})["pubdate"].(float64))
		videolist[i].Title = list[i].(map[string]interface{})["title"].(string)
		videolist[i].UpName = "unknown"
	}
	return videolist, nil
}

func crawUpInfo(mid uint) (UpInfo, error) {
	baseUrl := "https://api.bilibili.com/x/web-interface/card"
	var result UpInfo
	u, _ := url.Parse(baseUrl)
	postData := u.Query()
	postData.Add("mid", strconv.FormatUint(uint64(mid), 10))
	postData.Add("photo", "false")
	u.RawQuery = postData.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return result, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36")
	client := http.Client{}
	resp, errResp := client.Do(req)
	if errResp != nil {
		return result, errResp
	}
	defer resp.Body.Close()

	re, _ := io.ReadAll(resp.Body)
	responseSave := make(map[string]interface{})
	json.Unmarshal(re, &responseSave)
	if responseSave["code"].(float64) != 0 {
		return result, errors.New(responseSave["message"].(string))
	}
	result.Mid = mid
	result.Name = responseSave["data"].(map[string]interface{})["card"].(map[string]interface{})["name"].(string)
	return result, nil
}

// 将BV号保存至数据库
func saveBV(list *[]BvList, db *gorm.DB) error {
	Werr := db.Save(list).Error
	if Werr != nil {
		return Werr
	}
	return nil
}

// 从数据库获取BV表
func getAllBV(db *gorm.DB) ([]BvList, error) {
	var Bvlist []BvList
	db.Order("pubdate").Find(&Bvlist) //按时间排列
	return Bvlist, nil
}

// 从数据库获取UP表
func getUPList(db *gorm.DB) ([]UpInfo, error) {
	var list []UpInfo
	db.Find(&list)
	return list, nil
}

// 保存UP至数据库
func saveUPList(list *[]UpInfo, db *gorm.DB) error {
	Werr := db.Save(list).Error
	if Werr != nil {
		return Werr
	}
	return nil
}

// 初始化数据库,只在初始化时调用
func initDB() error {
	var dsn = dsnRaw
	dsn = strings.Replace(dsn, "username", usrName, -1)
	dsn = strings.Replace(dsn, "password", psd, -1)
	dsn = strings.Replace(dsn, "protocol", "tcp", -1)
	dsn = strings.Replace(dsn, "address", addr, -1)
	dsn = strings.Replace(dsn, "dbname", dbName, -1)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	db.Migrator().CreateTable(&UpInfo{})
	db.Migrator().CreateTable(&BvList{})
	db.Migrator().CreateTable(&VideoStat{})
	return nil
}

// 获取鉴权信息
func mixinKeyGet(imgUrl string, subUrl string, postData url.Values) {
	MIXIN_KEY_ENC_TAB := [...]uint8{46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35, 27, 43, 5, 49,
		33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13, 37, 48, 7, 16, 24, 55, 40,
		61, 26, 17, 0, 1, 60, 51, 30, 4, 22, 25, 54, 21, 56, 59, 6, 63, 57, 62, 11,
		36, 20, 34, 44, 52}
	var mixinKey []uint8
	var keyArr []string
	var beforeMD5 string
	postData.Add("wts", strconv.FormatInt(time.Now().Unix(), 10))
	rawWbiKey := imgUrl + subUrl
	for i := 0; i < 64; i++ {
		mixinKey = append(mixinKey, rawWbiKey[MIXIN_KEY_ENC_TAB[i]])
	}
	mixinKey = mixinKey[0:32]
	for key, _ := range postData {
		keyArr = append(keyArr, key)
	}

	sort.Strings(keyArr)
	beforeMD5 = ""
	for i := 0; i < len(keyArr); i++ {
		key := keyArr[i]
		var andChar string
		if i != 0 {
			andChar = "&"
		} else {
			andChar = ""
		}
		beforeMD5 = beforeMD5 + andChar + key + "=" + postData.Get(keyArr[i])
	}
	beforeMD5 = beforeMD5 + string(mixinKey)
	beforeMD5 = strings.Replace(beforeMD5, " ", "%20", -1) //替换空格字符串
	wRid := md5.Sum([]byte(beforeMD5))
	wRidHexString := hex.EncodeToString(wRid[:])
	postData.Set("w_rid", wRidHexString)
}

// 获取ticket参数
func getBiliTicket() (string, string, error) {
	//创建数据
	targetUrl := "https://api.bilibili.com/bapis/bilibili.api.ticket.v1.Ticket/GenWebTicket"
	keyId := "ec02"
	timeStamp := time.Now().Unix()
	hexsign := hmacSha256("ts" + strconv.FormatInt(timeStamp, 10))
	contextTs := strconv.FormatInt(timeStamp, 10)
	param := url.Values{}
	param.Set("key_id", keyId)
	param.Set("hexsign", hexsign)
	param.Set("context[ts]", contextTs)
	param.Set("csrf", "")
	//创建请求
	req, _ := http.NewRequest("POST", targetUrl, nil)
	req.Header.Set("Referer", "")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36")
	req.URL.RawQuery = param.Encode()
	//发出请求
	client := &http.Client{}
	resp, errGetBiliTicket := client.Do(req)
	if errGetBiliTicket != nil {
		return "", "", errors.New("can't get tickey")
	}
	body, _ := io.ReadAll(resp.Body)
	respStr := string(body)
	//提取img和sub
	var result map[string]interface{}
	json.Unmarshal([]byte(respStr), &result)
	code := result["code"].(float64)
	if code == -400 {
		return "", "", errors.New("parameter error")
	}
	img := result["data"].(map[string]interface{})["nav"].(map[string]interface{})["img"].(string)
	sub := result["data"].(map[string]interface{})["nav"].(map[string]interface{})["sub"].(string)
	imgParts := strings.Split(img, "/")
	img = imgParts[len(imgParts)-1]
	img = strings.Replace(img, ".png", "", -1)
	subPart := strings.Split(sub, "/")
	sub = subPart[len(subPart)-1]
	sub = strings.Replace(sub, ".png", "", -1)
	return img, sub, nil
}

// 获取时间戳
func getUnixTimeStamp() (int, error) {
	targetUrl := "https://api.bilibili.com/x/report/click/now"
	req, err := http.NewRequest("GET", targetUrl, nil)
	client := &http.Client{}
	resp, errGetTimeStamp := client.Do(req)
	if errGetTimeStamp != nil {
		log.Println("new request failed, err:", err)
		return -1, errGetTimeStamp
	}
	body, _ := io.ReadAll(resp.Body)
	respStr := string(body)
	var result map[string]interface{}
	_ = json.Unmarshal([]byte(respStr), &result)
	timeStamp := int(result["data"].(map[string]interface{})["now"].(float64))
	return timeStamp, nil
}

// 计算wbi签名
func hmacSha256(timeStamp string) string {
	secret := "XgwSnGZ1p" //密钥
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timeStamp))
	hmacHex := hex.EncodeToString(mac.Sum(nil))
	return hmacHex
}

// 获取bvid3和bvid4
func GetBvid3And4() (string, string, error) {
	targetUrl := "https://api.bilibili.com/x/frontend/finger/spi"
	req, _ := http.NewRequest("GET", targetUrl, nil)
	client := &http.Client{}
	resp, errGetBvid3 := client.Do(req)
	if errGetBvid3 != nil {
		return "", "", errGetBvid3
	}
	body, _ := io.ReadAll(resp.Body)
	respStr := string(body)
	var result map[string]interface{}
	json.Unmarshal([]byte(respStr), &result)
	if result["code"].(float64) != 0 {
		return "", "", errors.New("get bvid3,4 fail")
	}
	b_3 := result["data"].(map[string]interface{})["b_3"].(string)
	b_4 := result["data"].(map[string]interface{})["b_4"].(string)
	return b_3, b_4, nil
}

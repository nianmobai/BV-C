package main

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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

type BVLIST struct {
	mid     uint `gorm:"primaryKey"` //主键
	aid     uint
	bvid    string `gorm:"uniqueIndex"` //唯一索引
	ctime   uint   //创建时间
	pubdate uint   //发布时间
	title   string //标题
}
type UP struct {
	mid  uint   `gorm:"uniqueIndex"` //唯一索引
	name string `gorm:"size:100;not null"`
	fans int    `gorm:"not null"`
}

type VideoStat struct {
	bvid     string    `gorm:"size:12;primaryKey"` // BV号
	StatTime time.Time `gorm:"primaryKey;index"`   // 统计时间
	Views    int       // 播放量
	Likes    int       // 点赞数
	Coins    int       // 硬币数
	Reply    int       //回复
}

var tableName = "biliMoniter"
var usrName = "root"
var psd = "Nianmobai123/"
var protocol = "tcp"
var addr = "121.40.170.27:3360"
var dbName = "biliMonitor"
var dsnRaw = "username:password@protocol(address)/dbname?charset=utf8&parseTime=True"

func updateBVStat(list *[]BVLIST, db gorm.DB) error {
	var result []VideoStat
	for _, item := range *list {
		//根据BV号获取视频信息存入数据库当中
		info, err := getVideoDetail(item.bvid)
		if err != nil {
			continue
		}
		result = append(result, info)
	}
	err := db.Create(&result)
	if err != nil {
		return err.Error
	}
	return nil
}

// 获取视频当前时刻的信息
func getVideoDetail(bv string) (VideoStat, error) {
	result := VideoStat{}
	baseUrl := "https://api.bilibili.com/x/web-interface/view?bvid=bvReplace"
	url := strings.Replace(baseUrl, "bvReplace", bv, -1)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return result, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36")
	client := &http.Client{}
	resp, errR := client.Do(req)
	if errR != nil {
		return result, errR
	}
	defer resp.Body.Close()
	rawRes, _ := io.ReadAll(resp.Body)
	var data map[string]interface{}
	json.Unmarshal(rawRes, &data)
	if data["code"].(float64) != 0 {
		return result, errors.New("response error")
	}
	result.bvid = bv                                                                                             //BV号
	result.StatTime = time.Now()                                                                                 //时刻
	result.Views = int(data["data"].(map[string]interface{})["stat"].(map[string]interface{})["view"].(float64)) //播放量
	result.Likes = int(data["data"].(map[string]interface{})["stat"].(map[string]interface{})["like"].(float64)) //点赞数
	result.Coins = int(data["data"].(map[string]interface{})["stat"].(map[string]interface{})["coin"].(float64)) //投币数
	result.Reply = int(data["data"].(map[string]interface{})["stat"].(map[string]interface{})["reply"].(float64))
	return result, nil
}

// 将BV号保存至数据库
func updataBV(list *[]BVLIST, db gorm.DB) error {
	Werr := db.Save(list).Error
	if Werr != nil {
		return Werr
	}
	return nil
}

// 从数据库获取BV表
func getBV(list *[]UP, db gorm.DB) ([]BVLIST, error) {
	var Bvlist []BVLIST
	db.Find(&Bvlist)
	return Bvlist, nil
}

// 从数据库获取UP表
func getUPList(db gorm.DB) ([]UP, error) {
	var list []UP
	db.Find(&list)
	return list, nil
}

// 保存UP至数据库
func updateUPList(list *[]UP, db gorm.DB) error {
	Werr := db.Save(list).Error
	if Werr != nil {
		return Werr
	}
	return nil
}

// 检查数据库是否存在
func checkEnviroment() (bool, error) {
	var count int64
	var dsn = dsnRaw
	dsn = strings.Replace(dsn, "username", usrName, -1)
	dsn = strings.Replace(dsn, "password", psd, -1)
	dsn = strings.Replace(dsn, "protocol", protocol, -1)
	dsn = strings.Replace(dsn, "address", addr, -1)
	dsn = strings.Replace(dsn, "dbname", "", -1)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return false, err
	}
	errCheck := db.Raw("select count(*) FROM information_schema.schemata WHERE schema_name = " + tableName).Scan(&count).Error
	if errCheck != nil {
		return false, errCheck
	}
	return count > 0, nil
}

// 初始化数据库,只在初始化时调用
func initDB() error {
	var dsn = dsnRaw
	strings.Replace(dsn, "username", usrName, -1)
	strings.Replace(dsn, "password", psd, -1)
	strings.Replace(dsn, "protocol", "tcp", -1)
	strings.Replace(dsn, "address", addr, -1)
	strings.Replace(dsn, "dbname", "", -1)
	dbwithOutDB, errwithOutDB := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if errwithOutDB != nil {
		return errwithOutDB
	}
	errCreateDB := dbwithOutDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbName)).Error
	if errCreateDB != nil {
		return errCreateDB
	}
	dsn = dsnRaw
	strings.Replace(dsn, "username", usrName, -1)
	strings.Replace(dsn, "password", psd, -1)
	strings.Replace(dsn, "protocol", "tcp", -1)
	strings.Replace(dsn, "address", addr, -1)
	strings.Replace(dsn, "dbname", dbName, -1)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	log.Println(db.Migrator().CurrentDatabase()) //输出现在数据库
	if errCreateTbU := db.Migrator().CreateTable(&UP{}); errCreateTbU != nil {
		return errCreateTbU
	} //创建UP表
	if errCreateTbV := db.Table("Video").Create(&BVLIST{}); errCreateTbV != nil {
		return errCreateTbV.Error
	} //创建Video表
	if errCreateTbC := db.Table("VideoStat").Create(&VideoStat{}); errCreateTbC != nil {
		return errCreateTbC.Error
	} //创建数据统计表
	return nil
}

// 获取UP视频信息
func getBiliUserVideoList(mid string, pn int, ps int) ([]BVLIST, error) {
	baseUrl := "https://api.bilibili.com/x/series/recArchivesByKeywords"
	u, _ := url.Parse(baseUrl)
	postData := u.Query()
	postData.Add("mid", mid)
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
	var videolist = make([]BVLIST, len(list))
	for i := 0; i < len(list); i++ {
		videolist[i].aid = uint(list[i].(map[string]interface{})["aid"].(float64))
		videolist[i].bvid = list[i].(map[string]interface{})["bvid"].(string)
		videolist[i].ctime = uint(list[i].(map[string]interface{})["ctime"].(float64))
		videolist[i].pubdate = uint(list[i].(map[string]interface{})["pubdate"].(float64))
		videolist[i].title = list[i].(map[string]interface{})["title"].(string)
	}
	return videolist, nil
}

// 获取鉴权信息,不需要
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

package main

import (
	"encoding/json"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"slices"
	"strings"
)

var port = 12316
var vercode = "IDSFdsiaf+sd"

type Up struct {
	Mid uint   `json:"mid"`
	Ver string `json:"ver"`
}
type Bv struct {
	Bv  string `json:"bv"`
	Ver string `json:"ver"`
}

// 更新请求
func updateInfoHandle(w http.ResponseWriter, r *http.Request) {
	var dsn = dsnRaw
	data := make(map[string]interface{})

	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Add("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "GET" {
		dsn = strings.Replace(dsn, "username", usrName, -1)
		dsn = strings.Replace(dsn, "password", psd, -1)
		dsn = strings.Replace(dsn, "protocol", "tcp", -1)
		dsn = strings.Replace(dsn, "address", addr, -1)
		dsn = strings.Replace(dsn, "dbname", dbName, -1)
		db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})

		opt := r.URL.Query().Get("opt")
		if opt == "" {
			http.Error(w, "can't find parameter:opt", http.StatusBadRequest)
			return
		}
		if opt == "global" {
			globalUpdate()
			w.WriteHeader(http.StatusOK)
			data["code"] = 200
			data["message"] = "success"
		} else if opt == "vi" {
			list, _ := getAllBV(db)
			updateBVStat(&list, db)
			w.WriteHeader(http.StatusOK)
			data["code"] = 200
			data["message"] = "success"
		} else {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

	} else {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
	json.NewEncoder(w).Encode(data)
}

func addUPHandel(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	var quest Up
	var dbD []UpInfo
	var dsn = dsnRaw

	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Add("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Content-Type", "application/json")

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}
	if r.Method == "POST" {
		errP := r.ParseForm()
		if errP != nil {
			http.Error(w, errP.Error(), http.StatusBadRequest)
			return
		}
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields() //消除无意义字符

		errD := decoder.Decode(&quest)
		if errD != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		mid := quest.Mid
		ver := quest.Ver
		if ver != vercode {
			http.Error(w, "verify wrong", http.StatusBadRequest)
			return
		}
		dsn = strings.Replace(dsn, "username", usrName, -1)
		dsn = strings.Replace(dsn, "password", psd, -1)
		dsn = strings.Replace(dsn, "protocol", "tcp", -1)
		dsn = strings.Replace(dsn, "address", addr, -1)
		dsn = strings.Replace(dsn, "dbname", dbName, -1)
		db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{}) //打开数据库
		info, _ := crawUpInfo(mid)
		dbD = append(dbD, UpInfo{Mid: mid, Name: info.Name})
		errS := saveUPList(&dbD, db)
		if errS != nil {
			http.Error(w, errS.Error(), http.StatusBadRequest)
		}
		w.WriteHeader(http.StatusOK)
		data["code"] = 200
		data["data"] = make(map[string]interface{})
		data["data"].(map[string]interface{})["up_list"], _ = getUPList(db)
		data["message"] = "success"
	} else {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	json.NewEncoder(w).Encode(data)
}

func addVideoHandle(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	var quest Bv
	var dbD []BvList
	var dsn = dsnRaw

	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Add("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Content-Type", "application/json")

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}
	if r.Method == "POST" {
		errP := r.ParseForm()
		if errP != nil {
			http.Error(w, errP.Error(), http.StatusBadRequest)
			return
		}
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields() //消除无意义字符
		errD := decoder.Decode(&quest)
		if errD != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if quest.Ver != vercode {
			http.Error(w, "verify wrong", http.StatusBadRequest)
			return
		}

		_, ele, _ := crawVideoDetail(quest.Bv)
		dbD = append(dbD, ele)
		dsn = strings.Replace(dsn, "username", usrName, -1)
		dsn = strings.Replace(dsn, "password", psd, -1)
		dsn = strings.Replace(dsn, "protocol", "tcp", -1)
		dsn = strings.Replace(dsn, "address", addr, -1)
		dsn = strings.Replace(dsn, "dbname", dbName, -1)
		db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{}) //打开数据库
		saveBV(&dbD, db)

		w.WriteHeader(http.StatusOK)
		data["code"] = 200
		data["message"] = "success"
		data["data"] = make(map[string]interface{})
		data["data"].(map[string]interface{})["mid"] = ele.Mid
		data["data"].(map[string]interface{})["ctime"] = ele.Ctime
		data["data"].(map[string]interface{})["pubdate"] = ele.Pubdate
		data["data"].(map[string]interface{})["title"] = ele.Title
		data["data"].(map[string]interface{})["up_name"] = ele.UpName
	} else {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	json.NewEncoder(w).Encode(data)
}

// 获取请求
func getVideoDataHandle(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	var dsn = dsnRaw
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Add("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "GET" {
		dsn = strings.Replace(dsn, "username", usrName, -1)
		dsn = strings.Replace(dsn, "password", psd, -1)
		dsn = strings.Replace(dsn, "protocol", "tcp", -1)
		dsn = strings.Replace(dsn, "address", addr, -1)
		dsn = strings.Replace(dsn, "dbname", dbName, -1)
		db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{}) //打开数据库
		bv := r.URL.Query().Get("bv")
		timeOrder := r.URL.Query().Get("Timeorder")
		if bv == "" {
			w.WriteHeader(200)
			//data = map[string]interface{}{"data": nil, "code": 200, "message": "can't find parameter:bv"}
			data["data"] = nil
			data["code"] = 200
			data["message"] = "can't find parameter:bv"
			goto end
		}
		VideoInfo, errI := getVideoInfo(bv, db)
		if errI != nil {
			w.WriteHeader(200)
			//data = map[string]interface{}{"code": 200, "data": nil, "message": "video not exist"}
			data["data"] = nil
			data["code"] = 200
			data["message"] = "video not exist"
			goto end
		}
		if timeOrder == "" {
			timeOrder = "0"
		}
		VideoStatList := getVideoStats(bv, db)
		if timeOrder == "1" {
			slices.Reverse(VideoStatList)
		}
		w.WriteHeader(200)
		data["code"] = 200
		data["message"] = "success"
		data["data"] = make(map[string]interface{})
		data["data"].(map[string]interface{})["order"] = timeOrder
		data["data"].(map[string]interface{})["video_stats"] = VideoStatList
		data["data"].(map[string]interface{})["bv"] = VideoInfo.Bvid
		data["data"].(map[string]interface{})["title"] = VideoInfo.Title
		data["data"].(map[string]interface{})["pubdate"] = VideoInfo.Pubdate
		data["data"].(map[string]interface{})["up_mid"] = VideoInfo.Mid
	} else {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
end: //写入JSON数据
	json.NewEncoder(w).Encode(data)
}

func getUpListHandel(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	var dsn = dsnRaw
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Add("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "GET" {
		dsn = strings.Replace(dsn, "username", usrName, -1)
		dsn = strings.Replace(dsn, "password", psd, -1)
		dsn = strings.Replace(dsn, "protocol", "tcp", -1)
		dsn = strings.Replace(dsn, "address", addr, -1)
		dsn = strings.Replace(dsn, "dbname", dbName, -1)
		db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{}) //打开数据库
		uplist, err := getUPList(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(200)
		data["code"] = 0
		data["message"] = "success"
		data["data"].(map[string]interface{})["up_list"] = uplist
	} else {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	json.NewEncoder(w).Encode(data)
}

// 删除请求
func deleteUpHandel(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	var dsn = dsnRaw
	var quest Up
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Add("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Content-Type", "application/json")
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	if r.Method == "POST" {
		errP := r.ParseForm()
		if errP != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields() //消除无意义字符
		errD := decoder.Decode(&quest)
		if errD != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if quest.Ver != vercode {
			http.Error(w, "verify wrong", http.StatusBadRequest)
			return
		}

		dsn = strings.Replace(dsn, "username", usrName, -1)
		dsn = strings.Replace(dsn, "password", psd, -1)
		dsn = strings.Replace(dsn, "protocol", "tcp", -1)
		dsn = strings.Replace(dsn, "address", addr, -1)
		dsn = strings.Replace(dsn, "dbname", dbName, -1)
		db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{}) //打开数据库

		delUp(&[]uint{quest.Mid}, db)

		w.WriteHeader(http.StatusOK)
		data["code"] = 200
		data["message"] = "success"
	} else {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	json.NewEncoder(w).Encode(data)
}

func deleteVideoHandel(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	var dsn = dsnRaw
	var quest Bv
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Add("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Content-Type", "application/json")
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}
	if r.Method == "POST" {
		if errP := r.ParseForm(); errP != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields() //消除无意义字符
		errD := decoder.Decode(&quest)
		if errD != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if quest.Ver != vercode {
			http.Error(w, "verify wrong", http.StatusBadRequest)
			return
		}

		dsn = strings.Replace(dsn, "username", usrName, -1)
		dsn = strings.Replace(dsn, "password", psd, -1)
		dsn = strings.Replace(dsn, "protocol", "tcp", -1)
		dsn = strings.Replace(dsn, "address", addr, -1)
		dsn = strings.Replace(dsn, "dbname", dbName, -1)
		db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{}) //打开数据库

		delBv(&[]string{quest.Bv}, db)

		w.WriteHeader(http.StatusOK)
		data["code"] = 200
		data["message"] = "success"
	} else {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	json.NewEncoder(w).Encode(data)
}

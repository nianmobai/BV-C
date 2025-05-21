package main

import (
	"log"
	"net/http"
	"strconv"
)

func main() {
	exist, errCheck := checkEnviroment() //判断数据库是否存在
	if errCheck != nil {
		log.Println(errCheck)
		return //查询数据库异常
	}
	if !exist { //初始化
		errInit := initDB()
		if errInit != nil {
			log.Println(errCheck)
			return //初始化失败
		}
	}
	http.HandleFunc("/addup", adddUPHander)
	http.HandleFunc("/getdata", getVideoData)

	log.Println("start http server on port ", port)
	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		log.Println(err)
	}
}

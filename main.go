package main

import (
	"github.com/robfig/cron/v3"
	"log"
	"net/http"
	"strconv"
)

func main() {
	c := cron.New() //创建定时器

	c.AddFunc("@every 1h", globalUpdate)
	c.Start()
	defer c.Stop()
	//out数据
	http.HandleFunc("/query/getInfo", getVideoDataHandle)
	http.HandleFunc("/query/getUPList", getUpListHandel)
	//添加追踪项
	http.HandleFunc("/operate/add/up", addUPHandel)
	http.HandleFunc("/operate/add/video", addVideoHandle)
	//删除某一追踪项
	http.HandleFunc("/operate/delete/up", deleteUpHandel)
	http.HandleFunc("/operate/delete/up", deleteVideoHandel)
	//更新信息
	http.HandleFunc("/operate/updateInfo", updateInfoHandle)

	log.Println("start http server on port ", port)
	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		log.Println(err)
	} //服务异常退出
}

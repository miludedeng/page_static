package service

import (
	//"fmt"
	"github.com/astaxie/beego"
	"github.com/garyburd/redigo/redis"
	"page_static/config"
	"time"
)

var conn redis.Conn

func init() {
	if config.Basic.Storage != "redis" {
		return
	}
	var err error
	conn, err = redis.DialTimeout("tcp", config.Redis.Addr+":"+config.Redis.Port, 0, 1*time.Second, 1*time.Second)
	if err != nil {
		beego.Error("Redis Connect Failed!")
		panic(err)
	}
}

func GetHtmlRedis(cipherStr string) (html string, timeDifference int64) {
	html, err := redis.String(conn.Do("GET", cipherStr))
	if err != nil {
		return "", 0
	}
	date, dateErr := redis.Int64(conn.Do("GET", cipherStr+"_date"))
	if dateErr != nil {
		return "", 0
	}
	if date == 0 {
		timeDifference = 0
	} else {
		timeDifference = (time.Now().Unix() - date) / 60
	}
	return html, timeDifference
}

func GetHtmlAndSaveRedis(cipherStr, fullUrl string, concatCss bool) string {
	var html string
	if concatCss {
		html = GetHtmlConcatCss(fullUrl)
	} else {
		html = GetHtml(fullUrl)
	}
	conn.Do("SET", cipherStr, html)
	conn.Do("SET", cipherStr+"_date", time.Now().Unix())
	return html
}

package config

import (
	"github.com/astaxie/beego"
	"strconv"
)

type BasicT struct {
	HttpPort   int
	RunMode    string
	AppDomain  string
	MaxExpdate string
	ConcatCss  string
	Storage    string
}

type TextT struct {
	StaticPath string
}

type RedisT struct {
	Addr string
	Port string
}

var (
	Basic *BasicT
	Text  *TextT
	Redis *RedisT
)

func init() {
	Basic = &BasicT{}
	Text = &TextT{}
	Redis = &RedisT{}
	basicMap, _ := beego.AppConfig.GetSection("basic")
	textMap, _ := beego.AppConfig.GetSection("text")
	redisMap, _ := beego.AppConfig.GetSection("redis")
	Basic.AppDomain = "123"
	var httpPortErr error
	Basic.HttpPort, httpPortErr = strconv.Atoi(basicMap["httpport"])
	if httpPortErr != nil {
		Basic.HttpPort = 3000
	}
	Basic.RunMode = basicMap["runmode"]
	if Basic.RunMode == "" {
		Basic.RunMode = "prod"
	}
	Basic.AppDomain = basicMap["app_domain"]
	Basic.MaxExpdate = basicMap["max_expdate"]
	Basic.ConcatCss = basicMap["concat_css"]
	Basic.Storage = basicMap["storage"]
	Text.StaticPath = textMap["static_path"]
	Redis.Addr = redisMap["addr"]
	Redis.Port = redisMap["port"]
}

package main

import (
	"github.com/astaxie/beego"
	"page_static/config"
	_ "page_static/config"
	_ "page_static/controllers"
	_ "page_static/service"
)

func main() {
	beego.BConfig.AppName = "PageStatic"
	beego.BConfig.RunMode = config.Basic.RunMode
	beego.BConfig.Listen.HTTPPort = config.Basic.HttpPort
	beego.Run()
}

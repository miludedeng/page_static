package main

import (
	"github.com/astaxie/beego"
	"page_static/config"
	_ "page_static/config"
	_ "page_static/controllers"
	_ "page_static/service"
)

func main() {
	beego.BeeLogger.DelLogger("console")
	beego.BConfig.AppName = "PageStatic"
	beego.BConfig.RunMode = config.Basic.RunMode
	beego.BConfig.Listen.HTTPPort = config.Basic.HttpPort

	beego.SetLogger("file", `{"filename":"logs/log"}`)
	if beego.BConfig.RunMode == "prod" {
		beego.BeeLogger.DelLogger("console")
	}
	beego.Run()
}

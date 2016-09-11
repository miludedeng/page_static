package controllers

import (
	"github.com/astaxie/beego"
	"page_static/config"
	"page_static/service"
	"page_static/util"
	"strconv"
	"strings"
	"sync"
)

func init() {
	beego.Router("/*", &MainController{})
}

type MainController struct {
	beego.Controller
}

var cache = struct {
	sync.Mutex
	buffer *[]string
}{
	buffer: &[]string{},
}

type Storage struct {
	html           string
	timeDifference int64
	abPath         string
	isText         bool
}

func (c *MainController) Get() {
	url := c.Ctx.Request.URL.String()
	/* 配置文件中的域名 */
	domain := config.Basic.AppDomain
	if strings.HasSuffix(domain, "/") {
		domain = domain[0 : len(domain)-1]
	}
	/* 完整url路径 */
	fullUrl := domain + url
	/* 参数中携带nocache=true时，不使用缓存，建议只在测试时使用*/
	if c.GetString("nocache") == "true" {
		beego.Info("nocache=true")
		c.Ctx.WriteString(service.GetHtml(fullUrl))
		return
	}
	/* 过期时间 */
	// expDate := string(c.Ctx.Request.Header.Get("EXPDATE"))
	expDate, err := strconv.Atoi(c.Ctx.Request.Header.Get("EXPDATE"))
	if err != nil || expDate == 0 {
		expDate = config.Basic.MaxExpdate
	}
	cipherStr := util.EncodeUrl(url)
	useOldPage := false
	storage := &Storage{}
	if config.Basic.Storage == "text" {
		storage.isText = true
	}
	/* 获取静态文件并返回 */
	if storage.isText {
		storage.html, storage.timeDifference, storage.abPath = service.GetHtmlText(cipherStr)
	} else {
		storage.html, storage.timeDifference = service.GetHtmlRedis(cipherStr)
	}
	if storage.html != "" {
		useOldPage = true
		c.Ctx.WriteString(storage.html)
		beego.Info("Exist static file: ")
		beego.Info("\turl: " + fullUrl)
		beego.Info("\tmd5: " + cipherStr)
	}
	/* 如果静态文件的存储日期没有超出设置时间，则直接返回，否则继续存储*/
	if useOldPage && storage.timeDifference < int64(expDate) {
		return
	}
	/* 如果没有旧页面，并且已经有人访问过当前页面则阻止继续往下执行，并重复刷新访问者的浏览器 */
	if !cacheContains(cipherStr) {
		/* 如果之前已经存在静态页面，则静态页面应该返回给用户，此处不应该阻塞页面响应，所以此处使用异步加载html页面 */
		/* 如果页面没有静态页面则等待静态文件生成完成，并将http响应的页面返回给用户 */
		if storage.isText {
			if useOldPage {
				beego.Info("go create new")
				go service.GetHtmlAndSaveText(storage.abPath, fullUrl, config.Basic.ConcatCss == "on")
			} else {
				beego.Info("create new")
				bodyText := service.GetHtmlAndSaveText(storage.abPath, fullUrl, config.Basic.ConcatCss == "on")
				c.Ctx.WriteString(bodyText)
			}
		} else {
			if useOldPage {
				beego.Info("go create new")
				go service.GetHtmlAndSaveRedis(cipherStr, fullUrl, config.Basic.ConcatCss == "on")
			} else {
				beego.Info("create new")
				bodyText := service.GetHtmlAndSaveRedis(cipherStr, fullUrl, config.Basic.ConcatCss == "on")
				c.Ctx.WriteString(bodyText)
			}
		}
		beego.Info("staticize:", fullUrl, "  ###  ", cipherStr)
		cacheRemove(cipherStr)
	} else {
		c.Ctx.WriteString("<script>setTimeout('location.reload()',2000)</script>")
	}
}

func cacheContains(val string) (result bool) {
	cache.Lock()
	result = util.ContainsBySlice(*cache.buffer, val)
	cache.Unlock()
	return
}

func cahceAdd(val string) {
	cache.Lock()
	*cache.buffer = append(*cache.buffer, val)
	cache.Unlock()
}

func cacheRemove(val string) {
	cache.Lock()
	*cache.buffer = util.RemoveFromSlice(*cache.buffer, val)
	cache.Unlock()
}

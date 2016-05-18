package controllers

import (
	"github.com/astaxie/beego"
	"page_static/config"
	"page_static/service"
	"page_static/util"
	"strconv"
	"strings"
)

func init() {
	beego.Router("/*", &MainController{})
}

var md5S []string = []string{}

type MainController struct {
	beego.Controller
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
	nocache := c.GetString("nocache")
	if nocache == "true" {
		beego.Info("nocache=true")
		c.Ctx.WriteString(service.GetHtml(fullUrl))
		return
	}
	/* 过期时间 */
	expDate := string(c.Ctx.Request.Header.Get("EXPDATE"))
	if expDate == "" {
		expDate = config.Basic.MaxExpdate
	}
	expDateInt, _ := strconv.Atoi(expDate)
	expDateInt64 := int64(expDateInt)

	cipherStr := util.EncodeUrl(url)
	useOldPage := false
	var abPath string
	var html string
	var timeDifference int64
	/* 获取静态文件并返回 */
	if config.Basic.Storage == "text" {
		html, timeDifference, abPath = service.GetHtmlText(cipherStr)
		if html != "" {
			useOldPage = true
			c.Ctx.WriteString(html)
		}
	} else if config.Basic.Storage == "redis" {
		html, timeDifference = service.GetHtmlRedis(cipherStr)
		if html != "" {
			useOldPage = true
			c.Ctx.WriteString(html)
		}
	}
	if html != "" {
		beego.Info("Exist static file: ")
		beego.Info("\turl: " + fullUrl)
		beego.Info("\tmd5: " + cipherStr)
	}
	/* 如果静态文件的存储日期没有超出设置时间，则直接返回，否则继续存储*/
	beego.Info("\tTimeDifference:\t" + strconv.Itoa(int(timeDifference)))
	beego.Info("\tExpDate:\t" + strconv.Itoa(int(expDateInt64)))
	if html != "" && (timeDifference < expDateInt64 || util.ContainsBySlice(md5S, cipherStr)) {
		return
	}
	/* 如果没有旧页面，并且已经有人访问过当前页面则阻止继续往下执行，并重复刷新访问者的浏览器 */
	if !useOldPage && util.ContainsBySlice(md5S, cipherStr) {
		c.Ctx.WriteString("<script>setTimeout('location.reload()',2000)</script>")
		return
	}
	md5S = append(md5S, cipherStr)
	/* 如果之前已经存在静态页面，则静态页面应该返回给用户，此处不应该阻塞页面响应，所以此处使用异步加载html页面 */
	/* 如果页面没有静态页面则等待静态文件生成完成，并将http响应的页面返回给用户 */
	if "text" == config.Basic.Storage {
		if useOldPage {
			beego.Info("go create new")
			go service.GetHtmlAndSaveText(abPath, fullUrl, config.Basic.ConcatCss == "on")
		} else {
			beego.Info("create new")
			bodyText := service.GetHtmlAndSaveText(abPath, fullUrl, config.Basic.ConcatCss == "on")
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
	md5S = util.RemoveFromSlice(md5S, cipherStr)
}

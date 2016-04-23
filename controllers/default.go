package controllers

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/astaxie/beego"
	"page_static/config"
	"page_static/service"
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
	expDate := string(c.Ctx.Request.Header.Get("EXPDATE"))
	if expDate == "" {
		expDate = config.Basic.MaxExpdate
	}
	expDateInt, _ := strconv.Atoi(expDate)
	expDateInt64 := int64(expDateInt)
	url := c.Ctx.Request.URL.String()
	domain := config.Basic.AppDomain
	fullUrl := domain + url
	if strings.HasSuffix(domain, "/") {
		domain = domain[0 : len(domain)-1]
	}
	cipherStr := EncodeUrl(url)
	useOldPage := false
	var abPath string
	var html string
	var timeDifference int64
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
	if html != "" && timeDifference < expDateInt64 || ContainsBySlice(md5S, cipherStr) {
		return
	}
	if !useOldPage && ContainsBySlice(md5S, cipherStr) {
		c.Ctx.WriteString("<script>setTimeout('location.reload()',500)</script>")
		return
	}
	md5S = append(md5S, cipherStr)

	if "text" == config.Basic.Storage {
		if useOldPage {
			fmt.Println("go create new")
			go service.GetHtmlAndSaveText(abPath, fullUrl, config.Basic.ConcatCss == "on")
		} else {
			fmt.Println("create new")
			bodyText := service.GetHtmlAndSaveText(abPath, fullUrl, config.Basic.ConcatCss == "on")
			c.Ctx.WriteString(bodyText)
		}
	} else {
		if useOldPage {
			fmt.Println("go create new")
			go service.GetHtmlAndSaveRedis(cipherStr, fullUrl, config.Basic.ConcatCss == "on")
		} else {
			fmt.Println("create new")
			bodyText := service.GetHtmlAndSaveRedis(cipherStr, fullUrl, config.Basic.ConcatCss == "on")
			c.Ctx.WriteString(bodyText)
		}
	}
	md5S = RemoveFromSlice(md5S, cipherStr)
}

func ContainsBySlice(md5S []string, s string) bool {
	for _, v := range md5S {
		if v == s {
			return true
		}
	}
	return false
}

func RemoveFromSlice(s []string, e string) []string {
	indexS := []int{}
	for i, v := range s {
		if v == e {
			indexS = append(indexS, i)
		}
	}
	for _, v := range indexS {
		s = append(s[:v], s[v+1:]...)
	}
	return s
}

func EncodeUrl(url string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(string(url)))
	cipherEncode := md5Ctx.Sum(nil)
	return hex.EncodeToString(cipherEncode)
}

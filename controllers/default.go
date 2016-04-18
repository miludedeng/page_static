package controllers

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/astaxie/beego"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func init() {
	beego.Router("/*", &MainController{})
}

var md5S []string = []string{}

type MainController struct {
	beego.Controller
}

func (c *MainController) Get() {
	staticPath := beego.AppConfig.String("static_path")
	concatCss := beego.AppConfig.String("concat_css") == "on"
	beego.Info(concatCss)
	expDate := string(c.Ctx.Request.Header.Get("EXPDATE"))
	useOldPage := false
	if expDate == "" {
		expDate = beego.AppConfig.String("max_expdate")
	}
	expDateInt, _ := strconv.Atoi(expDate)
	expDateInt64 := int64(expDateInt)
	url := c.Ctx.Request.URL.String()
	domain := beego.AppConfig.String("app_domain")
	fullUrl := domain + url
	if strings.HasSuffix(domain, "/") {
		domain = domain[0 : len(domain)-1]
	}
	cipherStr := EncodeUrl(url)
	fileName := cipherStr + ".html"
	abDir := staticPath + "/" + fileName[0:1] + "/" + fileName[1:2]
	abPath := abDir + "/" + fileName
	if IsExist(abPath) {
		file, err1 := os.Open(abPath)
		fileState, _ := os.Stat(abPath)
		timeDifference := time.Now().Unix() - fileState.ModTime().Unix()
		timeDifference = timeDifference / 60
		if err1 != nil {
			panic(err1)
		}
		defer file.Close()
		fd, _ := ioutil.ReadAll(file)
		html := string(fd)
		c.Ctx.WriteString(html)
		useOldPage = true
		if timeDifference < expDateInt64 || ContainsBySlice(md5S, cipherStr) {
			return
		}
	}
	if !useOldPage && ContainsBySlice(md5S, cipherStr) {
		c.Ctx.WriteString("<script>setTimeout('location.reload()',500)</script>")
		return
	}
	md5S = append(md5S, cipherStr)
	errCreateDir := os.MkdirAll(abDir, 0755)
	if errCreateDir != nil {
		beego.Info(errCreateDir)
	}
	if useOldPage {
		fmt.Println("go create new")
		go GetResponseBodyText(abPath, fullUrl, concatCss)
	} else {
		fmt.Println("create new")
		bodyText := GetResponseBodyText(abPath, fullUrl, concatCss)
		c.Ctx.WriteString(bodyText)
	}
	md5S = RemoveFromSlice(md5S, cipherStr)
}

func GetResponseBodyText(abPath string, fullUrl string, concatCss bool) string {
	var html string
	if concatCss {
		html = GetHtmlConcatCss(fullUrl)
	} else {
		html = GetHtml(fullUrl)
	}
	os.Remove(abPath)
	newFile, errCreateFile := os.Create(abPath)
	if errCreateFile != nil {
		beego.Info(errCreateFile)
	}
	n, errWriterFile := io.WriteString(newFile, html)
	if errWriterFile != nil {
		beego.Info(errCreateFile)
	}
	beego.Info("写入文件：" + abPath + "    " + strconv.Itoa(n) + "字节")
	return html
}

func GetHtml(fullUrl string) string {
	resp, err := http.Get(fullUrl)
	if err != nil || resp.StatusCode != 200 {
		fmt.Println("error page")
		return ""
	}
	defer resp.Body.Close()
	body, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		fmt.Println("body parse error")
		return ""
	}
	return string(body)
}
func GetHtmlConcatCss(fullUrl string) string {
	doc, err := goquery.NewDocument(fullUrl)
	if err != nil {
		fmt.Println(err)
	}
	cssUrls := []string{}
	doc.Find("link[rel=stylesheet]").Each(func(i int, node *goquery.Selection) {
		cssUrl, _ := node.Attr("href")
		if cssUrl != "" {
			cssUrls = append(cssUrls, cssUrl)
		}
		node.Remove()
	})
	cssAll := ""
	for _, v := range cssUrls {
		resp, _ := http.Get(v)
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		cssAll = cssAll + "\r\n" + string(body)
		fmt.Println(v)
	}
	doc.Find("title").AfterHtml("\r\n<style>" + cssAll + "</style>\r\n")
	html, _ := doc.Html()
	re, _ := regexp.Compile("\\/\\*.*\\*\\/")
	html = re.ReplaceAllString(html, "")
	return html
}
func IsExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
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

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
	url := c.Ctx.Request.URL.String()
	/* 配置文件中的域名 */
	domain := beego.AppConfig.String("app_domain")
	if strings.HasSuffix(domain, "/") {
		domain = domain[0 : len(domain)-1]
	}
	/* 完整url路径 */
	fullUrl := domain + url
	/* 参数中携带nocache=true时，不使用缓存，建议只在测试时使用*/
	nocache := c.GetString("nocache")
	if nocache == "true" {
		fmt.Println("nocache=true")
		c.Ctx.WriteString(GetHtml(fullUrl))
		return
	}
	fmt.Println("no-nocache")
	/* 静态文件存储路径 */
	staticPath := beego.AppConfig.String("static_path")
	/* concat_css=on选项开启，页面中直接引用的css文件会被封到页面中，以减少请求次数 */
	concatCss := beego.AppConfig.String("concat_css") == "on"
	/* 请求Header中的过期时间 */
	expDate := string(c.Ctx.Request.Header.Get("EXPDATE"))

	useOldPage := false
	if expDate == "" {
		/* 如果请求Header中没有过期时间，则使用配置文件中的最大过期时间 */
		expDate = beego.AppConfig.String("max_expdate")
	}
	expDateInt, _ := strconv.Atoi(expDate)
	expDateInt64 := int64(expDateInt)

	/* 对url使用md5加密，并按照md5密文的前两个字母创建两层目层(这样做可预防文件过多导致的IO对写慢的问题) */
	cipherStr := EncodeUrl(url)
	fileName := cipherStr + ".html"
	abDir := staticPath + "/" + fileName[0:1] + "/" + fileName[1:2]
	abPath := abDir + "/" + fileName
	if IsExist(abPath) {
		file, err1 := os.Open(abPath)
		if err1 != nil {
			panic(err1)
		}
		defer file.Close()
		fileState, _ := os.Stat(abPath)
		/* 文件最后修改时间与当前时间的时间差，用于判断文件是否过期 */
		timeDifference := time.Now().Unix() - fileState.ModTime().Unix()
		timeDifference = timeDifference / 60
		/* 读取生成的静态文件中的文本内容并返回(此处不考虑是否过期，也就是说如果文件过期了，第一次访问者仍然访问的是过期文件) */
		htmlBytes, _ := ioutil.ReadAll(file)
		html := string(htmlBytes)
		c.Ctx.WriteString(html)
		useOldPage = true
		/* 如果文件没有过期，则直接return */
		if timeDifference < expDateInt64 || ContainsBySlice(md5S, cipherStr) {
			return
		}
	}
	/* 如果没有旧页面，并且已经有人访问过当前页面则阻止继续往下执行，并重复刷新访问者的浏览器 */
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
		/* 如果之前已经存在静态页面，则静态页面应该返回给用户，此处不应该阻塞页面响应，所以此处使用异步加载html页面 */
		fmt.Println("go create new")
		go GetResponseBodyText(abPath, fullUrl, concatCss)
	} else {
		/* 如果页面没有静态页面则等待静态文件生成完成，并将http响应的页面返回给用户 */
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

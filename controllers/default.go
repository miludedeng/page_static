package controllers

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/astaxie/beego"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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
	expDate := string(c.Ctx.Request.Header.Get("EXPDATE"))
	useOldPage := false
	if expDate == "" {
		expDate = beego.AppConfig.String("max_expdate")
	}
	expDateInt, _ := strconv.Atoi(expDate)
	expDateInt64 := int64(expDateInt)
	url := c.Ctx.Request.URL.String()
	domain := beego.AppConfig.String("app_domain")
	if strings.HasSuffix(domain, "/") {
		domain = domain[0 : len(domain)-1]
	}
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(string(url)))
	cipherEncode := md5Ctx.Sum(nil)
	cipherStr := hex.EncodeToString(cipherEncode)
	fileName := cipherStr + ".html"
	folder1 := fileName[0:1]
	folder2 := fileName[1:2]
	abPath := staticPath + "/" + folder1 + "/" + folder2 + "/" + fileName
	if isExist(abPath) {
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
		if timeDifference <= expDateInt64 || ContainsBySlice(md5S, cipherStr) {
			return
		}
	}
	if !useOldPage && ContainsBySlice(md5S, cipherStr) {
		fmt.Println(2, md5S)
		c.Ctx.WriteString("<script>setTimeout('location.reload()',500)</script>")
		return
	}
	md5S = append(md5S, cipherStr)
	errCreateDir := os.MkdirAll(staticPath+"/"+folder1+"/"+folder2, 0755)
	if errCreateDir != nil {
		beego.Info(errCreateDir)
	}
	resp, err2 := http.Get(domain + url)
	if err2 != nil || resp.StatusCode != 200 {
		c.Ctx.WriteString("error page")
		return
	}
	defer resp.Body.Close()
	body, err3 := ioutil.ReadAll(resp.Body)
	if err3 != nil {
		c.Ctx.WriteString("body parse error")
		return
	}
	if !useOldPage {
		c.Ctx.WriteString(string(body))
	}
	os.Remove(abPath)
	newFile, errCreateFile := os.Create(abPath)
	if errCreateFile != nil {
		beego.Info(errCreateFile)
	}
	n, errWriterFile := io.WriteString(newFile, string(body))
	if errWriterFile != nil {
		beego.Info(errCreateFile)
	}
	beego.Info("写入文件：" + abPath + "    " + strconv.Itoa(n) + "字节")
	md5S = RemoveFromSlice(md5S, cipherStr)
	fmt.Println(1, md5S)
	return
}

func isExist(filename string) bool {
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

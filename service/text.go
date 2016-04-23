package service

import (
	//"fmt"
	"github.com/astaxie/beego"
	"io"
	"io/ioutil"
	"os"
	"page_static/config"
	"strconv"
	"time"
)

func GetHtmlText(cipherStr string) (html string, timeDifferencetime int64, abPath string) {
	fileName := cipherStr + ".html"
	abDir := config.Text.StaticPath + "/" + cipherStr[0:1] + "/" + cipherStr[1:2]
	abPath = abDir + "/" + fileName

	if IsExist(abPath) {
		file, err1 := os.Open(abPath)
		fileState, _ := os.Stat(abPath)
		defer file.Close()
		if err1 != nil {
			panic(err1)
		}
		fd, _ := ioutil.ReadAll(file)
		html := string(fd)
		timeDifference := (time.Now().Unix() - fileState.ModTime().Unix()) / 60
		return html, timeDifference, abPath
	} else {
		errCreateDir := os.MkdirAll(abDir, 0755)
		if errCreateDir != nil {
			beego.Info(errCreateDir)
		}
		return "", 0, abPath
	}
}

func GetHtmlAndSaveText(abPath string, fullUrl string, concatCss bool) string {
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

func IsExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

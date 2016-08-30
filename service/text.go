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
		file, err := os.Open(abPath)
		fileState, _ := os.Stat(abPath)
		defer file.Close()
		if err != nil {
			panic(err)
		}
		fd, _ := ioutil.ReadAll(file)
		html := string(fd)
		fileModifyTime := fileState.ModTime()
		timeDifference := (time.Now().Unix() - fileModifyTime.Unix()) / 60
		beego.Info("Time sub is " + strconv.Itoa(int((time.Now().Unix()-fileModifyTime.Unix())/60)))
		beego.Info("File modify time is " + fileModifyTime.Format("2006-01-02 15:04:05"))
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
	if html == "" {
		return ""
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

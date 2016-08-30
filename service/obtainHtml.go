package service

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

func GetHtml(fullUrl string) string {
	resp, err := http.Get(fullUrl)
	if err != nil || resp.StatusCode != 200 {
		fmt.Println("error page")
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
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
		resp, err := http.Get(v)
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		cssAll = cssAll + "\r\n" + string(body)
	}
	doc.Find("title").AfterHtml("\r\n<style>" + cssAll + "</style>\r\n")
	html, _ := doc.Html()
	re, _ := regexp.Compile("\\/\\*.*\\*\\/")
	html = re.ReplaceAllString(html, "")
	return html
}

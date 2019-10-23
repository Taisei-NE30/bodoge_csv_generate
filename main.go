package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/sclevine/agouti"
	"log"
	"strconv"
	"strings"
	"time"

	//"net/http"
)

func main() {
	url := "https://bodoge.hoobby.net/games"
	driver := agouti.ChromeDriver()

	err := driver.Start()
	if err != nil {
		log.Printf("Failed to start driver: %v", err)
	}
	defer driver.Stop()

	page, err := driver.NewPage(agouti.Browser("chrome")) // クロームを起動。page型の返り値（セッション）を返す。
	if err != nil {
		log.Printf("Failed to open page: %v", err)
	}

	err = page.Navigate(url) // 指定したurlにアクセスする
	if err != nil {
		log.Printf("Failed to navigate: %v", err)
	}

	curContentsDom, err := page.HTML()
	if err != nil {
		log.Printf("Failed to get html: %v", err)
	}

	readerCurContents := strings.NewReader(curContentsDom)
	contentsDom, _ := goquery.NewDocumentFromReader(readerCurContents)
	gameList := contentsDom.Find(".list--games > ul").Children()
	listLen := gameList.Length()

	for i := 0; i < listLen; i++ {
		page.Find(".list--game > ul > li:nth-child(" + strconv.Itoa(i) + ") > a").Click()
		time.Sleep(5 * time.Second) //ブラウザが反応するまで待つ
		curContentsDom, err := page.HTML()
		if err != nil {
			log.Printf("Failed to get html: %v", err)
		}
		readerCurContents := strings.NewReader(curContentsDom)
		contentsDom, _ := goquery.NewDocumentFromReader(readerCurContents)
		contentsDom.Find("")
	}

}

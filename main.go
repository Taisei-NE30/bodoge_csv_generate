package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/sclevine/agouti"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	//"net/http"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	_ = json.NewEncoder(f).Encode(token)
}

func main() {
	mechanismStart := 6
	mechanism := []string{
		"auction", "diceRoll", "tilePlacement", "bluff", "areaMajority", "hiddenRoles", "cooperative",
		"workerPlacement", "balance", "draft", "network", "stock", "trickTaking", "burst", "setCollection",
		"handManagement", "deckBuilding", "batting", "negotiation", "team", "actionPoint", "variablePhaseOrder",
		"actionPlot", "realTime", "memory", "reasoning", "word", "action", "storyMaking", "variablePlayerPower",
		"drawing", "legacy", "escapeRoom",
	}
	spreadsheetId := "1FIgJ7QfdaWDwZ8KOEydTss-eBmQB6_38nsTS-hy-EUg"

	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	//If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	url := "https://bodoge.hoobby.net/games"
	driver := agouti.ChromeDriver()

	err = driver.Start()
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
	numRe := regexp.MustCompile(`[0-9]+`)
	timeRe := regexp.MustCompile(`（.+）`)

	for i := 0; i < listLen; i++ {
		var row []string
		err := page.FindByXPath("//div[@class='list--games']/ul/li[position()=" + strconv.Itoa(i+1) +"]/a").Click()
		if err != nil {
			log.Fatalf("Failed to click: %v", err)
		}
		time.Sleep(5 * time.Second) //ブラウザが反応するまで待つ

		curContentsDom, err := page.HTML()
		if err != nil {
			log.Printf("Failed to get html: %v", err)
		}
		readerCurContents := strings.NewReader(curContentsDom)
		contentsDom, _ := goquery.NewDocumentFromReader(readerCurContents)
		contentsDom.Find("div.product > table > tbody > tr").Each(func(index int, s *goquery.Selection) {
			switch index {
			case 0:
				row = append(row, s.Find("td").Text())

			case 2:
				data := s.Find("td").Text()
				players := numRe.FindAllString(data, -1)
				row = append(row, players[0], players[1])
				time := strings.Trim(timeRe.FindString(data), "（）")
				if strings.Contains(time, "未登録") {
					row = append(row, "")
				} else {
					row = append(row, time)
				}

			case 3:
				age := s.Find("td").Text()
				row = append(row, numRe.FindString(age))

			case 4:
				year := s.Find("td").Text()
				row = append(row, numRe.FindString(year))
			}
		})
		fmt.Println(row)
	}

}

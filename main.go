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
	mechanisms := []string {
		"Auction / Bidding", "Dice Rolling", "Tile Placement", "Bluff", "Area Majority / Area Control / Area influence",
		"Mafia Game / Concealment", "Co-operative Play", "Worker Placement", "Balance Game", "Drafting",
		"Route / Network Building", "Investment", "Trick-taking", "Burst", "Set Collection", "Hand Management",
		"Deck / Pool Building", "Batting", "Negotiation", "Partnerships", "Action Point System", "Variable Phase Order",
		"Simultaneous Action Selection", "Real-time", "Memory", "Deduction", "Word Game", "Action", "Storytelling",
		"Variable Player Powers", "drawing", "Legacy System", "Mystery",
	}
	themes := []string {
		"Civilization", "Fantasy", "Cthulhu", "Galaxy / Star", "Science Fiction", "War / Militaly", "Exploring",
		"City Builder", "Territory", "Animal", "Mafia / Yakuza", "Detective", "Spy / Agent", "Zombie",
		"Ninja / Samurai", "Pirates / Vikings", "Farming", "Music", "Sports", "Train / Railway", "Non-Themed",
	}

	var inputData [][]interface{}
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


	var row []interface{}
	numRe := regexp.MustCompile(`[0-9]+`)
	playerRe := regexp.MustCompile(`(?m)[0-9]+人(用|～[0-9]+人)`)
	timeRe := regexp.MustCompile(`（.+）`)

	for {
		readerCurContents := strings.NewReader(curContentsDom)
		contentsDom, _ := goquery.NewDocumentFromReader(readerCurContents)
		gameList := contentsDom.Find(".list--games > ul").Children()
		listLen := gameList.Length()

		if listLen < 60 {
			break
		}

		for i := 0; i < listLen; i++ {
			err := page.FindByXPath("//div[@class='list--games']/ul/li[position()=" + strconv.Itoa(i+1) + "]/a").Click()
			if err != nil {
				log.Fatalf("Failed to click: %v", err)
			}
			time.Sleep(1 * time.Second) //ブラウザが反応するまで待つ

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
					players := playerRe.FindString(data)
					playerNum := numRe.FindAllString(players, -1)
					if len(playerNum) == 1 {
						row = append(row, playerNum[0], playerNum[0])
					} else {
						row = append(row, playerNum[0], playerNum[1])
					}
					gameTime := strings.Trim(timeRe.FindString(data), "（）")
					if strings.Contains(gameTime, "未登録") {
						row = append(row, "", "")
					} else {
						timeTrim := numRe.FindAllString(gameTime, -1)
						if len(timeTrim) == 1 {
							row = append(row, timeTrim[0], timeTrim[0])
						} else {
							row = append(row, timeTrim[0], timeTrim[1])
						}

					}

				case 3:
					age := s.Find("td").Text()
					row = append(row, numRe.FindString(age))

				case 4:
					year := s.Find("td").Text()
					row = append(row, numRe.FindString(year))
				}
			})
			designer := contentsDom.Find("div.credit > table > tbody > tr:nth-child(1) > td > a").Text()
			row = append(row, designer)

			isNoData := true

			if val, exists := contentsDom.Find("div.mechanics > div").Attr("class"); val == "empty" || !exists {
				for i := 0; i < len(mechanisms); i++ {
					row = append(row, 0)
				}
			} else {
				for _, mechanism := range mechanisms {
					isNoData = true
					contentsDom.Find("div.mechanics > ul > li").Each(func(index int, selection *goquery.Selection) {
						if mechanism == selection.Find("a > div.en").Text() {
							row = append(row, 1)
							isNoData = false
						}
					})
					if isNoData {
						row = append(row, 0)
					}
				}
			}

			if val, exists := contentsDom.Find("div.concepts > div").Attr("class"); val == "empty" || !exists {
				for i := 0; i < len(mechanisms); i++ {
					row = append(row, 0)
				}
			} else {
				for _, theme := range themes {
					isNoData = true
					contentsDom.Find("div.concepts > ul > li").Each(func(index int, selection *goquery.Selection) {
						if theme == selection.Find("a > div.en").Text() {
							row = append(row, 1)
							isNoData = false
						}
					})
					if isNoData {
						row = append(row, 0)
					}
				}
			}
			fmt.Println(row)
			inputData = append(inputData, row)
			row = nil

			err = page.Back()
			if err != nil {
				log.Fatalf("Failed to back: %v", err)
			}
			time.Sleep(1 * time.Second)
		}
		writeRange := "シート1!A2"
		valueRange := &sheets.ValueRange{
			Values: inputData,
		}
		_, err = srv.Spreadsheets.Values.Update(spreadsheetId, writeRange, valueRange).
			ValueInputOption("RAW").Do()
		if err != nil {
			log.Fatalf("Unable to retrieve data from sheet. %v", err)
		} else {
			fmt.Println("Success!")
		}
	}

}

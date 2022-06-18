package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	tb "gopkg.in/telebot.v3"
)

var (
	botKey  = os.Getenv("AI_KEY")
	botAPI  = "https://icap.iconiq.ai/talk"
	tgToken = os.Getenv("TOKEN")
)

var (
	httpClient = &http.Client{Timeout: time.Second * 10}
	IDS        map[int64]string
	btn        tb.InlineButton
)

func Talk(message string, sessionID string) string {
	params := map[string]string{
		"botkey":      botKey,
		"client_name": "uuiprod-un18e6d73c-user-19422",
		"sessionid":   sessionID,
		"channel":     "7",
		"input":       message,
		"id":          "true",
	}
	req, err := http.NewRequest("POST", botAPI, strings.NewReader(encodeParams(params)))
	if err != nil {
		return "Failed to get response"
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := httpClient.Do(req)
	if err != nil {
		return "Failed to get response"
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "Failed to get response"
	}
	var respMap map[string]interface{}
	err = json.Unmarshal(respBody, &respMap)
	if err != nil {
		return "Failed to get response"
	}
	resP := respMap["responses"].([]interface{})[0].(string)
	msg, _, _ := getMetaData(resP)
	return msg
}

func encodeParams(params map[string]string) string {
	var pairs []string
	for k, v := range params {
		pairs = append(pairs, k+"="+v)
	}
	return strings.Join(pairs, "&")
}

func genRandomID() (int, error) {
	b := make([]byte, 4)
	rand.Read(b)
	return strconv.Atoi(fmt.Sprintf("%d", int(b[0])<<24+int(b[1])<<16+int(b[2])<<8+int(b[3]))[:9])
}

func getChatSessionID(chatID int64) string {
	if IDS == nil {
		IDS = make(map[int64]string)
	}
	if IDS[chatID] == "" {
		ID, _ := genRandomID()
		IDS[chatID] = fmt.Sprintf("%d", ID)
	}
	return IDS[chatID]
}

func main() {
	bot, err := tb.NewBot(tb.Settings{
		Token: tgToken,
		Poller: &tb.LongPoller{
			Timeout: 10 * time.Second,
		},
	})
	if err != nil {
		panic(err)
	}
	bot.Handle("/start", func(c tb.Context) error {
		return c.Reply("Hello, I'm Mitsuki AI")
	})
	bot.Handle("/talk", func(c tb.Context) error {
		return c.Reply(Talk(c.Message().Payload, getChatSessionID(c.Chat().ID)), &tb.SendOptions{
			ParseMode: tb.ModeHTML,
		})
	})
	bot.Start()
}

func getMetaData(resp string) (string, string, bool) {
	var (
		media   string
		card    bool
		buttons []map[string]string
	)
	if matches := regexp.MustCompile(`<image>(.*?)</image>`).FindAllStringSubmatch(resp, -1); matches != nil {
		for _, match := range matches {
			resp = strings.Replace(resp, match[0], "", -1)
			if strings.Contains(match[1], "https://web23.secure-secure.co.uk/square-bear.co.uk/pandorabots/giphylogo.png") {
				continue
			}
			media = match[1]
		}
	}
	if matches := regexp.MustCompile(`<card>(.*?)</card>`).FindAllStringSubmatch(resp, -1); matches != nil {
		resp = strings.Replace(resp, `<card>`, "", -1)
		resp = strings.Replace(resp, `</card>`, "\n", -1)
		card = true
	}
	resp = strings.Replace(resp, "<reply>", "<button>", -1)
	resp = strings.Replace(resp, "</reply>", "</button>", -1)
	if matches := regexp.MustCompile(`<button>(.*?)</button>`).FindAllStringSubmatch(resp, -1); matches != nil {
		for _, match := range matches {
			resp = strings.Replace(resp, match[0], "", -1)
			url, text := "", "Link"
			if regexp.MustCompile(`<url>(.*?)</url>`).FindStringSubmatch(match[1]) != nil {
				url = regexp.MustCompile(`<url>(.*?)</url>`).FindStringSubmatch(match[1])[1]
			}
			if regexp.MustCompile(`<text>(.*?)</text>`).FindStringSubmatch(match[1]) != nil {
				text = regexp.MustCompile(`<text>(.*?)</text>`).FindStringSubmatch(match[1])[1]
			}
			if url != "" {
				if regexp.MustCompile(`<postback>(.*?)</postback>`).FindStringSubmatch(match[1]) != nil {
					url = regexp.MustCompile(`<postback>(.*?)</postback>`).FindStringSubmatch(match[1])[1]
				}
			}
			buttons = append(buttons, map[string]string{"text": text, "url": url})
		}
	}
	resp = strings.Replace(resp, "<title>", "<strong>", -1)
	resp = strings.Replace(resp, "</title>", "</strong>", -1)
	resp = strings.Replace(resp, "<subtitle>", "<i>", -1)
	resp = strings.Replace(resp, "</subtitle>", "</i>", -1)
	return resp, media, card
}

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/joho/godotenv"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// ----------------------Info----------------------//

type Info struct {
	Plu           string        `json:"plu"`
	Description   string        `json:"description"`
	UnitPrice     string        `json:"unitPrice"`
	Variants      []Variants    `json:"variants"`
	ProductGroups []interface{} `json:"productGroups"`
}
type Variants struct {
	Name string `json:"name"`
	Upc  string `json:"upc"`
}

//----------------------Webhook----------------------//

type Author struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	IconURL string `json:"icon_url"`
}

type Test struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}
type Thumbnail struct {
	URL string `json:"url"`
}

type Image struct {
	URL string `json:"url"`
}

type Footer struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url"`
}
type Embeds struct {
	Author      Author    `json:"author"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
	Color       int       `json:"color"`
	Fields      []Test    `json:"fields"`
	Thumbnail   Thumbnail `json:"thumbnail"`
	Image       Image     `json:"image"`
	Footer      Footer    `json:"footer"`
}
type Top struct {
	Username  string   `json:"username"`
	AvatarURL string   `json:"avatar_url"`
	Content   string   `json:"content"`
	Embeds    []Embeds `json:"embeds"`
}

var (
	mu sync.Mutex
)

func GetProxy() string {
	mu.Lock()
	file, err := os.Open("proxies.txt")
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var txtlines []string
	for scanner.Scan() {
		txtlines = append(txtlines, scanner.Text())
	}
	_ = file.Close()
	if len(txtlines) == 0 {
		panic("Please add proxies to proxies.txt")
	}
	index := rand.Intn(len(txtlines))
	mu.Unlock()
	proxy := strings.Split(txtlines[index], ":")
	proxy_url := "http://" + proxy[2] + ":" + proxy[3] + "@" + proxy[0] + ":" + proxy[1]
	return proxy_url
}

func main() {
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeout(30),
		tls_client.WithClientProfile(tls_client.Chrome_105),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithProxyUrl(GetProxy()),
	}
	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		log.Println("Error creating client: ", err)
	}
	url := "https://www.jdsports.de/campaign/Neuheiten/?facet-new=latest&sort=latest"
	req, _ := http.NewRequest("GET", url, nil)
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	url_ := ParseUrl(resp.Body)
	DataObject := GetInfo(url_, client)
	fmt.Print(DataObject)

	WebHook(DataObject, client, url_)

}

func GetIMG(url string, client tls_client.HttpClient) string {
	url = "https://www.jdsports.de" + url
	req, _ := http.NewRequest("GET", url, nil)
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	doc := html.NewTokenizer(resp.Body)
	for {
		tokenType := doc.Next()
		if tokenType == html.ErrorToken {
			break
		}
		token := doc.Token()
		if tokenType == html.StartTagToken && token.DataAtom == atom.Picture {
			for {
				tokenType := doc.Next()
				if tokenType == html.ErrorToken {
					break
				}
				token := doc.Token()
				if tokenType == html.StartTagToken && token.DataAtom == atom.Source {
					//check for <srcset> tag
					for _, a := range token.Attr {
						if a.Key == "srcset" {
							test := strings.Split(a.Val, " ")
							return test[0]
						}
					}
				}
			}
		}
	}
	return ""
}

func init() {
	err := godotenv.Load("config/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

const (
	Image_URL = "https://cdn.discordapp.com/attachments/1013517214906859540/1039155134556536894/01IHswd8_400x400.jpeg"
)

func Time() string {
	date := time.Now().Format("15:04:05")
	time := time.Now().UnixNano() / int64(time.Millisecond)
	time_final := fmt.Sprintf("%s.%d", date, time%1000)
	return time_final
}

func GetName(dataObject string) string {
	test := strings.Split(dataObject, ":")
	trim := strings.TrimSpace(test[2])
	trim2 := strings.TrimSpace(trim[1 : len(trim)-10])
	trim3 := strings.Split(trim2, "\"")
	trim4 := strings.Split(trim3[0], "-")
	return trim4[1]
}

func WebHook(dataObject string, client tls_client.HttpClient, url string) {
	img := GetIMG(url, client)
	url = "https://www.jdsports.de" + url
	var webhookURL = os.Getenv("DISCORD_WEBHOOK_URL_TEST")
	name := GetName(dataObject)
	namehyperlink := fmt.Sprintf("[%s](%s)", name, url)

	var fields []Test
	fields = append(fields, Test{
		Name:   "Price",
		Value:  "€" + "100",
		Inline: false,
	})
	test := strings.Split(dataObject, ":")
	for _, value := range test {
		if strings.Contains(value, "_jdsportsde.") {
			test := value[2 : len(value)-23]
			test2 := strings.Split(test, ",")
			size := strings.TrimSpace(test2[0])
			final := size[0 : len(size)-1]
			fields = append(fields, Test{
				Name:   "Sizes",
				Value:  final,
				Inline: true,
			})
		}
	}
	payload := &Top{
		Username:  "JD Monitor",
		AvatarURL: Image_URL,
		Content:   "",
		Embeds: []Embeds{
			{
				Color:       16777215,
				Description: namehyperlink,
				Fields:      fields,
				Thumbnail: Thumbnail{
					URL: img,
				},
				Footer: Footer{
					IconURL: Image_URL,
					Text:    "JD | Monitor " + Time(),
				},
			},
		},
	}
	payloadBuf := new(bytes.Buffer)
	_ = json.NewEncoder(payloadBuf).Encode(payload)

	if webhookURL == "" {
		fmt.Println("SET DISCORD_WEBHOOK_URL ENV VAR")
	}
	SendWebhook, err := http.NewRequest("POST", webhookURL, payloadBuf)
	if err != nil {
		fmt.Println(err)
	}
	SendWebhook.Header.Set("content-type", "application/json")

	sendWebhookRes, err := client.Do(SendWebhook)
	if err != nil {
		fmt.Print(err)
	}
	if sendWebhookRes.StatusCode != 204 {
		fmt.Printf("Webhook failed with status %d\n", sendWebhookRes.StatusCode)
	}
	defer sendWebhookRes.Body.Close()
}

func GetSize(data string) {
	test := strings.Split(data, ":")
	for _, value := range test {
		if strings.Contains(value, "_jdsportsde.") {
			test := value[2 : len(value)-23]
			test2 := strings.Split(test, ",")
			holaa := strings.TrimSpace(test2[0])
			final := holaa[0 : len(holaa)-1]
			fmt.Println(final)
		}
	}

}

func GetInfo(url string, client tls_client.HttpClient) string {
	url = "https://www.jdsports.de" + url
	req, _ := http.NewRequest("GET", url, nil)
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	doc := html.NewTokenizer(resp.Body)
	//get the var dataObject from the body
	for tokenType := doc.Next(); tokenType != html.ErrorToken; {
		token := doc.Token()
		if tokenType == html.StartTagToken && token.DataAtom == atom.Script {
			for _, attr := range token.Attr {
				if attr.Key == "type" && attr.Val == "text/javascript" {
					for tokenType := doc.Next(); tokenType != html.ErrorToken; {
						token := doc.Token()
						if strings.Contains(token.Data, "dataObject") {
							return token.Data
						}
						tokenType = doc.Next()
					}
				}
			}
		}
		tokenType = doc.Next()
	}
	return ""
}

// take the url of the first product on the page
func ParseUrl(body io.Reader) string {
	doc := html.NewTokenizer(body)
	for tokenType := doc.Next(); tokenType != html.ErrorToken; {
		token := doc.Token()
		if tokenType == html.StartTagToken {
			if token.DataAtom != atom.A {
				tokenType = doc.Next()
				continue
			}
			for _, attr := range token.Attr {
				if strings.Contains(attr.Val, "product") {
					return attr.Val
				}
			}
		}
		tokenType = doc.Next()
	}
	return ""
}

/*
1. Save the first url of the page
2. if the url is change, update new url
2. Get the dataObject from the page
3. Parse the dataObject to get the product name, price, color, size, etc
4. send the data

- Error handling
- Use Go routine to find all the product info
*/

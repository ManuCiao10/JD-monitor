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
	// "github.com/gocolly/colly/v2"
)

// ----------------------Info----------------------//

type Info struct {
	Plu         string     `json:"plu"`
	Description string     `json:"description"`
	UnitPrice   string     `json:"unitPrice"`
	Variants    []Variants `json:"variants"`
	// ProductGroups []interface{} `json:"productGroups"`
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
	mu  sync.Mutex
	URL string = "https://www.jdsports.de/campaign/Neuheiten/?facet-new=latest&sort=latest"
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

// func SaveInfo(dataObject string) {
// 	//remove last character
// 	dataObject = dataObject[:len(dataObject)-2]
// 	//remove first character
// 	dataObject = dataObject[30:]
// 	//add missing quotes
// 	dataObject = strings.ReplaceAll(dataObject, "plu", "\"plu\"")
// 	dataObject = strings.ReplaceAll(dataObject, "description", "\"description\"")
// 	dataObject = strings.ReplaceAll(dataObject, "unitPrice", "\"unitPrice\"")
// 	dataObject = strings.ReplaceAll(dataObject, "variants", "\"variants\"")
// 	dataObject = strings.ReplaceAll(dataObject, "name", "\"name\"")
// 	dataObject = strings.ReplaceAll(dataObject, "upc", "\"upc\"")
// 	dataObject = strings.ReplaceAll(dataObject, "platform", "\"platform\"")
// 	dataObject = strings.ReplaceAll(dataObject, "pageName", "\"pageName\"")
// 	dataObject = strings.ReplaceAll(dataObject, "pageType", "\"pageType\"")
// 	dataObject = strings.ReplaceAll(dataObject, "//Page Title", "")
// 	dataObject = strings.ReplaceAll(dataObject, "//Page Type", "")
// 	dataObject = strings.ReplaceAll(dataObject, "//Product Name", "")
// 	dataObject = strings.ReplaceAll(dataObject, "//Product Price", "")
// 	dataObject = strings.ReplaceAll(dataObject, "//Product Code", "")
// 	//is on sale? true/false
// 	dataObject = strings.ReplaceAll(dataObject, "//is on sale? true/false", "")

// 	fmt.Println(dataObject)

// 	//JSON Formatter & Validator
// 	var prettyJSON bytes.Buffer
// 	err := json.Indent(&prettyJSON, []byte(dataObject), "", "\t")
// 	if err != nil {
// 		log.Println("JSON parse error: ", err)
// 		return
// 	}

// 	//save data to struct Info
// 	var info []Info
// 	_ = json.Unmarshal(prettyJSON.Bytes(), &info)
// 	// if err != nil {
// 	// 	fmt.Println("JSON unmarshal error: ", err)
// 	// 	continue
// 	// }
// 	fmt.Println(info)
// }

func main() {
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeout(30),
		tls_client.WithClientProfile(tls_client.Chrome_105),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithProxyUrl(GetProxy()),
	}
	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		fmt.Println("Error creating client: ", err)
	}
	req, _ := http.NewRequest("GET", URL, nil)
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	Monitor(resp.Body)
}

func Monitor(body io.Reader) {
	FIRST_URL := ParseUrl(body)
	if FIRST_URL == "" {
		fmt.Println("No url found")
	}
	for {
		options := []tls_client.HttpClientOption{
			tls_client.WithTimeout(30),
			tls_client.WithClientProfile(tls_client.Chrome_105),
			tls_client.WithNotFollowRedirects(),
			tls_client.WithProxyUrl(GetProxy()),
		}
		client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
		if err != nil {
			fmt.Println("Error creating client: ", err)
		}
		req, err := http.NewRequest("GET", URL, nil)
		if err != nil {
			fmt.Println("Error creating request: ", err)
		}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request: ", err)
		}
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(FIRST_URL)
		URL_TMP := ParseUrl(bytes.NewReader(body))
		fmt.Println(URL_TMP)
		if (FIRST_URL != URL_TMP) && (URL_TMP != "") {
			fmt.Println("URL changed")
			//save new url
			FIRST_URL = URL_TMP
			DataObject := GetInfo(FIRST_URL, client)
			if DataObject == "" {
				fmt.Println("No data found")
			}
			//send webhook
			WebHook(DataObject, client, FIRST_URL)
		} else {
			fmt.Println("URL not changed")
		}
		time.Sleep(10 * time.Second)
	}
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
	var webhookURL = os.Getenv("DISCORD_WEBHOOK_URL")
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

- get the price
- fix check only size available

- Error handling
- Add colly to find the url etc..
*/

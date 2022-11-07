package main

import (
	"bufio"
	"encoding/json"

	// "encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	// "github.com/corpix/uarand"
)

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
	// user_agent := uarand.GetRandom()
	// req.Header.Set("User-Agent", user_agent)// req.Heade
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	url_ := ParseUrl(resp.Body)
	fmt.Println(url_)
	DataObject := GetInfo(url_, client)
	JsonParser(DataObject)
	// log.Print("Response status: ", resp.Status)

}

func JsonParser(data string) {
	// get all the names of the products in the string
	var dataObject map[string]interface{}
	json.Unmarshal([]byte(data), &dataObject)
	fmt.Println(dataObject)

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
						if tokenType == html.TextToken {
							if strings.Contains(token.Data, "dataObject") {
								return token.Data
							}
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
*/

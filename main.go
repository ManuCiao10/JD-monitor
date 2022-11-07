package main

import (
	"bufio"
	"fmt"

	// "io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	// "github.com/corpix/uarand"
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
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(bodyText))
	log.Print("Response status: ", resp.Status)

}

/*
1. Save the first product in a set or in the cache
2. if the product in the html is changed
3. then send a notification with the new product
*/

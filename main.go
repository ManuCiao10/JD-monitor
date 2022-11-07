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
	"github.com/corpix/uarand"
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
	url := "https://www.jdsports.de/campaign/Neuheiten/?facet-new=latest"
	req, _ := http.NewRequest("GET", url, nil)
	user_agent := uarand.GetRandom()
	req.Header.Set("User-Agent", user_agent)
	req.Header.Set("authority", "www.jdsports.de")
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("accept-language", "it-IT,it;q=0.9,en-US;q=0.8,en;q=0.7,de;q=0.6,fr;q=0.5")
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("cookie", `_scid=1d4bbf7e-736d-410c-ae69-4a4fbf7ee39e; AKA_A2=A; language=de; 49746=; gdprsettings2={"functional":true,"performance":true,"targeting":true}; gdprsettings3={"functional":true,"performance":true,"targeting":true}; _uetsid=585c41c05e2311ed9b6667cf329e695e; _uetvid=585c68005e2311eda4a3c7234e979ac8; _gid=GA1.2.1251931115.1667790210; mt.sc=%7B%22i%22%3A1667790210392%2C%22d%22%3A%5B%5D%7D; mt.v=2.2056514234.1667790210394; intPopSeen=1; usersCountry=CA; _ga_DR7R5FQ65N=GS1.1.1667790211.1.0.1667790211.0.0.0; _ga=GA1.1.1846370877.1667790210; _gcl_au=1.1.258432700.1667790211; _tt_enable_cookie=1; _ttp=b1fdb321-392f-44d3-9ce3-619eb1aa874d; cto_bundle=2Yl1el9kUHVMVHd3Qm9TNUs3dzNvY0IlMkI2bDY0JTJCbExjZlElMkJXZUtVekFDb1VUdmcyT3hNMXk1emJidnRRdUJrNXQlMkJja1FyOXh4Y2RFNlh2OVVONTllNyUyQlg0R2dSbUJWTlQlMkZaZERPQTRTUVZUTzBpVU9UbGoyRDBNV25rQUxES1ZES3klMkJKVXh6b2h4NiUyQmtTeTIyUkxqVEE4YWIwQlVSNmN2Z3ZOaXp5QkFFd1V2Z0hOdUJTanRDUE0lMkZmWGRpWGhnVVBiMHY; _derived_epik=dj0yJnU9eFQ0WWowQlJoeXBkYTBoOHpmVTE2UUJjSkpOOGZkVWombj1HOGIwWWM4NURUVWZtNmxyMDlMc1Z3Jm09MSZ0PUFBQUFBR05vZFlNJnJtPTEmcnQ9QUFBQUFHTm9kWU0mc3A9Mg; _pin_unauth=dWlkPVlUWTNPR1prWldVdE0yRmxZUzAwT0dWakxXSmlNVGd0WVRZMk56UmxZalJoTmpNMg; __pr.1prf=k9tkEipYTm; mt.sac_4437383=t`)
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("referer", "https://www.jdsports.de/campaign/Neuheiten/?facet-new=latest")
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="107", "Chromium";v="107", "Not=A?Brand";v="24"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	req.Header.Set("sec-fetch-dest", "document")
	req.Header.Set("sec-fetch-mode", "navigate")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("sec-fetch-user", "?1")
	req.Header.Set("upgrade-insecure-requests", "1")
	// req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36")
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

	// fmt.Printf("%s\n", bodyText)
	// resp, err := client.Do(req)
	// if err != nil {
	// 	log.Println("Error sending request: ", err)
	// }
	// defer resp.Body.Close()
	// body, _ := io.ReadAll(resp.Body)
	log.Print("Response status: ", resp.Status)
	// log.Print("Response body: ", string(body))





}

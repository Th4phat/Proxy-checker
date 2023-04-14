package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Print("== Http Proxy checker ==\n\n")
		fmt.Printf("usage:%s [input file] [output file]", os.Args[0])
		return
	}
	var corndog []string
	mypp := getmypp()
	proxies := ppl(os.Args[1])
	results := make(chan string)

	for _, proxy := range proxies {
		go saygex(proxy, mypp, results)
	}

	for i := 0; i < len(proxies); i++ {
		result := <-results
		if result != "" {
			corndog = append(corndog, result)
		}
	}
	err := ioutil.WriteFile(os.Args[2], []byte(strings.Join(corndog, "\n")), 0644)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Print("Finished yipee!")
}

func saygex(proxy, mypp string, results chan string) {
	proxyuri := parsUrl(proxy)
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyuri),
		},
		Timeout: 30 * time.Second,
	}
	resp, err := client.Get("https://ipinfo.io/ip")
	if err != nil {
		results <- ""
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		results <- ""
		return
	}

	ip := strings.TrimSpace(string(body))
	if ip != mypp {
		fmt.Println("Proxy", proxy, "working!")
		results <- proxy
		return
	}

	results <- ""
	return
}

func getmypp() string {
	resp, err := http.Get("https://ipinfo.io/ip")
	if err != nil {
		fmt.Println("Error getting public IP address:", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error getting public IP address:", err)
		return ""
	}

	return strings.TrimSpace(string(body))
}

func parsUrl(urlStr string) *url.URL {
	parsed, err := url.Parse(strings.TrimRight("http://"+urlStr, "\r"))
	if err != nil {
		panic(fmt.Sprintf("failed to parse URL %q: %s", urlStr, err))
	}
	return parsed
}
func ppl(uncheck_file string) []string {
	var unchecklist []string
	file, err := os.Open(uncheck_file)
	if err != nil {
		fmt.Println("Error:", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		unchecklist = append(unchecklist, scanner.Text())
	}
	rmdupe_uncheck_list := noduping(unchecklist)
	return rmdupe_uncheck_list
}
func noduping(proxies []string) []string {
	uniqueProxies := make(map[string]struct{})
	for _, proxy := range proxies {
		uniqueProxies[proxy] = struct{}{}
	}
	var result []string
	for proxy := range uniqueProxies {
		result = append(result, proxy)
	}
	return result
}

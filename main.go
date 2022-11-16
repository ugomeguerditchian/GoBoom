package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func proxy_getter() []string {
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
	//convert the body to string
	proxy_list := string(body)
	//split the string by new line
	proxy_list = strings.Replace(proxy_list, "\r", "", -1)
	var proxy_list_final []string = strings.Split(proxy_list, "\n")
	return proxy_list_final

}

func handler_proxy(domain string, proxy string) string {
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
	if resp.Status == "503 Too many open connections" {
		return "error"
	}
	if resp.Status == "401 Unauthorized" {
		return "error"
	}
	if resp.Status == "409 Conflict" {
		return "error"
	}
	if resp.Status == "404 Not Found" {
		return "error"
	}
	if resp.Status == "502 Bad Gateway" {
		return "error"
	}
	if resp.Status == "504 Gateway Timeout" {
		return "error"
	}
	if resp.Status == "407 Proxy Authentication Required" {
		return "error"
	}
	//if resp.StatusCode exist
	if resp.StatusCode != 0 {
		return resp.Status
	} else {
		return "error"
	}
}

func handler_proxy_thread(domain string, proxy []string) string {
	//just connect to the website and check the status code
	//if handler_proxy return error check the next proxy
	for true {
		for i := 0; i < len(proxy); i++ {
			var result = handler_proxy(domain, proxy[i])
			if result != "error" {
				return result
			}
		}
	}
	return "error"
}

// create a chunk function to split the proxy list and return a list of list string with a len of total_to_divided
func chunkSlice(s []string, total_to_divided int) [][]string {
	var divided [][]string
	var chunk float32 = float32(len(s) / total_to_divided)
	var start int = 0
	var end int = 0
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
func main() {
	//ask for a domain name
	fmt.Println("Enter a domain name: ")
	var domain string
	fmt.Scanln(&domain)
	//ask for thread number
	fmt.Println("Enter the number of thread: ")
	var thread int
	fmt.Scanln(&thread)
	//call the proxy_getter function
	proxy_list := proxy_getter()
	//create chunk of proxy list for each thread
	var chunk_list [][]string = chunkSlice(proxy_list, thread)
	//create a channel to store the result
	var result_channel = make(chan string)
	//create a thread for each chunk
	quit := make(chan bool)
	for true {
		for i := 0; i < thread; i++ {
			//fmt.Print("Thread ", i, " is running")
			go func(i int) {
				for {
					select {
					case <-quit:
						return
					default:
						result_channel <- handler_proxy_thread(domain, chunk_list[i])
					}
				}
			}(i)
		}
		//get the result from the channel
		for i := 0; i < thread; i++ {
			var result = <-result_channel
			if result != "error" {
				fmt.Println(result)
				//wait 0.5 second
				time.Sleep(200 * time.Millisecond)
				break
			} else {
				fmt.Println(result)
				quit <- true
			}
		}
	}

}

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
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
	var real_proxy_list []string
	//test each proxy in the list
	for i := 0; i < len(proxy_list_final); i++ {
		//create a proxy url
		proxy_url, err := url.Parse("http://" + proxy_list_final[i])
		if err != nil {
			log.Fatal(err)
		}
		//create a transport
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxy_url),
		}
		//create a client
		client := &http.Client{
			Transport: transport,
		}
		//create a request
		req, err := http.NewRequest("GET", "https://www.google.com", nil)
		if err != nil {
			log.Fatal(err)
		}
		//send the request
		resp, err := client.Do(req)
		//if the status code is 200, add the proxy to the list
		if resp.StatusCode == 200 {
			real_proxy_list = append(real_proxy_list, proxy_list_final[i])
		}
	}

	return real_proxy_list

}

func handler_proxy(domain string, proxy string) string {
	//just connect to the website and check the status code
	//use proxy to connect
	//format of proxy is ip:port
	//format of domain is domain.com
	proxyUrl, err := url.Parse("http://" + proxy)
	if err != nil {
		log.Fatal(err)
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
	//fmt.Print(resp.Status)
	return resp.Status
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
	//call the handler function
	//handler(domain)
	//call the proxy_getter function
	proxy_list := proxy_getter()
	//create chunk of proxy list for each thread
	var chunk int = len(proxy_list) / thread
	var proxy_list_chunk [][]string
	for i := 0; i < thread; i++ {
		proxy_list_chunk = append(proxy_list_chunk, proxy_list[i*chunk:(i+1)*chunk])
	}
	//create a channel
	c := make(chan string)
	for i := 0; i < thread; i++ {
		go func() {
			for _, proxy := range proxy_list_chunk[i] {
				c <- handler_proxy(domain, proxy)
			}
		}()
	}
	for i := 0; i < len(proxy_list); i++ {
		fmt.Println(<-c)
	}

}

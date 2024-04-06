package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

const (
	maxHops     = 64
	packetSize  = 32
	defaultPort = 33434
	timeout     = 5 * time.Second
)

func main() {
	hostname := os.Args[1]
	if hostname == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s <hostname>\n", os.Args[0])
		return
	}

	ips, err := net.LookupIP(hostname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not get IPs: %v\n", err)
		return
	}
	ip := ips[0]

	fmt.Printf(
		"traceroute to %s (%s), %d hops max, %d byte packets",
		hostname,
		ip,
		maxHops,
		packetSize,
	)

	// Create socket to recieve ICMP packets
	sendSock, err := unix.Socket(unix.AF_INET, unix.SOCK_RAW, unix.IPPROTO_ICMP)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not create socket: %v\n", err)
		return
	}
	// Create socket to send UDP packets
	recvSock, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_UDP)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not create socket: %v\n", err)
		return
	}
	timeout := unix.Timeval{Sec: int64(timeout), Usec: 0}
	err = unix.SetsockoptTimeval(recvSock, unix.SOL_SOCKET, unix.SO_RCVTIMEO, &timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not set timeout: %v\n", err)
		return
	}
}

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

	ip := ips[0].To4()

	fmt.Printf(
		"traceroute to %s (%s), %d hops max, %d byte packets\n",
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

	// Set timeout for recvSock
	timeout := unix.Timeval{Sec: int64(timeout), Usec: 0}
	err = unix.SetsockoptTimeval(recvSock, unix.SOL_SOCKET, unix.SO_RCVTIMEO, &timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not set timeout: %v\n", err)
		return
	}

	defer unix.Close(sendSock)
	defer unix.Close(recvSock)
	begin := time.Now()
	destAddr := &unix.SockaddrInet4{Port: defaultPort, Addr: [4]byte{ip[0], ip[1], ip[2], ip[3]}}

	for {
		begin = time.Now()
		// Set ttl
		err = unix.SetsockoptInt(sendSock, unix.IPPROTO_IP, unix.IP_TTL, maxHops)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not set TTL: %v\n", err)
			return
		}

		err = unix.Sendto(
			sendSock,
			[]byte{0},
			0,
			destAddr,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not send packet: %v\n", err)
			return
		}

		var buf [packetSize]byte
		_, _, err = unix.Recvfrom(recvSock, buf[:], 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not receive packet: %v\n", err)
			return
		}

		end := time.Now()
		duration := end.Sub(begin)

		// Print the message
		fmt.Printf(string(buf[:]), duration)
		break
	}
}

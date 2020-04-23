package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const PingTimeout = time.Second
const DataLen = 56

type Pinger struct {
	hostname string
	ipAddr   *net.IPAddr
	icmpID   int
	seqNum   int
}

func NewPinger(addr string) (*Pinger, error) {
	ipAddr, err := net.ResolveIPAddr("ip4", addr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &Pinger{
		hostname: addr,
		ipAddr:   ipAddr,
		icmpID:   os.Getpid(),
		seqNum:   0,
	}, nil
}

func (p *Pinger) StartPing() {
	fmt.Printf("PING %s (%s): 56 data bytes\n", p.hostname, p.ipAddr)
	connection, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer connection.Close()

	interval := time.NewTicker(time.Second)
	defer interval.Stop()

	for {
		select {
		case <-interval.C:
			p.pingWithTimeout(connection)
			p.seqNum++
		}
	}

}

func (p *Pinger) AddSigtermHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		p.printStatistics()
		os.Exit(0)
	}()
}

func (p *Pinger) printStatistics() {
	fmt.Printf("\n--- %s ping statistics ---\n", p.hostname)
}

func (p *Pinger) pingWithTimeout(conn *icmp.PacketConn) {
	ctx, cancel := context.WithTimeout(context.Background(), PingTimeout)
	defer cancel()

	resultChan := make(chan bool)
	go p.ping(conn, resultChan)

	select {
	case <-resultChan:
		return
	case <-ctx.Done():
		fmt.Printf("Request timeout for icmp_seq %d\n", p.seqNum)
	}

}

func (p *Pinger) ping(conn *icmp.PacketConn, resultChan chan bool) error {
	sendTime := time.Now()

	err := p.sendICMP(conn, p.seqNum)

	readBuf := make([]byte, 1500)
	numBytes, _, err := conn.ReadFrom(readBuf)
	if err != nil {
		return err
	}
	resultChan <- true

	recvTime := time.Now()

	readBuf = readBuf[:numBytes]

	var rtt float32
	rtt = float32(recvTime.Sub(sendTime).Microseconds()) / 1000

	fmt.Printf("%d bytes from %s: icmp_seq=%d time=%.3f ms\n", numBytes, p.ipAddr, p.seqNum, rtt)

	return nil
}

// Sends ICMP echo packet and returns seqNum.
func (p *Pinger) sendICMP(conn *icmp.PacketConn, seqNum int) error {

	data := make([]byte, DataLen)
	for idx, _ := range data {
		data[idx] = '0'
	}

	icmpBody := &icmp.Echo{
		ID:   os.Getpid(),
		Seq:  seqNum,
		Data: data,
	}

	icmpMsg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: icmpBody,
	}

	msgBytes, err := icmpMsg.Marshal(nil)
	if err != nil {
		return err
	}

	conn.WriteTo(msgBytes, p.ipAddr)

	return nil
}

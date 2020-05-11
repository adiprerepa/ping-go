package main

import (
	"flag"
	"fmt"
	"github.com/adiprerepa/ping-go/src/pkg/agent"
	"net"
	"os"
	"os/signal"
	"time"
)

var howToUse = `
Ping - Written in Golang 1.13.
Author: Aditya Prerepa
Check Me Out: http://adiprerepa.github.io/

Usage:

	ping [-c count] [-w deadline] [-t timeout] [-p pad pattern] [-q quiet output] [-i interval] [-ttl max time to live] destination

Some Examples:	
	
	# Ping My Website Forever
	sudo ./ping adiprerepa.github.io
	
	# Ping My Website 10 Times
	sudo ./ping -c 5 adiprerepa.github.io

	# Give the ping a deadline to complete (in seconds)
	sudo ./ping -w 10s adiprerepa.github.io
	
	# Give the ping a pad pattern (Works only in special cases)
	sudo ./ping -p 01010101010 adiprerepa.github.io

	# Make the Ping shut up (No output until completion)
	sudo ./ping --quiet_output adiprerepa.github.io

	# Give the ping a timeout (in seconds)
	sudo ./ping -t 5s adiprerepa.github.io

	# Give the ping an interval (time in between pings, in seconds)
	sudo ./ping -i 1s adiprerepa.github.io

	# Give the ping a max Time to live
	sudo ./ping -ttl 100 adiprerepa.github.io
	
You can ping Ipv6, set a max TTL, and much more. 
Unit Tests cover all the core functions.
Happy Pinging!

NOTE: Ping needs to be run in sudo mode for ICMP to work.
`

func main() {
	timeout := flag.Duration("t", time.Second*100000, "")
	deadline := flag.Duration("w", time.Second, "")
	count := flag.Int("c", int(^uint(0) >> 1), "")
	pad := flag.String("p", "00000000", "")
	ttl := flag.String("ttl", "255", "")
	quietOutput := flag.Bool("quiet_output", false, "")
	interval := flag.Duration("i", time.Second, "")
	flag.Usage = func() {
		fmt.Printf(howToUse)
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		return
	}
	var ip string
	pingDestination := flag.Arg(0)
	res, err := agent.IsIPv4(pingDestination)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(1)
	}
	if res {
		addr, err := net.LookupIP(pingDestination)
		if err != nil {
			fmt.Printf("LOOKUP ERR :%s\n", err.Error())
		}
		ip = addr[0].String()
	} else {
		ip = pingDestination
	}
	options := &agent.PresentOptions{}
	_ = options.ParseCountFlag(*count)
	_ = options.ParseTimeoutFlag(*timeout)
	_ = options.ParseDeadlineFlag(*deadline)
	err = options.ParsePadding(*pad)
	if err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}
	_ = options.ParseIntervalFlag(*interval)
	_ = options.ParseIPAddress(ip)
	_ = options.ParseTTL(*ttl)
	pinger := agent.BuildPinger(options)
	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt)
	go func() {
		for range interruptChannel {
			pinger.Stop()
		}
	}()
	if !*quietOutput {
		pinger.OnEchoComplete = func(p *agent.PingPacket, exceededTTL bool) {
			fmt.Printf("%d Bytes from %s: icmp_seq=%d time=%v ttl=%v exceeded_max_ttl:%v\n", p.NumberOfBytes, p.DestinationAddress, p.ICMPSequenceNumber, p.RoundTripTime,
				p.TimeToLive, exceededTTL)
		}
	}
	pinger.OnProcessComplete = func(p *agent.CompletedPingStatistics) {
		fmt.Printf("\n-----------ping statistics-----------\n")
		fmt.Printf("%d transmitted packets, %d received packets, %d lost packets, %v%% packet recovery, %v%% packet loss\n",
			p.PacketsReceived + p.PacketsLost, p.PacketsReceived, p.PacketsLost, p.PercentReceived, p.PercentLost)
		fmt.Printf("packets exceeded max ttl: %v avg round trip: %v\n", p.ExceededTTL, p.AverageRTT)
	}

	fmt.Println("Aditya's Pinger!")
	fmt.Printf("PING: %s:\n", ip)
	pinger.Driver()
}

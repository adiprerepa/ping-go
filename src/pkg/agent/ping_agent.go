package agent

import (
	"errors"
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"net"
	"sync"
	"syscall"
	"time"
)

// PingerAgent is the agent which most of the
// sending and receiving/networking functions operate on.
type PingerAgent struct {
	// command-line options
	options PresentOptions
	// for statistics
	packetsSent int
	packetsRecieved int
	// When a user interrupts the program with ctrl+c,
	// a signal is sent.
	stopPing chan bool
	// to make sure the packets we receive track with the packets we send.
	packetId int
	packetTracker int64
	sequence int
	// our logPacket() function logs the packet's individual RTT in roundTripTimes[]
	roundTripTimes []time.Duration
	// time to live
	maxTTL int
	numExceededTTL int
	// Callbacks to the main function to print statistics.
	OnEchoComplete func(p *PingPacket, exceededTTL bool)
	OnProcessComplete func (c * CompletedPingStatistics)
}

// PingPacket represents an individual ICMP packet.
type PingPacket struct {
	// Most of these fields are given to us by the receieved packet.
	RoundTripTime      time.Duration
	DestinationAddress string
	ICMPSequenceNumber int
	TimeToLive         int
	NumberOfBytes      int
	data               []byte
}

// The callback OnProcessComplete() takes in this struct
// to print all of the statistics.
type CompletedPingStatistics struct {
	AverageRTT time.Duration
	PacketsReceived int
	PacketsLost int
	PercentReceived float64
	PercentLost float64
	Destination string
	ExceededTTL int
}

// Driver is the basically the main function, this is what
// orchestrates the sending and receiving.
func (p* PingerAgent) Driver() {
	var connection *icmp.PacketConn
	if p.options.isIpv4 {
		// Listen for Incoming ipv4 ICMP packets
		if connection = p.listenICMP("ip4:icmp"); connection != nil {
			connection.IPv4PacketConn().SetControlMessage(ipv4.FlagTTL, true)
		} else {
			return
		}
	} else {
		// Listen for Incoming ipv6 ICMP Packets
		if connection = p.listenICMP("ip6:ipv6-icmp"); connection != nil {
			connection.IPv6PacketConn().SetControlMessage(ipv6.FlagHopLimit, true)
		} else {
			return
		}
	}
	// When the program exists, clean up.
	defer connection.Close()
	defer p.setStatisticsHandler()
	// Used to let goroutines finish when the program is interrupted/finished (mutex lock)
	var waitGroup sync.WaitGroup
	// we send packets back from ReceiveICMPPacket() in this channel.
	packetChannel := make(chan *PingPacket, 5)
	defer close(packetChannel)
	waitGroup.Add(1)
	// Receive ICMP Packets on a separate goroutine.
	go p.ReceiveICMPPacket(connection, packetChannel, &waitGroup)
	err := p.SendICMPPacket(connection)
	if err != nil {
		fmt.Printf("Could not send the ICMP packet: %s\n", err.Error())
	}
	// Set Tickers which have channels (reactive), that
	// go to the next iteration every custom-set interval
	timeoutTicker := time.NewTicker(p.options.timeout)
	intervalTicker := time.NewTicker(p.options.interval)
	defer timeoutTicker.Stop()
	defer intervalTicker.Stop()
	for {
		select {
		// Ctrl+C
		case <- p.stopPing:
			waitGroup.Wait()
			return
		// Packet Timeout exceeded
		case <- timeoutTicker.C:
			waitGroup.Wait()
			close(p.stopPing)
			return
		// every time the intervalTicker ticks, we send/receive another packet
		case <- intervalTicker.C:
			if p.packetsSent > 0 && p.packetsSent >= p.options.count  {
				continue
			}
			err = p.SendICMPPacket(connection)
			if err != nil {
				fmt.Printf("ERROR: %s\n", err.Error())
			}
		// We received a packet from packetChannel, we log it for stats
		case receivedPacket := <- packetChannel:
			err := p.logPacket(receivedPacket)
			if err != nil {
				fmt.Printf("ERROR: %s\n", err.Error())
			}
		}
		// If we reached the user-specified ount
		if p.options.count > 0 && p.packetsRecieved >= p.options.count {
			close(p.stopPing)
			waitGroup.Wait()
			return
		}
	}
}

// GetPingStatistics() Calculates statistics to display to the user
// based on round trip time, time to live, and the packet yield.
func (p* PingerAgent) GetPingStatistics() *CompletedPingStatistics{
	percentReceived := float64(p.packetsRecieved) / float64(p.packetsSent) * 100
	percentLost := float64(p.packetsSent - p.packetsRecieved) / float64(p.packetsSent) * 100
	var total time.Duration
	for _, rtt := range p.roundTripTimes {
		total += rtt
	}
	// average RTT
	avg := total / time.Duration(len(p.roundTripTimes))
	return &CompletedPingStatistics{
		AverageRTT:      avg,
		PacketsReceived: p.packetsRecieved,
		PacketsLost:     p.packetsSent - p.packetsRecieved,
		PercentReceived: percentReceived,
		PercentLost:     percentLost,
		ExceededTTL: 	 p.numExceededTTL,
		Destination:     p.options.ipAddress,
	}
}

// ReceiveICMPPacket is run as a goroutine and sends packets back via a packetChannel.
func (p *PingerAgent) ReceiveICMPPacket(connection *icmp.PacketConn, packetChannel chan <- *PingPacket, group *sync.WaitGroup) {
	defer group.Done()
	for {
		select {
		// Keyboard Interrupt (Ctrl+C)
		case <-p.stopPing:
			return
		default:
			var err error
			// We need to Read the packet by whatever the -w argument was
			connection.SetReadDeadline(time.Now().Add(p.options.deadline))
			if err != nil {
				fmt.Printf("Couldn't set the Read Deadline on the icmp Packet Connection: %s\n", err.Error())
			}
			receivedBytes := make([]byte, 1024)
			var numberOfBytes, timeToLive int
			var source net.IP
			if p.options.isIpv4 {
				var message *ipv4.ControlMessage
				// actually receive the message
				numberOfBytes, message, _, err = connection.IPv4PacketConn().ReadFrom(receivedBytes)
				if message != nil {
					// build up some of the packet
					timeToLive = message.TTL
					source = message.Src
				}

			} else {
				// ipv6 of ^
				var message *ipv6.ControlMessage
				numberOfBytes, message, _, err = connection.IPv6PacketConn().ReadFrom(receivedBytes)
				if message != nil {
					timeToLive = message.HopLimit
				}
			}
			if err != nil {
				if networkError, status := err.(*net.OpError); status {
					// if the network error is timeout we are ok
					if !networkError.Timeout() {
						close(p.stopPing)
						return
					} else {
						continue
					}
				}
			}
			// send the packet back to the channel.
			packetChannel <- &PingPacket{
				data:          receivedBytes,
				TimeToLive:    timeToLive,
				NumberOfBytes: numberOfBytes,
				DestinationAddress: source.String(),
			}
		}
	}
}

// SendICMPPacket sends an echo packet, similar to those in pings,
func (p* PingerAgent) SendICMPPacket(connnection *icmp.PacketConn) error {
	var packetType icmp.Type
	if p.options.isIpv4 {
		packetType = ipv4.ICMPTypeEcho
	} else {
		packetType = ipv6.ICMPTypeEchoRequest
	}
	destination, err := net.ResolveIPAddr("ip", p.options.ipAddress)
	if err != nil {
		fmt.Printf("ERROR: Could not Resolve IP: %s\n", p.options.ipAddress)
		return err
	}
	// Append the Tracker to the Packet Data - so we can trace
	packetData := append(TimeToBytes(time.Now()), IntToBytes(p.packetTracker)...)
	// Populate the ICMP Packet
	packetMessage := &icmp.Message{
		Type:     packetType,
		Code:     0,
		Checksum: 0,
		Body:     &icmp.Echo{
			ID:   p.packetId,
			Seq:  p.sequence,
			Data: packetData,
		},
	}
	// Marshal the packet into bytes
	packetBytes, err := packetMessage.Marshal(nil)
	if err != nil {
		return err
	}
	for {
		_, err := connnection.WriteTo(packetBytes, destination)
		if networkErr, status := err.(*net.OpError); status && err != nil {
			if networkErr.Err == syscall.ENOBUFS {
				// it always sends as long as it is an ENOBUFS error
				continue
			}
		}
		p.sequence++
		p.packetsSent++
		break
	}
	return nil
}

// http://www.networksorcery.com/enp/protocol/icmpv6.htm - protocol number 58 for ipv6
func (p *PingerAgent) logPacket(received *PingPacket) error {
	tripCompleted := time.Now()
	var protocol int
	if p.options.isIpv4 {
		protocol = 1
	} else {
		protocol = 58
	}
	// get the message from the bytes
	message, err := icmp.ParseMessage(protocol, received.data)
	if err != nil {
		return err
	}
	// if it's not an echo response
	if message.Type != ipv4.ICMPTypeEchoReply && message.Type != ipv6.ICMPTypeEchoReply {
		return nil
	}
	switch receivedType := message.Body.(type) {
	case *icmp.Echo:
		if p.packetId != receivedType.ID {
			return nil
		}
		// we need all 16 bytes because we receive the time and tracker as well.
		if len(receivedType.Data) < 16 {
			return errors.New(fmt.Sprintf("Bad Data, %d %v", len(receivedType.Data), receivedType.Data))
		}
		// get the timestamp and tracker from the data
		packetTracker := BytesToInt(receivedType.Data[8:])
		packetSentTimestamp := BytesToTime(receivedType.Data[:8])
		if packetTracker != p.packetTracker {
			// not our packet
			return nil
		}
		// rtt = packet_recv_time - packet_sent_tiem
		received.RoundTripTime = tripCompleted.Sub(packetSentTimestamp)
		received.ICMPSequenceNumber = receivedType.Seq
		p.packetsRecieved++
	default:
		return errors.New(fmt.Sprintf("bad ICMP reply"))
	}
	// add the time to the slice for averaging
	p.roundTripTimes = append(p.roundTripTimes, received.RoundTripTime)
	exceeded := false
	if received.TimeToLive > p.maxTTL {
		exceeded = true
	}
	// initiate the callback to print the stats
	onCompleteHandler := p.OnEchoComplete
	if onCompleteHandler != nil {
		onCompleteHandler(received, exceeded)
	}
	return nil
}

// listenICMP() Listens for ICMP Packets - Note: ping needs to be run in sudo mode.
func (p* PingerAgent) listenICMP(networkProtocol string) *icmp.PacketConn {
	connection, err := icmp.ListenPacket(networkProtocol, "")
	if err != nil {
		fmt.Printf("Could not listen for ICMP packets for %s: %s\n", p.options.ipAddress, err.Error())
		fmt.Print("Did you forget to run in sudo mode?\n")
		close(p.stopPing)
		return nil
	}
	return connection
}

// setStatisticsHandler initiates the statistics callback
func (p* PingerAgent) setStatisticsHandler() {
	statsHandler := p.OnProcessComplete
	if statsHandler != nil {
		stats := p.GetPingStatistics()
		statsHandler(stats)
	}
}

// Stop notifies all goroutines to stop through the stopPing channel.
func (p *PingerAgent) Stop() {
	close(p.stopPing)
}
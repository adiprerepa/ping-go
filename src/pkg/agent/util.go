package agent

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"net"

	//"net"
	"strings"
	"time"
)

// Utility Functions

// BuildPinger builds the pinger pased on the command line options
func BuildPinger(options *PresentOptions) *PingerAgent {
	tracker := rand.New(rand.NewSource(time.Now().UnixNano()))
	return &PingerAgent{
		options:           *options,
		packetsSent:       0,
		packetsRecieved:   0,
		stopPing:          make(chan bool),
		packetId:          tracker.Intn(math.MaxInt16),
		packetTracker:     tracker.Int63n(math.MaxInt64),
		numExceededTTL:    0,
		maxTTL: 	       options.timeToLive,
		sequence:          0,
	}
}

func BytesToInt(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b))
}

func IntToBytes(tracker int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(tracker))
	return b
}

func TimeToBytes(t time.Time) []byte {
	nsec := t.UnixNano()
	b := make([]byte, 8)
	for i := uint8(0); i < 8; i++ {
		b[i] = byte((nsec >> ((7 - i) * 8)) & 0xff)
	}
	return b
}

func BytesToTime(b []byte) time.Time {
	var nsec int64
	for i := uint8(0); i < 8; i++ {
		nsec += int64(b[i]) << ((7 - i) * 8)
	}
	return time.Unix(nsec/1000000000, nsec%1000000000)
}

func IsIPv4(address string) (bool, error) {
	if strings.Count(address, ":") >= 2 {
		return false, nil
	}
	ips, err := net.LookupIP(address)
	if len(ips) == 0 {
		fmt.Printf("Invalid Host. %s\n", err.Error())
		return false, err
	}
	if net.ParseIP(ips[0].String()) != nil {
		return true, nil
	}
	fmt.Print(err)
	return false, err
	//return net.ParseIP(address) != nil
}

func isIPv6(address string) bool {
	return strings.Count(address, ":") >= 2
}
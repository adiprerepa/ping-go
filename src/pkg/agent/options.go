package agent

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

// interval in ms
type PresentOptions struct {
	count		         int
	interval 		     time.Duration
	timeout 			 time.Duration
	deadline 			 time.Duration
	ipAddress 			 string
	isIpv4				 bool
	logOutput            bool
	timeToLive           int
	padding 	 		 string
}

// In a more advanced version, these functions would have safeguards from
// letting corrupted values enter the data structure.

func (p *PresentOptions) ParseCountFlag(option int) error {
	p.count = option
	return nil
}

func (p *PresentOptions) ParseIntervalFlag(option time.Duration) error {
	p.interval = option
	return nil
}

func (p *PresentOptions) ParseTimeoutFlag(option time.Duration) error {
	p.timeout = option
	return nil
}

func (p *PresentOptions) ParseDeadlineFlag(option time.Duration) error {
	p.deadline = option
	return nil
}


func (p *PresentOptions) ParseIPAddress(option string) error {
	var err error
	var res bool
	if isIPv6(option) {
		p.ipAddress = option
		p.isIpv4 = false
		return nil
	} else if res, err = IsIPv4(option); err == nil && res {
		p.ipAddress = option
		p.isIpv4 = true
		return nil
	}
	if err != nil {
		return err
	}
	return errors.New(fmt.Sprintf("Error: %s is neither a valid ipv4 or ipv6 address", option))
}

func (p *PresentOptions) SetLogOption(option bool) error {
	p.logOutput = option
	return nil
}

func (p *PresentOptions) ParseTTL(option string) error {
	result, err := strconv.Atoi(option)
	if err != nil {
		return err
	}
	if result < 0 || result > 256 {
		return errors.New("ttl cannot be > 256")
	}
	p.timeToLive = result
	return nil
}

func (p *PresentOptions) ParsePadding(option string) error {
	for _, char := range option {
		// padding needs to be binary
		if char != '0' && char != '1' {
			return errors.New("Invalid Padding Format. -p option needs to be only 0s and 1s.\n")
		}
	}
	p.padding = option
	return nil
}
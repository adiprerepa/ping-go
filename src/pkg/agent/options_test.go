package agent

import (
	"errors"
	"testing"
	"time"
)

// Tests for the options parser
// found (2) bugs testing.

func TestPresentOptions_ParseCountFlag(t *testing.T) {
	tests := []struct {
		desc string
		options PresentOptions
		inCount int
		expected error
	}{
		{
			desc: "replace-value",
			options: PresentOptions{
				count: 5,
			},
			inCount: 10,
			expected: nil,
		},
		{
			desc: "same-value",
			options: PresentOptions{
				count: 0,
			},
			inCount: 0,
			expected: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if err := tt.options.ParseCountFlag(tt.inCount); err != nil || tt.inCount != tt.options.count {
				t.Errorf("%s: expected count & error: %v %v, got count & error: %v %v",
					tt.desc, tt.inCount, tt.expected, tt.options.count, err)
			}
		})
	}
}

func TestPresentOptions_ParseIntervalFlag(t *testing.T) {
	tests := []struct {
		desc string
		options PresentOptions
		inDuration time.Duration
		expected error
	}{
		{
			desc: "replace-time",
			options: PresentOptions{
				interval: time.Duration(500),
			},
			inDuration: time.Duration(50),
			expected: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t* testing.T) {
			if err := tt.options.ParseIntervalFlag(tt.inDuration); err != tt.expected || tt.inDuration != tt.options.interval {
				t.Errorf("%s: expected duration & error %v %v, got duration & error %v %v",
					tt.desc, tt.inDuration, tt.expected, tt.options.interval, err)
			}
		})
	}
}

func TestPresentOptions_ParseTimeoutFlag(t *testing.T) {
	tests := []struct {
		desc string
		options PresentOptions
		inDuration time.Duration
		expected error
	}{
		{
			desc: "replace-time",
			options: PresentOptions{
				timeout: time.Duration(500),
			},
			inDuration: time.Duration(50),
			expected: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t* testing.T) {
			if err := tt.options.ParseTimeoutFlag(tt.inDuration); err != tt.expected || tt.inDuration != tt.options.timeout {
				t.Errorf("%s: expected duration & error %v %v, got duration & error %v %v",
					tt.desc, tt.inDuration, tt.expected, tt.options.timeout, err)
			}
		})
	}
}

func TestPresentOptions_ParseDeadlineFlag(t *testing.T) {
	tests := []struct {
		desc string
		options PresentOptions
		inDuration time.Duration
		expected error
	}{
		{
			desc: "replace-time",
			options: PresentOptions{
				deadline: time.Duration(500),
			},
			inDuration: time.Duration(50),
			expected: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t* testing.T) {
			if err := tt.options.ParseDeadlineFlag(tt.inDuration); err != tt.expected || tt.inDuration != tt.options.deadline {
				t.Errorf("%s: expected duration & error %v %v, got duration & error %v %v",
					tt.desc, tt.inDuration, tt.expected, tt.options.deadline, err)
			}
		})
	}
}

func TestPresentOptions_ParseIPAddress(t *testing.T) {
	tests := []struct {
		desc string
		inIP string
		options PresentOptions
		expectedStatus bool
		expected error
	}{
		{
			desc: "empty-ip",
			inIP: "",
			options: PresentOptions{
				ipAddress: "",
				isIpv4: false,
			},
			expectedStatus: false,
			expected: errors.New(""),
		},
		{
			desc: "invalid-ip",
			inIP: "fooey-Im-invalid",
			options: PresentOptions{
				ipAddress: "0.0.0.0",
				isIpv4:    false,
			},
			expectedStatus: false,
			expected: errors.New("fooey-Im-invalid is neither a valid ipv4 or ipv6 address"),
		},
		{
			desc: "valid-ipv4",
			inIP: "192.99.20.3",
			options: PresentOptions{
				ipAddress: "0.0.0.0",
				isIpv4:    false,
			},
			expectedStatus: true,
			expected: nil,
		},
		{
			desc: "valid-ipv6",
			inIP: "2001:4860:4860::8888",
			options: PresentOptions{
				ipAddress: "0.0.0.0.",
				isIpv4:    true,
			},
			expectedStatus: false,
			expected: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			err := tt.options.ParseIPAddress(tt.inIP)
			if  ((err == nil && tt.expected != nil ) || (err != nil && tt.expected == nil)) || tt.options.isIpv4 != tt.expectedStatus || tt.options.ipAddress != tt.inIP {
				// if we have an error expected it wont be set in the struct so dont check
				if !(tt.expected != nil && err != nil) {
					t.Errorf("%s: expected ip %v status %v error %v, got ip %v status %v error %v", tt.desc, tt.inIP, tt.expectedStatus,
						tt.expected, tt.options.ipAddress, tt.options.isIpv4, err)
				}
			}
		})
	}
}

func TestPresentOptions_ParseTTL(t *testing.T) {
	tests := []struct {
		desc string
		inTTL string
		options PresentOptions
		expectedTTL int
		expectedErr error
	}{
		{
			desc: "empty",
			inTTL: "",
			options: PresentOptions{},
			expectedTTL: 0,
			expectedErr: errors.New(""),
		},
		{
			desc: "bad-string",
			inTTL: "foo-foo-foo",
			options: PresentOptions{},
			expectedErr: errors.New(""),
		},
		{
			desc: "valid-ttl",
			inTTL: "146",
			expectedTTL: 146,
			options: PresentOptions{},
			expectedErr: nil,
		},
		{
			desc: "negative-ttl",
			inTTL: "-250",
			expectedTTL: -250,
			options: PresentOptions{
				timeToLive: 0,
			},
			expectedErr: errors.New(""),
		},
		{
			desc: "ttl-above-256",
			inTTL: "1055",
			expectedTTL: 1055,
			options: PresentOptions{},
			expectedErr: errors.New("ttl cannot be > 256"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if err := tt.options.ParseTTL(tt.inTTL); (err != nil && tt.expectedErr == nil) || (err == nil && tt.expectedErr != nil) || tt.options.timeToLive != tt.expectedTTL {
				//fmt.Printf("%v %v", tt.inTTL, tt.options.timeToLive)
				if !(tt.expectedErr != nil && err != nil) {
					t.Errorf("%s: expected ttl & error %v %v, got ttl & error %v %v", tt.desc, tt.inTTL, tt.expectedErr, tt.options.timeToLive, err)
				}
			}
		})
	}
}

func TestPresentOptions_ParsePadding(t *testing.T) {
	tests := []struct {
		desc string
		options PresentOptions
		inOption string
		expectedError error
	}{
		{
			desc: "empty-padding",
			options: PresentOptions{},
			inOption: "",
			expectedError: nil,
		},
		{
			desc: "invalid-padding",
			options: PresentOptions{},
			inOption: "0101f",
			expectedError: errors.New("Invalid Padding Format. -p option needs to be only 0s and 1s.\n"),
		},
		{
			desc: "valid-padding",
			options: PresentOptions{},
			inOption: "010101001",
			expectedError: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if err := tt.options.ParsePadding(tt.inOption); ((err != nil && tt.expectedError == nil) || (err == nil && tt.expectedError != nil)) || tt.options.padding != tt.inOption {
				if !(err != nil && tt.expectedError != nil) {
					t.Errorf("%s: expected padding & error %v %v got %v %v", tt.desc, tt.inOption, tt.expectedError, tt.options.padding, err)
				}
			}
		})
	}
}
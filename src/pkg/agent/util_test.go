package agent

import (
	"bytes"
	"errors"
	"testing"
	"time"
)

// Tests for utility functions.
// Found (3) bugs testing.

func TestBytesToInt(t *testing.T) {
	tests := []struct {
		desc string
		in []byte
		expected int64
	}{
		{
			desc: "test-1",
			in: []byte{0, 0, 0, 0, 0, 0, 0, 55},
			expected: 55,
		},
		{
			desc: "test-2",
			in: []byte{0,0,0,0,0,0,5,19},
			expected: 1299,
		},
		{
			desc: "large-int64",
			in: []byte{0,5,173,233,217,156,153,191},
			expected: 1598594773457343,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if conv := BytesToInt(tt.in); conv != tt.expected {
				t.Errorf("%v: expected %v got %v", tt.desc, tt.expected, conv)
			}
		})
	}
}

func TestIntToBytes(t *testing.T) {
	tests := []struct {
		desc string
		in int64
		expected []byte
	}{
		{
			desc: "test-1",
			in: 55,
			expected:  []byte{0, 0, 0, 0, 0, 0, 0, 55},
		},
		{
			desc: "test-2",
			in: 1299,
			expected: []byte{0,0,0,0,0,0,5,19},
		},
		{
			desc: "large",
			in: 1598594773457343,
			expected: []byte{0,5,173,233,217,156,153,191},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if resBytes := IntToBytes(tt.in); bytes.Compare(resBytes, tt.expected) != 0 {
				t.Errorf("%s byte slices are not equal, expected %08b got %08b", tt.desc, tt.expected, resBytes)
			}
		})
	}
}

func TestBytesToTime(t *testing.T) {
	tests := []struct {
		desc string
		in []byte
		expectedNanos int64
	}{
		{
			desc: "test-time-bytes-1",
			in: []byte{22, 6, 193, 20, 157, 16, 175, 198},
			expectedNanos: 1587168212973301702,
		},
		{
			desc: "time-test-bytes-2",
			in: []byte{22, 6, 193, 157, 64, 145, 97, 211},
			expectedNanos: 1587168799831974355,
		},
		{
			desc: "test-time-bytes-3",
			in: []byte{22, 6, 193, 157, 64, 145, 136, 111},
			expectedNanos: 1587168799831984239,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if nanos := BytesToTime(tt.in).UnixNano(); nanos != tt.expectedNanos {
				t.Errorf("%s: expected nanos %v got %v", tt.desc, tt.expectedNanos, nanos)
			}
		})
	}
	// time.Unix(0, nanos)
}

func TestTimeToBytes(t *testing.T) {
	tests := []struct {
		desc string
		in int64
		expectedTime []byte
	}{
		{
			desc: "test-time-bytes-1",
			in : 1587168212973301702,
			expectedTime: []byte{22, 6, 193, 20, 157, 16, 175, 198},
		},
		{
			desc: "time-test-bytes-2",
			in: 1587168799831974355,
			expectedTime: []byte{22, 6, 193, 157, 64, 145, 97, 211},
		},
		{
			desc: "test-time-bytes-3",
			in: 1587168799831984239,
			expectedTime: []byte{22, 6, 193, 157, 64, 145, 136, 111},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if time := TimeToBytes(time.Unix(0, tt.in)); bytes.Compare(time, tt.expectedTime) != 0 {
				t.Errorf("%s: expected nanos %08b got %08b", tt.desc, tt.expectedTime, time)
			}
		})
	}
}

func TestIsIPv4(t *testing.T) {
	tests := []struct{
		desc string
		in string
		expected bool
		expectedErr error
	}{
		{
			desc: "empty-ip-bad",
			in: "",
			expected: false,
			expectedErr: errors.New(""),
		},
		{
			desc: "invalid-ipv4",
			in: "0.f.f.3",
			expected: false,
			expectedErr: errors.New(""),
		},
		{
			desc: "valid-ipv4",
			in: "1.1.1.1",
			expected: true,
			expectedErr: nil,
		},
		{
			desc: "valid-ipv6",
			in: "2001:4860:4860::8888",
			expected: false,
			expectedErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if valid, err := IsIPv4(tt.in); valid != tt.expected || ((tt.expectedErr != nil && err == nil) || (tt.expectedErr == nil) && (err != nil)) {
				if !(tt.expectedErr != nil && err != nil) {
					t.Errorf("%s: expected %v & err %v, got %v & err %v", tt.desc,
						tt.expected, tt.expectedErr, valid, err)
				}
			}
		})
	}
}
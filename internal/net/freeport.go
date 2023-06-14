package net

import (
	"errors"
	"fmt"
	"net"
	"strconv"
)

// GetFreePorts is a simple function to get a range of free ports.
// Requests (up to) count free port[s] from within the open/closed pair of low-high numbers.
// Pass 0 in as low/high to get the bottom or top of the unprivileged range.
func GetFreePorts(count, low, high int) ([]int, error) {
	ports := make([]int, 0)
	blockSize := 128

	defaultMin := 2014
	defaultMax := 65535
	if low == 0 {
		low = defaultMin
	}
	if high == 0 {
		high = defaultMax
	}
	if low >= high {
		return ports, errors.New(fmt.Sprintf("Lowest port :%d must be smaller than highest port :%d", low, high))
	}

	currPos := low
	for currPos < high || len(ports) >= count {
		maxPos := currPos + blockSize // maxPos will be exclusive of range
		if maxPos >= high {
			maxPos = high
		}

		for currPos < maxPos {
			// try to get port here
			addr, err := net.ResolveTCPAddr("tcp", "localhost:"+strconv.Itoa(currPos))
			if err == nil {
				l, err := net.ListenTCP("tcp", addr)
				if err == nil {
					l.Close()
					ports = append(ports, currPos)
					if len(ports) >= count {
						return ports, nil
					}
				}
			}
			currPos++
		}
	}
	return ports, nil
}

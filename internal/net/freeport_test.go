package net

import (
	"fmt"
	"sort"
	"testing"
)

func TestFreePort(t *testing.T) {
	ports, err := GetFreePorts(1024, 8080, 10000)
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
	}
	fmt.Printf("Ports: %d\n", len(ports))
	sort.Ints(ports)
	fmt.Printf("Min port:%d, Max port:%d\n", ports[0], ports[len(ports)-1])
	if len(ports) != 1024 {
		t.Fail()
	}
}

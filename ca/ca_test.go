package ca

import (
	"reflect"
	"runtime"
	"testing"
)

type ipTest func(string) bool

var ipTests = []struct {
	fn      ipTest
	ip      string
	isValid bool
}{
	{isValidIP, "1", false},
	{isValidIP, "17.0.0.0", true},
	{isValidIP, "10.0.0.0", true},
	{isValidPrivateIP, "1", false},
	{isValidPrivateIP, "17.0.0.0", false},
	{isValidPrivateIP, "10.0.0.0", true},
}

// isValidIP according to IPv4
func TestIsValidIP(t *testing.T) {
	for _, ip := range ipTests {
		fnName := runtime.FuncForPC(reflect.ValueOf(ip.fn).Pointer()).Name()
		if ip.fn(ip.ip) != ip.isValid {
			t.Errorf("Error: %v(%v) != %v\n", fnName, ip.ip, ip.isValid)
		}
	}
}

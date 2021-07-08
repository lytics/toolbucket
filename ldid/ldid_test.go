package ldid

import (
	"fmt"
	"testing"
)

func TestBit(t *testing.T) {
	for i := 0; i < 100; i++ {
		fmt.Println(int64(nownano()))
	}
	for i := 0; i < 100; i++ {
		fmt.Println(int64(now10milli()))
	}
	for i := 0; i < 100; i++ {
		fmt.Println(int64(sleeptime(now10milli())))
	}
}

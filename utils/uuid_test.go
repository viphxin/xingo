package utils

import "testing"

func Test(t *testing.T) {
	a := NewUUIDGenerator("hhh")
	for {
		t.Log(a.Get())
	}
}

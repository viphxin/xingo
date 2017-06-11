package timer

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func test(a ...interface{}) {
	fmt.Println(a[0], "============", a[1])
}

func Test(t *testing.T) {
	s := NewSafeTimerScheduel()
	go func() {
		for {
			df := <-s.GetTriggerChannel()
			df.Call()
		}
	}()
	go func() {
		for {
			s.CreateTimer(int64(rand.Int31n(5000)), test, []interface{}{22, 33})
			time.Sleep(1 * time.Second)
		}
	}()
	for {
		s.CreateTimer(int64(rand.Int31n(5000)), test, []interface{}{22, 33})
		time.Sleep(2 * time.Second)
	}
}
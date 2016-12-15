package timer

import "time"

type DelayCall struct {
	f func(v ...interface{})
	args []interface{}
}

func (this *DelayCall)Call(){
	this.f(this.args...)
}

type Timer struct {
	durations time.Duration
	delayCall *DelayCall
}

func NewTimer(durations time.Duration, f func(v ...interface{}), args []interface{}) *Timer{
	return &Timer{
		durations : durations,
		delayCall: &DelayCall{
			f: f,
			args: args,
		},
	}
}

func (this *Timer)Run(){
	go func() {
		time.Sleep(this.durations)
		this.delayCall.Call()
	}()
}

func (this *Timer)GetDurations() time.Duration{
	return this.durations
}

func (this *Timer)GetFunc() *DelayCall{
	return this.delayCall
}

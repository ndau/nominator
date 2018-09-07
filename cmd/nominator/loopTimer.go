package main

import "time"

// LoopTimer is a timer that keeps re-starting itself after the LoopTime.
// Each time it restarts, it calls the LoopFunc in its own goroutine.
// The granularity (precision) of the timer is controlled by TickTime.
type LoopTimer struct {
	TickTime time.Duration
	LoopTime time.Duration
	LoopFunc func()
	t        *time.Timer
	nextLoop time.Time
}

// NewLoopTimer creates a loopTimer
func NewLoopTimer(tickTime, loopTime time.Duration, loopFunc func()) *LoopTimer {
	t := &LoopTimer{
		TickTime: tickTime,
		LoopTime: loopTime,
		LoopFunc: loopFunc,
		nextLoop: time.Now().Add(loopTime),
	}
	t.t = time.AfterFunc(tickTime, t.tickFunc)
	return t
}

// TickFunc is called after every tick; if the current time is after the nextLoop
// time, we call the LoopFunc() and add the LoopTime to the nextLoop time. Note this is
// NOT added to the current time -- this ensures that on average we'll be no more than
// one TickTime away from the loopTime (as long as no one calls AtMost).
func (l *LoopTimer) tickFunc() {
	if time.Now().After(l.nextLoop) {
		go l.LoopFunc()
		l.nextLoop = l.nextLoop.Add(l.LoopTime)
	}
	l.t = time.AfterFunc(l.TickTime, l.tickFunc)
}

package main

import "time"

type timeout struct {
	hard              bool
	firstTimeInfinite bool
	infinite          bool
	firstTime         time.Duration
	time              time.Duration

	c     *<-chan time.Time
	timer *time.Timer
}

func (t *timeout) Start() {
	if !t.infinite && !t.firstTimeInfinite {
		t.timer = time.NewTimer(t.firstTime)
		t.c = &t.timer.C
	} else {
		ch := make(<-chan time.Time)
		t.c = &ch
	}
}

func (t *timeout) Reset() {
	if !t.hard && !t.infinite {
		if t.firstTimeInfinite {
			t.timer = time.NewTimer(t.time)
			t.c = &t.timer.C
		} else {
			t.timer.Reset(t.time)
		}
	} else {
		time.Sleep(1 * time.Millisecond)
	}
}

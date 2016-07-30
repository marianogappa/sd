package main

import (
	"testing"
	"time"
)

func TestTimeoutEventuallyHappens(t *testing.T) {
	timeout := timeout{
		hard:              false,
		firstTimeInfinite: false,
		infinite:          false,
		firstTime:         1 * time.Millisecond,
		time:              1 * time.Millisecond,
	}

	timeout.Start()
	<-*timeout.c
}

func TestFirstTimeInfinite(t *testing.T) {
	timeout := timeout{
		firstTimeInfinite: true,
		firstTime:         1 * time.Millisecond,
		time:              1 * time.Millisecond,
	}
	control := time.After(100 * time.Millisecond)

	timeout.Start()
	for {
		select {
		case <-control:
			return
		case <-*timeout.c:
			t.Error("first time wasn't infinite")
			return
		}
	}
}

func TestInfinite(t *testing.T) {
	timeout := timeout{
		infinite:  true,
		firstTime: 1 * time.Millisecond,
		time:      1 * time.Millisecond,
	}
	control := time.After(100 * time.Millisecond)

	timeout.Start()
	for {
		select {
		case <-control:
			return
		case <-*timeout.c:
			t.Error("first time wasn't infinite")
			return
		}
	}
}

func TestFirstTime(t *testing.T) {
	timeout := timeout{
		firstTimeInfinite: false,
		firstTime:         100 * time.Millisecond,
		time:              1 * time.Millisecond,
	}
	controlTime := 200 * time.Millisecond
	control := time.NewTimer(controlTime)

	timeout.Start()
	for {
		select {
		case <-control.C:
			t.Error("control happened before timeout")
		case <-*timeout.c:
			if controlTime == 100*time.Millisecond {
				return
			}
			timeout.Reset()
			controlTime = 100 * time.Millisecond
			control.Reset(controlTime)
		}
	}

}

// has no semantic meaning other than "first time; no reset" on its own
func TestHardTimeout(t *testing.T) {
	timeout := timeout{
		hard:      true,
		firstTime: 1 * time.Millisecond,
		time:      1 * time.Millisecond,
	}
	controlTime := 100 * time.Millisecond
	control := time.NewTimer(controlTime)

	timeout.Start()
	for {
		select {
		case <-control.C:
			if controlTime == 100*time.Millisecond {
				t.Error("control happened before timeout")
				return
			}
			return
		case <-*timeout.c:
			if controlTime != 100*time.Millisecond {
				t.Error("control happened before timeout")
				return
			}
			timeout.Reset()
			controlTime = 150 * time.Millisecond
			control.Reset(controlTime)
		}
	}
}

func TestNonFirstTime(t *testing.T) {
	timeout := timeout{
		firstTimeInfinite: false,
		firstTime:         100 * time.Millisecond,
		time:              10 * time.Millisecond,
	}
	controlTime := 200 * time.Millisecond
	control := time.NewTimer(controlTime)

	timeout.Start()
	for {
		select {
		case <-control.C:
			t.Error("control happened before timeout")
			return

		case <-*timeout.c:
			if controlTime == 50*time.Millisecond {
				return
			}
			controlTime = 50 * time.Millisecond
			control.Reset(controlTime)

			timeout.Reset()
		}
	}
}

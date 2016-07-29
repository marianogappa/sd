package main

import "testing"

func TestDefaultArgs(t *testing.T) {
	follow, infinite, patience, timeoutF, hardTimeout := mustResolveOptions([]string{})
	if follow != false || infinite != false || patience != -1 || timeoutF != 10 || hardTimeout != 0 {
		t.Errorf("default options are incorrect: %v, %v, %v, %v, %v", follow, infinite, patience, timeoutF, hardTimeout)
	}
}

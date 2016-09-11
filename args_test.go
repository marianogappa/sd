package main

import (
	"reflect"
	"testing"
	"time"
)

func TestResolveOptions(t *testing.T) {
	tests := []struct {
		args     []string
		expected options
		fails    bool
	}{
		{
			args: []string{},
			expected: options{
				follow:      false,
				infinite:    false,
				intersection: false,
				patience:    -1,
				timeoutF:    10,
				hardTimeout: 0,
			},
		},
		{
			args: []string{"-f"},
			expected: options{
				follow:      true,
				infinite:    false,
				intersection: false,
				patience:    -1,
				timeoutF:    10,
				hardTimeout: 0,
			},
		},
		{
			args: []string{"-i"},
			expected: options{
				follow:      false,
				infinite:    true,
				intersection: false,
				patience:    -1,
				timeoutF:    10,
				hardTimeout: 0,
			},
		},
		{
			args: []string{"-p", "0"},
			expected: options{
				follow:      false,
				infinite:    false,
				intersection: false,
				patience:    0,
				timeoutF:    10,
				hardTimeout: 0,
			},
		},
		{
			args: []string{"-t", "5"},
			expected: options{
				follow:      false,
				infinite:    false,
				intersection: false,
				patience:    -1,
				timeoutF:    5,
				hardTimeout: 0,
			},
		},
		{
			args: []string{"-h", "120"},
			expected: options{
				follow:      false,
				infinite:    false,
				intersection: false,
				patience:    -1,
				timeoutF:    10,
				hardTimeout: 120,
			},
		},
		{
			args: []string{"-i", "-f"},
			expected: options{
				follow:      true,
				infinite:    true,
				intersection: false,
				patience:    -1,
				timeoutF:    10,
				hardTimeout: 0,
			},
		},
		{
			args:  []string{"-p"},
			fails: true,
		},
		{
			args:  []string{"-t"},
			fails: true,
		},
		{
			args:  []string{"-h"},
			fails: true,
		},
		{
			args:  []string{"-if"},
			fails: true,
		},
		{
			args: []string{"-f", "-t", "1", "-i", "-p", "2", "-h", "3"},
			expected: options{
				follow:      true,
				infinite:    true,
				intersection: false,
				patience:    2,
				timeoutF:    1,
				hardTimeout: 3,
			},
		},
		{
			args: []string{"--follow", "--timeout", "1", "--infinite", "--patience", "2", "--hard-timeout", "3"},
			expected: options{
				follow:      true,
				infinite:    true,
				intersection: false,
				patience:    2,
				timeoutF:    1,
				hardTimeout: 3,
			},
		},
		{
			args: []string{"--intersection"},
			expected: options{
				follow:      false,
				infinite:    false,
				intersection: true,
				patience:    -1,
				timeoutF:    10,
				hardTimeout: 0,
			},
		},
	}

	for _, ts := range tests {
		result, err := resolveOptions(ts.args)

		if ts.fails && err == nil {
			t.Errorf("should have failed with %v", result)
		}

		if !ts.fails && err != nil {
			t.Errorf("should not have failed resolving options with %v", result)
		}

		if !ts.fails && !reflect.DeepEqual(*result, ts.expected) {
			t.Errorf("default options are incorrect: %v was not equal to %v", result, ts.expected)
		}
	}
}

func TestResolveTimeouts(t *testing.T) {
	tests := []struct {
		options      *options
		stdinTimeout timeout
		cmdTimeout   timeout
	}{
		{
			options: &options{
				follow:      false,
				infinite:    false,
				intersection: false,
				patience:    -1,
				timeoutF:    10,
				hardTimeout: 0,
			},
			stdinTimeout: timeout{
				hard:              false,
				firstTimeInfinite: false,
				infinite:          false,
				firstTime:         10 * time.Second,
				time:              10 * time.Second,
			},
			cmdTimeout: timeout{
				hard:              false,
				firstTimeInfinite: false,
				infinite:          false,
				firstTime:         10 * time.Second,
				time:              10 * time.Second,
			},
		},
		{
			options: &options{
				follow:      true,
				infinite:    false,
				intersection: false,
				patience:    -1,
				timeoutF:    10,
				hardTimeout: 0,
			},
			stdinTimeout: timeout{
				hard:              false,
				firstTimeInfinite: false,
				infinite:          true,
				firstTime:         10 * time.Second,
				time:              10 * time.Second,
			},
			cmdTimeout: timeout{
				hard:              false,
				firstTimeInfinite: false,
				infinite:          false,
				firstTime:         10 * time.Second,
				time:              10 * time.Second,
			},
		},
		{
			options: &options{
				follow:      false,
				infinite:    true,
				intersection: false,
				patience:    -1,
				timeoutF:    10,
				hardTimeout: 0,
			},
			stdinTimeout: timeout{
				hard:              false,
				firstTimeInfinite: false,
				infinite:          false,
				firstTime:         10 * time.Second,
				time:              10 * time.Second,
			},
			cmdTimeout: timeout{
				hard:              false,
				firstTimeInfinite: false,
				infinite:          true,
				firstTime:         10 * time.Second,
				time:              10 * time.Second,
			},
		},
		{
			options: &options{
				follow:      false,
				infinite:    false,
				intersection: false,
				patience:    0,
				timeoutF:    10,
				hardTimeout: 0,
			},
			stdinTimeout: timeout{
				hard:              false,
				firstTimeInfinite: true,
				infinite:          false,
				firstTime:         0 * time.Second,
				time:              10 * time.Second,
			},
			cmdTimeout: timeout{
				hard:              false,
				firstTimeInfinite: true,
				infinite:          false,
				firstTime:         0 * time.Second,
				time:              10 * time.Second,
			},
		},
		{
			options: &options{
				follow:      true,
				infinite:    false,
				intersection: false,
				patience:    20,
				timeoutF:    10,
				hardTimeout: 0,
			},
			stdinTimeout: timeout{
				hard:              false,
				firstTimeInfinite: false,
				infinite:          true,
				firstTime:         20 * time.Second,
				time:              10 * time.Second,
			},
			cmdTimeout: timeout{
				hard:              false,
				firstTimeInfinite: false,
				infinite:          false,
				firstTime:         20 * time.Second,
				time:              10 * time.Second,
			},
		},
		{
			options: &options{
				follow:      false,
				infinite:    false,
				intersection: false,
				patience:    -1,
				timeoutF:    10,
				hardTimeout: 120,
			},
			stdinTimeout: timeout{
				hard:              true,
				firstTimeInfinite: false,
				infinite:          false,
				firstTime:         120 * time.Second,
				time:              10 * time.Second,
			},
			cmdTimeout: timeout{
				hard:              true,
				firstTimeInfinite: false,
				infinite:          false,
				firstTime:         120 * time.Second,
				time:              10 * time.Second,
			},
		},
		{
			options: &options{
				follow:      true,
				infinite:    false,
				intersection: false,
				patience:    -1,
				timeoutF:    30,
				hardTimeout: 0,
			},
			stdinTimeout: timeout{
				hard:              false,
				firstTimeInfinite: false,
				infinite:          true,
				firstTime:         30 * time.Second,
				time:              30 * time.Second,
			},
			cmdTimeout: timeout{
				hard:              false,
				firstTimeInfinite: false,
				infinite:          false,
				firstTime:         30 * time.Second,
				time:              30 * time.Second,
			},
		},
	}

	for _, ts := range tests {
		stdinTimeout, cmdTimeout := resolveTimeouts(ts.options)

		if !reflect.DeepEqual(stdinTimeout, ts.stdinTimeout) {
			t.Errorf("stdinTimeout resolved incorrectly: %v was not equal to %v", stdinTimeout, ts.stdinTimeout)
		}

		if !reflect.DeepEqual(cmdTimeout, ts.cmdTimeout) {
			t.Errorf("cmdTimeout resolved incorrectly: %v was not equal to %v", cmdTimeout, ts.cmdTimeout)
		}
	}
}

package main

import (
	"reflect"
	"testing"
)

var tests = []struct {
	args     []string
	expected options
	fails    bool
}{
	{
		args: []string{},
		expected: options{
			follow:      false,
			infinite:    false,
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
			patience:    2,
			timeoutF:    1,
			hardTimeout: 3,
		},
	},
}

func TestResolveOptions(t *testing.T) {
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

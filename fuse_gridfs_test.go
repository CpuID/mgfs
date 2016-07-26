package main

import (
	"testing"
)

func TestBuildGridFsPath(t *testing.T) {
	// key = result
	// value[0] = parent_dir
	// value[1] = filename
	tests := map[string][]string{
		"asdf/zzzz": []string{
			"/asdf",
			"zzzz",
		},
		"ffff": []string{
			"/",
			"ffff",
		},
		"aaa/bbb/ccc/ddd": []string{
			"/aaa/bbb/ccc",
			"ddd",
		},
	}
	for k, v := range tests {
		result, err := buildGridFsPath(v[0], v[1])
		if err != nil {
			t.Error(v[0], ":", v[1], "- Expected nil error, got", err)
		}
		if result != k {
			t.Error(v[0], ":", v[1], "- Expected", k, "got", result)
		}
	}
}

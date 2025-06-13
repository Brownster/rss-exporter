package main

import (
	"reflect"
	"strconv"
	"testing"
)

func TestParseStringSlice_TrailingComma(t *testing.T) {
	tests := []struct {
		input    string
		expected []int
	}{
		{"200,", []int{200}},
		{"200,301,", []int{200, 301}},
		{",", []int{}},
		{"", []int{}},
	}
	for _, tt := range tests {
		got, err := parseStringSlice[int](tt.input, strconv.Atoi)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", tt.input, err)
		}
		if !reflect.DeepEqual(got, tt.expected) {
			t.Errorf("for %q expected %v got %v", tt.input, tt.expected, got)
		}
	}
}

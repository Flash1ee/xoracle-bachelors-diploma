package command

import (
	"testing"
)

func TestRoundUpToNearestMultiple(t *testing.T) {
	testCases := []struct {
		n        int
		k        int
		expected int
	}{
		{100, 3, 102},
		{17, 5, 20},
		{50, 10, 50},
		{123, 7, 126},
		{8, 2, 8},
		{0, 10, 0},
	}

	for _, testCase := range testCases {
		result := RoundUpToNearestMultiple(testCase.n, testCase.k)
		if result != testCase.expected {
			t.Errorf("n: %d, k: %d - Expected: %d, Got: %d", testCase.n, testCase.k, testCase.expected, result)
		}
	}
}

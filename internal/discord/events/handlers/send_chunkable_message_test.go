package discord_event_handlers

import (
	"testing"
)

func TestBinSearch(t *testing.T) {
	cases := []struct {
		name     string
		sizes    []int
		expected int
	}{
		{
			name:     "even elements count",
			sizes:    []int{100, 300, 500, 1000, 1500, 2000, 2500, 3000, 3500, 4000},
			expected: 1500,
		},
		{
			name:     "odd elements count",
			sizes:    []int{100, 300, 500, 1000, 1500, 2000, 2500, 3000, 3500},
			expected: 1500,
		},
		{
			name:     "first element",
			sizes:    []int{1500, 2000},
			expected: 1500,
		},
		{
			name:     "last element",
			sizes:    []int{100, 300, 1500},
			expected: 1500,
		},
		{
			name:     "one element",
			sizes:    []int{1500},
			expected: 1500,
		},
		{
			name:     "no element",
			sizes:    []int{2000, 3000, 3500, 4000},
			expected: -1,
		},
		{
			name:     "empty slice",
			sizes:    []int{},
			expected: -1,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			index := binSearch(len(c.sizes), func(u1 int) bool {
				return c.sizes[u1] < 2000
			})
			if index >= 0 {
				if c.sizes[index] != c.expected {
					t.Fatalf("expected %d, got %d", c.expected, c.sizes[index])
				}
			} else {
				if index != c.expected {
					t.Fatalf("expected %d, got %d", c.expected, index)
				}
			}
		})
	}
}

package soyusage

import (
	"testing"

	"github.com/theothertomelliott/must"
)

func TestScopeCycles(t *testing.T) {
	var tests = []struct {
		name     string
		scope    *scope
		expected int
	}{
		{
			name: "empty stack",
			scope: &scope{
				templateName: "a",
			},
			expected: 0,
		},
		{
			name: "one call",
			scope: &scope{
				templateName: "a",
				callStack: []*scope{
					&scope{
						templateName: "b",
					},
				},
			},
			expected: 0,
		},
		{
			name: "simple cycle",
			scope: &scope{
				templateName: "a",
				callStack: []*scope{
					&scope{
						templateName: "a",
					},
				},
			},
			expected: 1,
		},
		{
			name: "multiple cycles",
			scope: &scope{
				templateName: "a",
				callStack: []*scope{
					&scope{
						templateName: "a",
					},
					&scope{
						templateName: "a",
					},
					&scope{
						templateName: "a",
					},
					&scope{
						templateName: "a",
					},
				},
			},
			expected: 4,
		},
		{
			name: "complex cycle",
			scope: &scope{
				templateName: "a",
				callStack: []*scope{
					&scope{
						templateName: "a",
					},
					&scope{
						templateName: "b",
					},
					&scope{
						templateName: "c",
					},
					&scope{
						templateName: "a",
					},
					&scope{
						templateName: "b",
					},
					&scope{
						templateName: "c",
					},
				},
			},
			expected: 2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.scope.callCycles()
			must.BeEqual(t, test.expected, got)
		})
	}
}

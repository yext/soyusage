package soyusage_test

import "testing"

func TestAnalyzeForLoops(t *testing.T) {
	var tests = []analyzeTest{
		{
			name: "foreach loops create variables",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param list
				*/
				{template .main}
					{foreach $item in $list}
						{$item.field}
					{/foreach}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"list": map[string]interface{}{
					"field": map[string]interface{}{
						"*": struct{}{},
					},
				},
			},
		},
		{
			name: "handles ranged indexing",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param a
				*/
				{template .main}
					{foreach $item in range(3)}
						{$a[$item].b}
					{/foreach}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"*": struct{}{},
					},
				},
			},
		},
	}
	testAnalyze(t, tests)
}

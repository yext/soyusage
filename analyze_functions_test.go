package soyusage_test

import "testing"

// TestAnalyzeFunctions verifies behavior when analyzing function calls
func TestAnalyzeFunctions(t *testing.T) {
	var tests = []analyzeTest{
		{
			name: "unknown function gives unknown usage",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param a
				*/
				{template .main}
					{myFunc($a.b)}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"?": struct{}{},
					},
				},
			},
		},
		{
			name: "length does not affect usage",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param a
				*/
				{template .main}
					{if length($a) > 0}
						{$a[0].b}
					{/if}
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
		{
			name: "augmentMap adds to both maps",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param a
				* @param b
				*/
				{template .main}
					{let $c: augmentMap($a,$b)/}
					{$c.d}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"d": map[string]interface{}{
						"*": struct{}{},
					},
				},
				"b": map[string]interface{}{
					"d": map[string]interface{}{
						"*": struct{}{},
					},
				},
			},
		},
		{
			name: "augmentMap and quoteKeysIfJs do not affect structure",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param a
				* @param b
				*/
				{template .main}
					{let $x: augmentMap($a,$b)/}
					{let $y: quoteKeysIfJs($a)/}
					{$x.c}
					{$y.d}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"c": map[string]interface{}{
						"*": struct{}{},
					},
					"d": map[string]interface{}{
						"*": struct{}{},
					},
				},
				"b": map[string]interface{}{
					"c": map[string]interface{}{
						"*": struct{}{},
					},
				},
			},
		},
	}
	testAnalyze(t, tests)
}

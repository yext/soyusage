package soyusage_test

import "testing"

// TestAnalyzeConstantMapAccess executes a set of tests to verify the Analyze function's handling
// of using constant values to access map entries.
//
// This allows for situations where there are a fixed set of map fields being accessed, but there is logic
// in the template that selects between them based on other inputs.
func TestAnalyzeConstantMapAccess(t *testing.T) {
	var tests = []analyzeTest{
		{
			name: "handles mapping of string values",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param profile
				*/
				{template .main}
					{let $textField}
						c_lifeAbout
					{/let}
					{let $textField2: 'c_other'/}
					{$profile[$textField]}
					{$profile[$textField2]}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"profile": map[string]interface{}{
					"c_other": map[string]interface{}{
						"*": struct{}{},
					},
					"c_lifeAbout": map[string]interface{}{
						"*": struct{}{},
					},
				},
			},
		},
		{
			name: "handles mapping from a switch statement",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param profile
				* @param category
				* @param about
				*/
				{template .main}
					{let $textField}
						{switch $category}
							{case 'Auto'}
								c_autoAbout
							{case 'Home'}
								c_homeAbout
							{case $about}
								c_lifeAbout
						{/switch}
					{/let}
					{if $profile[$textField]}
						{$profile[$textField]}
					{/if}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"category": map[string]interface{}{
					"*": struct{}{},
				},
				"about": map[string]interface{}{
					"*": struct{}{},
				},
				"profile": map[string]interface{}{
					"c_autoAbout": map[string]interface{}{
						"*": struct{}{},
					},
					"c_homeAbout": map[string]interface{}{
						"*": struct{}{},
					},
					"c_lifeAbout": map[string]interface{}{
						"*": struct{}{},
					},
				},
			},
		},
		{
			name: "handles mapping from a list literal",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param profile
				*/
				{template .main}
					{let $list: [
						'c_education',
						'c_awards'
					]/}
					{foreach $item in $list}
						{$profile[$item]}
					{/foreach}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"profile": map[string]interface{}{
					"c_education": map[string]interface{}{
						"*": struct{}{},
					},
					"c_awards": map[string]interface{}{
						"*": struct{}{},
					},
				},
			},
		},
		{
			name: "handles map literal inside list",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param profile
				*/
				{template .main}
					{let $list: [
						['field': 'c_education'],
						['field': 'c_awards']
					]/}
					{foreach $item in $list}
						{$profile[$item.field]}
					{/foreach}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"profile": map[string]interface{}{
					"c_education": map[string]interface{}{
						"*": struct{}{},
					},
					"c_awards": map[string]interface{}{
						"*": struct{}{},
					},
				},
			},
		},
		{
			name: "handles ranged map literal, no start",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param profile
				*/
				{template .main}
					{foreach $i in range(2)}
						{$profile['field' + $i]}
					{/foreach}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"profile": map[string]interface{}{
					"field0": map[string]interface{}{
						"*": struct{}{},
					},
					"field1": map[string]interface{}{
						"*": struct{}{},
					},
				},
			},
		},
		{
			name: "handles ranged map literal",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param profile
				*/
				{template .main}
					{foreach $i in range(1,3)}
						{$profile['field' + $i]}
					{/foreach}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"profile": map[string]interface{}{
					"field1": map[string]interface{}{
						"*": struct{}{},
					},
					"field2": map[string]interface{}{
						"*": struct{}{},
					},
				},
			},
		},
		{
			name: "handles ranged map literal with increment",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param profile
				*/
				{template .main}
					{foreach $i in range(2,6,2)}
						{$profile['field' + $i]}
					{/foreach}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"profile": map[string]interface{}{
					"field2": map[string]interface{}{
						"*": struct{}{},
					},
					"field4": map[string]interface{}{
						"*": struct{}{},
					},
				},
			},
		},
		{
			name: "handles keyed map literal",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param profile
				*/
				{template .main}
					{let $m: [
						'first': 'c_education',
						'second': 'c_awards'
					]/}
					{foreach $i in keys($m)}
						{$profile[$i]}
					{/foreach}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"profile": map[string]interface{}{
					"c_education": map[string]interface{}{
						"*": struct{}{},
					},
					"c_awards": map[string]interface{}{
						"*": struct{}{},
					},
				},
			},
		},
		{
			name: "handles mapping from an if statement",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param profile
				* @param category
				*/
				{template .main}
					{let $textField}
						{if $category == 'Auto'}
							c_autoAbout
						{else}
							c_lifeAbout
						{/if}
					{/let}
					{if $profile[$textField]}
						{$profile[$textField]}
					{/if}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"category": map[string]interface{}{
					"*": struct{}{},
				},
				"profile": map[string]interface{}{
					"c_autoAbout": map[string]interface{}{
						"*": struct{}{},
					},
					"c_lifeAbout": map[string]interface{}{
						"*": struct{}{},
					},
				},
			},
		},
		{
			name: "handles mapping from nested statements",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param profile
				* @param category
				* @param about
				*/
				{template .main}
					{let $textField}
						{switch ($category ?: '')}
							{case 'Auto'}
								c_autoAbout
							{case 'Home'}
								c_homeAbout
							{default}
								{if $about == 'Life'}
									c_lifeAbout
								{else}
									c_about
								{/if}
						{/switch}
					{/let}
					{let $value: $profile[$textField] /}
					{$value}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"category": map[string]interface{}{
					"*": struct{}{},
				},
				"about": map[string]interface{}{
					"*": struct{}{},
				},
				"profile": map[string]interface{}{
					"c_autoAbout": map[string]interface{}{
						"*": struct{}{},
					},
					"c_homeAbout": map[string]interface{}{
						"*": struct{}{},
					},
					"c_lifeAbout": map[string]interface{}{
						"*": struct{}{},
					},
					"c_about": map[string]interface{}{
						"*": struct{}{},
					},
				},
			},
		},
	}
	testAnalyze(t, tests)
}

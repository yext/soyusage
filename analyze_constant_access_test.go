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
			name: "handles complex mapping",
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
	}
	testAnalyze(t, tests)
}

package soyusage_test

import (
	"testing"
)

func TestAnalyzeCall(t *testing.T) {
	var tests = []analyzeTest{
		{
			name: "call params are recorded",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param data
				* @param param
				*/
				{template .main}
					{call .callee data="$data"}
						{param passByParam: $param/}
					{/call}
				{/template}

				/**
				* @param passByParam
				* @param? dataChild
				*/
				{template .callee}
					{$passByParam.paramChild}
					{$dataChild}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"data": map[string]interface{}{
					"dataChild": map[string]interface{}{
						"*": struct{}{},
					},
				},
				"param": map[string]interface{}{
					"paramChild": map[string]interface{}{
						"*": struct{}{},
					},
				},
			},
		},
		{
			name: "call params with all",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param data
				*/
				{template .main}
					{call .callee data="all"}
					{/call}
				{/template}

				/**
				* @param data
				*/
				{template .callee}
					{$data.dataChild}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"data": map[string]interface{}{
					"dataChild": map[string]interface{}{
						"*": struct{}{},
					},
				},
			},
		},
		{
			name: "handles call cycle",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param data
				*/
				{template .main}
					{call .callee data="$data"}
					{/call}
				{/template}

				/**
				* @param child
				*/
				{template .callee}
					{$child.value}
					{call .callee data="$child.data"}
					{/call}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"data": map[string]interface{}{
					"child": map[string]interface{}{
						"data": map[string]interface{}{
							"child": map[string]interface{}{
								"data": struct{}{},
								"value": map[string]interface{}{
									"*": map[string]interface{}{},
								},
							},
						},
						"value": map[string]interface{}{
							"*": struct{}{},
						},
					},
				},
			},
		},
	}
	testAnalyze(t, tests)
}

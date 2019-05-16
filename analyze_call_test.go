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
					"dataChild": "*",
				},
				"param": map[string]interface{}{
					"paramChild": "*",
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
					"dataChild": "*",
				},
			},
		},
		{
			name: "sub calls preserve all",
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
					{$data.child}
					{call .subcallee data="all"}
					{/call}
				{/template}

				/**
				* @param data
				*/
				{template .subcallee}
					{$data.child2}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"data": map[string]interface{}{
					"child":  "*",
					"child2": "*",
				},
			},
		},
		{
			name: "handles simple cycle",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param data
				* @param x
				*/
				{template .callee}
					{$x}
					{call .callee data="$data.child"}
						{param x: $data.value /}
					{/call}
				{/template}
			`,
			},
			templateName: "test.callee",
			expected: map[string]interface{}{
				"data": map[string]interface{}{
					"child": map[string]interface{}{
						"data": map[string]interface{}{
							"child": map[string]interface{}{
								"data": "*",
							},
							"value": "*",
						},
					},
					"value": "*",
				},
				"x": "*",
			},
		},
		{
			name: "handles nested call cycle",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param data
				* @param altValue
				*/
				{template .main}
					{call .callee data="$data"}
						{param x}
							{$altValue}
						{/param}
					{/call}
				{/template}

				/**
				* @param child
				* @param x
				* @param? y
				*/
				{template .callee}
					{$x}
					{$y}
					{call .callee data="$child.data"}
						{param x: $child.value /}
						{param y}constant{/param}
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
								"data": map[string]interface{}{
									"child": "*",
								},
								"value": "*",
							},
						},
						"value": "*",
					},
					"y": "*",
				},
				"altValue": "*",
			},
		},
		{
			name: "handles nested cycle with all data",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param data
				* @param altValue
				*/
				{template .main}
					{call .callee data="$data"}
						{param x}
							{$altValue}
						{/param}
					{/call}
				{/template}

				/**
				* @param child
				* @param x
				*/
				{template .callee}
					{$x}
					{call .callee data="all"}
						{param x: $child.value /}
					{/call}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"data": map[string]interface{}{
					"child": "*",
				},
				"altValue": "*",
			},
		},
	}
	testAnalyze(t, tests)
}

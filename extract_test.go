package soyusage_test

import (
	"testing"

	"github.com/theothertomelliott/must"
	"github.com/yext/soyusage"

	"github.com/yext/soy"
	"github.com/yext/soy/data"
)

func TestExtract(t *testing.T) {
	var tests = []extractTest{
		{
			name: "missing params are ignored",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param a
				*/
				{template .main}
					{$a.b | json}
				{/template}
			`,
			},
			templateName: "test.main",
			in: data.New(map[string]interface{}{
				"a": map[string]interface{}{},
			}),
			expected: data.New(map[string]interface{}{
				"a": map[string]interface{}{},
			}),
		},
		{
			name: "handles if",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param a
				* @param b
				*/
				{template .main}
					{if $a}
						{$b}
					{/if}
				{/template}
			`,
			},
			templateName: "test.main",
			in: data.New(map[string]interface{}{
				"a": map[string]interface{}{
					"item1": "value1",
					"item2": "value2",
				},
				"b": "bvalue",
			}),
			expected: data.New(map[string]interface{}{
				"a": "",
				"b": "bvalue",
			}),
		},
		{
			name: "iteration is handled",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param list
				*/
				{template .main}
					{foreach $item in $list}
						{$item.value}
					{/foreach}
				{/template}
			`,
			},
			templateName: "test.main",
			in: data.New(map[string]interface{}{
				"list": []interface{}{
					map[string]interface{}{
						"value":  1,
						"unused": "ignore1",
					},
					map[string]interface{}{
						"value":  2,
						"unused": "ignore2",
					},
				},
			}),
			expected: data.New(map[string]interface{}{
				"list": []interface{}{
					map[string]interface{}{
						"value": 1,
					},
					map[string]interface{}{
						"value": 2,
					},
				},
			}),
		},
		{
			name: "unknown map index",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param map
				* @param index
				*/
				{template .main}
					{$map[$index].value}
				{/template}
			`,
			},
			templateName: "test.main",
			in: data.New(map[string]interface{}{
				"map": map[string]interface{}{
					"a": map[string]interface{}{
						"value":  1,
						"unused": "ignore1",
					},
					"b": map[string]interface{}{
						"value":  2,
						"unused": "ignore2",
					},
				},
			}),
			expected: data.New(map[string]interface{}{
				"map": map[string]interface{}{
					"a": map[string]interface{}{
						"value": 1,
					},
					"b": map[string]interface{}{
						"value": 2,
					},
				},
			}),
		},
		{
			name: "removes unused parameters",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param a
				*/
				{template .main}
					{$a.b | json}
				{/template}
			`,
			},
			templateName: "test.main",
			in: data.New(map[string]interface{}{
				"a": map[string]interface{}{
					"b": "value",
					"c": "not used",
				},
				"d": "also not used",
			}),
			expected: data.New(map[string]interface{}{
				"a": map[string]interface{}{
					"b": "value",
				},
			}),
		},
		{
			name: "print outputs complete structure",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param a
				*/
				{template .main}
					{$a}
				{/template}
			`,
			},
			templateName: "test.main",
			in: data.New(map[string]interface{}{
				"a": map[string]interface{}{
					"b": "value",
					"c": "another value",
				},
				"d": "also not used",
			}),
			expected: data.New(map[string]interface{}{
				"a": map[string]interface{}{
					"b": "value",
					"c": "another value",
				},
			}),
		},
		{
			name: "applies recursion up to limit",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param data
				* @param x
				*/
				{template .callee}
					{$x}
					{call .callee data="$data"}
						{param x: $data.value /}
					{/call}
				{/template}
			`,
			},
			templateName: "test.callee",
			in: data.New(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{
						"data": map[string]interface{}{
							"data": map[string]interface{}{
								"value":  "level 4",
								"unused": "4th unused",
							},
							"value":  "level 3",
							"unused": "3rd unused",
						},
						"value":  "level 2",
						"unused": "2nd unused",
					},
				},
				"x": "another value",
			}),
			expected: data.New(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{
						"data": map[string]interface{}{
							"data": map[string]interface{}{
								"value":  "level 4",
								"unused": "4th unused",
							},
							"value":  "level 3",
							"unused": "3rd unused",
						},
						"value": "level 2",
					},
				},
				"x": "another value",
			}),
		},
		{
			name: "recurses to full depth",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param data
				* @param x
				*/
				{template .callee}
					{$x}
					{call .callee data="$data"}
						{param x: $data.value /}
					{/call}
				{/template}
			`,
			},
			templateName: "test.callee",
			in: data.New(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{
						"data": map[string]interface{}{
							"data": map[string]interface{}{
								"value":  "level 4",
								"unused": "4th unused",
							},
							"value":  "level 3",
							"unused": "3rd unused",
						},
						"value":  "level 2",
						"unused": "2nd unused",
					},
				},
				"x": "another value",
			}),
			expected: data.New(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{
						"data": map[string]interface{}{
							"data": map[string]interface{}{
								"value": "level 4",
							},
							"value": "level 3",
						},
						"value": "level 2",
					},
				},
				"x": "another value",
			}),
			recursionDepth: 5,
		},
		{
			name: "handles combined constant and variable values",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param profile
				* @param locale
				* @param alternative
				*/
				{template .main}
					{let $textField}
						{if $locale == 'en'}
							c_lifeAbout
						{else}
							{$alternative}
						{/if}
					{/let}
					{$profile[$textField]}
				{/template}
			`,
			},
			templateName: "test.main",
			in: data.New(map[string]interface{}{
				"profile": map[string]interface{}{
					"c_lifeAbout": "life about",
					"second":      "value 2",
					"third":       "value 3",
				},
				"locale":      "loc",
				"alternative": "alt",
			}),
			expected: data.New(map[string]interface{}{
				"profile": map[string]interface{}{
					"c_lifeAbout": "life about",
					"second":      "value 2",
					"third":       "value 3",
				},
				"locale":      "loc",
				"alternative": "alt",
			}),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bundle := soy.NewBundle()
			for name, content := range test.templates {
				bundle = bundle.AddTemplateString(name, content)
			}
			registry, err := bundle.Compile()
			if err != nil {
				t.Fatal(err)
			}
			if test.recursionDepth == 0 {
				test.recursionDepth = 2
			}
			params, err := soyusage.AnalyzeTemplate(test.templateName, registry, soyusage.Recursion(test.recursionDepth))
			if err != nil {
				t.Fatal(err)
			}
			got := soyusage.Extract(test.in.(data.Map), params)
			must.BeEqual(t, test.expected.(data.Map), got)
			if t.Failed() {
				t.Log(jsonSprint(mapUsage(params)))
			}
		})
	}
}

type extractTest struct {
	name           string
	templates      map[string]string
	templateName   string
	in             data.Value
	expected       data.Value
	recursionDepth int
}

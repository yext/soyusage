package soyusage_test

import (
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	"github.com/theothertomelliott/must"
	"github.com/yext/soy"
	"github.com/yext/soy/template"
	"github.com/yext/soyusage"
)

func TestAnalyzeParamHierarchy(t *testing.T) {
	var tests = []analyzeTest{
		{
			name: "printed parameters give full usage",
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
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"b": "*",
				},
			},
		},
		{
			name: "explicit map access is listed",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param a
				*/
				{template .main}
					{let $c: $a['b'] /}
					{$c.d | json}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"d": "*",
					},
				},
			},
		},
		{
			name: "inexplicit map access is listed",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param a
				* @param b
				*/
				{template .main}
					{let $c: $a[$b] /}
					{$c.d}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"[?]": map[string]interface{}{
						"d": "*",
					},
				},
				"b": "*",
			},
		},
		{
			name: "explicit list access is listed",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param a
				*/
				{template .main}
					{$a[5]?.c | json}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"c": "*",
				},
			},
		},
		{
			name: "dot indexes are handled",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param a
				*/
				{template .main}
					{$a.5.c | json}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"c": "*",
				},
			},
		},
		{
			name: "let creating alias",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param a
				* @param b
				* @param c
				*/
				{template .main}
					{let $x: $a/}
					{let $y: $b ?: $c/}
					{let $w: $a ? $b: $c/}
					{let $u}
						text {$c.e}
					{/let}
					{$x.z}
					{$y.z}
					{$w.v}
					{$u}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"z": "*",
				},
				"b": map[string]interface{}{
					"z": "*",
					"v": "*",
				},
				"c": map[string]interface{}{
					"e": "*",
					"z": "*",
					"v": "*",
				},
			},
		},
		{
			name: "supports concatenation in let",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param a
				*/
				{template .main}
					{let $z: $a.b + ' ' + $a.c /}
					{$z}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"b": "*",
					"c": "*",
				},
			},
		},
		{
			name: "assignment doesn't leak up",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param a
				* @param? b
				*/
				{template .main}
					{let $x: $a/}
					{if true}
						{let $x: $b/}
						{$x.y}
					{/if}
					{$x.z}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"z": "*",
				},
				"b": map[string]interface{}{
					"y": "*",
				},
			},
		},
		{
			name: "if doesn't introduce full usage",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param a
				*/
				{template .main}
					{if $a}
						{$a.b}
					{/if}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"b": "*",
				},
			},
		},
		{
			name: "handles switch statements",
			templates: map[string]string{
				"test.soy": `
				{namespace test}
				/**
				* @param a
				*/
				{template .main}
					{switch $a.b}
						{case 'value1'}
							{$a.value1}
						{case 'value2'}
							{$a.value2}
						{default}
							{$a.default}
					{/switch}
				{/template}
			`,
			},
			templateName: "test.main",
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"b":       "*",
					"value1":  "*",
					"value2":  "*",
					"default": "*",
				},
			},
		},
	}
	testAnalyze(t, tests)
}

type analyzeTest struct {
	name         string
	templates    map[string]string
	templateName string
	expected     map[string]interface{}
	expectedErr  error
}

func testAnalyze(t *testing.T, tests []analyzeTest) {
	t.Helper()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Helper()
			bundle := soy.NewBundle()
			for name, content := range test.templates {
				bundle = bundle.AddTemplateString(name, content)
			}
			registry, err := bundle.Compile()
			if err != nil {
				t.Fatal(err)
			}
			got, err := soyusage.AnalyzeTemplate(test.templateName, registry)
			must.BeEqual(t, test.expected, mapUsage(got))
			must.BeEqualErrors(t, test.expectedErr, err)
			if t.Failed() {
				t.Log(jsonSprint(mapUsageFull(registry, got)))
			}
		})
	}
}

func jsonSprint(value interface{}) string {
	out, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(out)
}

func mapUsage(params soyusage.Params) map[string]interface{} {
	var out = make(map[string]interface{})
	for name, param := range params {
		var mappedParam interface{} = mapUsage(param.Children)
		sort.Slice(param.Usage, func(i int, j int) bool {
			var order = map[soyusage.UsageType]int{
				soyusage.UsageUnknown: 10,
				soyusage.UsageFull:    9,
				soyusage.UsageMeta:    7,
				soyusage.UsageExists:  6,
			}
			return order[param.Usage[i].Type] < order[param.Usage[j].Type]
		})
		for _, usage := range param.Usage {
			var newValue string
			switch usage.Type {
			case soyusage.UsageMeta:
				if len(param.Children) == 0 {
					newValue = "m"
				}
			case soyusage.UsageExists:
				if len(param.Children) == 0 {
					newValue = "e"
				}
			case soyusage.UsageFull:
				newValue = "*"
			case soyusage.UsageUnknown:
				newValue = "?"
			}
			if newValue != "" {
				mappedParam = newValue
			}
		}
		out[name.String()] = mappedParam
	}
	return out
}

func mapUsageFull(registry *template.Registry, params soyusage.Params) map[string]interface{} {
	var out = make(map[string]interface{})
	for name, param := range params {
		var paramOut = map[string]interface{}{
			"Children": mapUsageFull(registry, param.Children),
		}

		var usageList []interface{}
		for _, usage := range param.Usage {
			var usageValue = map[string]interface{}{}
			switch usage.Type {
			case soyusage.UsageFull:
				usageValue["Type"] = "Full"
				usageValue["String"] = usage.Node().String()
			case soyusage.UsageUnknown:
				usageValue["Type"] = "Unknown"
			case soyusage.UsageMeta:
				usageValue["Type"] = "Meta"
			case soyusage.UsageExists:
				usageValue["Type"] = "Exists"
			case soyusage.UsageReference:
				usageValue["Type"] = "Reference"
			default:
				usageValue["Type"] = fmt.Sprint(usage.Type)
			}
			usageValue["Template"] = usage.Template
			usageValue["File"] = registry.Filename(usage.Template)
			usageValue["Col"] = registry.ColNumber(usage.Template, usage.Node())
			usageValue["Pos"] = usage.Node().Position()
			usageValue["Line"] = registry.LineNumber(usage.Template, usage.Node())

			usageList = append(usageList, usageValue)
		}
		paramOut["Usage"] = usageList
		out[name.String()] = paramOut
	}
	return out
}

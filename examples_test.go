package soyusage_test

import (
	"fmt"

	"github.com/yext/soy"
	"github.com/yext/soyusage"
)

func ExampleAnalyzeTemplate() {
	bundle := soy.NewBundle()
	bundle = bundle.AddTemplateString(
		"example.soy",
		`
{namespace example}
/**
* @param a
*/
{template .main}
	{$a.b}
	{myFunc($a.c)}
{/template}
		`,
	)
	registry, _ := bundle.Compile()
	tree, _ := soyusage.AnalyzeTemplate("example.main", registry)
	// Get usage information for $a.b
	usage := tree[soyusage.Name("a")].Children[soyusage.Name("b")].Usage
	// Because $a.b was printed, all of it's child fields are used
	if usage[0].Type == soyusage.UsageFull {
		fmt.Println("$a.b: Full usage")
	}

	// Get usage information for $a.c
	usage = tree[soyusage.Name("a")].Children[soyusage.Name("c")].Usage
	// Because $a.c was used as a parameter to a function, usage of child fields
	// cannot be known.
	if usage[0].Type == soyusage.UsageUnknown {
		fmt.Println("$a.c: Unknown usage")
	}

	// Output: $a.b: Full usage
	// $a.c: Unknown usage
}

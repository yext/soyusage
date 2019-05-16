package soyusage

import (
	"github.com/robfig/soy/ast"
)

const (
	usageUndefined UsageType = iota
	// UsageFull indicates that the whole parameter is used
	// This usually means that it's value was printed.
	UsageFull
	// UsageUnknown indicates that the parameter's usage cannot be
	// known. This usually means that it was used as a parameter
	// to a function.
	UsageUnknown
	// UsageMeta indicates that a meta property of the parameter was used (length, isFirst, isLast, etc)
	UsageMeta
	// UsageReference indicates that the parameter was used in a reference,
	// such as a parameter to a call or assigned to a variable.
	UsageReference
	// UsageExists indicates that the parameter was used in an if check to determine
	// if it had a value
	UsageExists
)

// Usage provides details of the manner in which a param was used.
type (
	// Params specifies a collection of parameters, organized by name.
	Params map[string]*Param

	// Param defines a single parameter, or field within a parent parameter.
	// It contains details of how the parameter was used within the analyzed templates.
	Param struct {
		// Children identifies all fields within this Param
		Children Params
		// Usage describes how this parameter or field was used
		Usage UsageByTemplate

		// A constant value for this param
		constant interface{}
	}

	// UsageType specifies the manner in which a parameter was used.
	UsageType int

	Usage struct {
		// Type indicates how the parameter was used, see constants for details.
		Type UsageType
		// Template provides the name of the template containing the usage.
		Template string

		node ast.Node
	}
	// UsageByTemplate organizes usages by where they occurred
	UsageByTemplate map[string][]Usage
)

func (p *Param) addUsageToLeaves(usage Usage) {
	p.Usage[usage.Template] = append(p.Usage[usage.Template], usage)
	if len(p.Children) == 0 {
		p.Usage[usage.Template] = append(p.Usage[usage.Template], usage)
		return
	}
	for _, child := range p.Children {
		child.addUsageToLeaves(usage)
	}
}

func (p *Param) addChild(name string, child *Param) *Param {
	p.Children[name] = child
	return child
}

func (p *Param) getChildOrNew(name string) *Param {
	if child, exists := p.Children[name]; exists {
		return child
	}
	return p.addChild(name, newParam())
}

// Node provides a reference to the AST node where the param was used.
// This can be used with the Template name and template.Registry to identify
// where in the file the usage occurred.
func (u Usage) Node() ast.Node {
	return u.node
}

func (p *Param) isConstant() bool {
	return p.constant != nil
}

func newParam() *Param {
	return &Param{
		Children: make(Params),
		Usage:    make(UsageByTemplate),
	}
}

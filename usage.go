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
	// UsageExists indicates that the parameter was used in an if check to determine
	// if it had a value
	UsageExists
	// UsageReference indicates that the parameter was used in a reference,
	// such as a parameter to a call or assigned to a variable.
	UsageReference
)

// Usage provides details of the manner in which a param was used.
type (
	// Params specifies a collection of parameters, organized by name.
	Params map[Identifier]*Param

	// Param defines a single parameter, or field within a parent parameter.
	// It contains details of how the parameter was used within the analyzed templates.
	Param struct {
		// Children identifies all fields within this Param
		Children Params
		// Usage describes how this parameter or field was used
		Usage []Usage

		// A constant value for this param
		constant interface{}
	}

	// Identifier names a parameter
	Identifier interface {
		String() string
	}

	Name     string
	MapIndex struct{}

	// UsageType specifies the manner in which a parameter was used.
	UsageType int

	Usage struct {
		// Type indicates how the parameter was used, see constants for details.
		Type UsageType
		// Template provides the name of the template containing the usage.
		Template string

		node ast.Node
	}
)

func (n Name) String() string {
	return string(n)
}

func (MapIndex) String() string {
	return "[?]"
}

func (p *Param) addUsageToLeaves(usage Usage) {
	if len(p.Children) == 0 {
		for _, otherUsage := range p.Usage {
			if otherUsage.Template == usage.Template &&
				otherUsage.Type == usage.Type &&
				otherUsage.node.Position() == usage.node.Position() {
				return
			}
		}
		p.Usage = append(p.Usage, usage)
		return
	}
	for _, child := range p.Children {
		child.addUsageToLeaves(usage)
	}
}

func (p *Param) addChild(name Identifier, child *Param) *Param {
	p.Children[name] = child
	return child
}

func (p *Param) getChildOrNew(name Identifier) *Param {
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
	}
}

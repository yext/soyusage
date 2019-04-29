package soyusage

import (
	"fmt"

	"github.com/robfig/soy/ast"
	"github.com/robfig/soy/template"
)

var _ error = &usageError{}

type usageError struct {
	message string
	cause   error

	node     ast.Node
	template string
	registry *template.Registry
}

func newErrorf(s *scope, node ast.Node, message string, args ...interface{}) *usageError {
	return &usageError{
		message: fmt.Sprintf(message, args...),

		node:     node,
		template: s.templateName,
		registry: s.registry,
	}
}

func wrapError(s *scope, node ast.Node, err error) *usageError {
	return &usageError{
		cause: err,

		node:     node,
		template: s.templateName,
		registry: s.registry,
	}
}

func (u *usageError) Error() string {
	if u.cause != nil {
		return u.cause.Error()
	}
	return fmt.Sprintf("%s (line %d, col %d): %v\n%v", u.Filename(), u.Row(), u.Col(), u.message, u.node)
}

func (u *usageError) Filename() string {
	return u.registry.Filename(u.template)
}

func (u *usageError) Col() int {
	return u.registry.ColNumber(u.template, u.node)
}

func (u *usageError) Row() int {
	return u.registry.LineNumber(u.template, u.node)
}

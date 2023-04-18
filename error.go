package soyusage

import (
	"fmt"

	"github.com/yext/soy/ast"
	"github.com/yext/soy/template"
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
	if err == nil {
		return nil
	}
	return &usageError{
		cause: err,

		node:     node,
		template: s.templateName,
		registry: s.registry,
	}
}

func (u *usageError) Error() string {
	if u.cause != nil {
		return fmt.Sprintf("%v\n%v", u.cause.Error(), u.posText())
	}
	return fmt.Sprintf("%v (%v)", u.message, u.posText())
}

func (u *usageError) posText() string {
	var shortNode = fmt.Sprint(u.node)
	if len(shortNode) > 10 {
		shortNode = shortNode[:10]
	}
	return fmt.Sprintf("%s, line %d, col %d near %q", u.filename(), u.row(), u.col(), shortNode)
}

func (u *usageError) filename() string {
	return u.registry.Filename(u.template)
}

func (u *usageError) col() int {
	return u.registry.ColNumber(u.template, u.node)
}

func (u *usageError) row() int {
	return u.registry.LineNumber(u.template, u.node)
}

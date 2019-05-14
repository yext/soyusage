package soyusage

import "github.com/robfig/soy/template"

// scope represents the usage at the current position in the stack
type scope struct {
	registry     *template.Registry
	templateName string
	callStack    []*scope
	parameters   Params
	variables    map[string][]*Param
}

// isRecursive returns true iff this scope is part of a recursive call stack
// A call stack is considered recursive if any templates appear in the call stack more than once.
func (s *scope) isRecursive() bool {
	return s.callCycles() > 1
}

// callCycles returns the number of cycles in this scope's call stack
func (s *scope) callCycles() int {
	var cycles int
	for _, stackCall := range s.callStack {
		if stackCall.templateName == s.templateName {
			cycles++
		}
	}
	return cycles
}

func (s *scope) finalCycle() *scope {
	var cycle *scope
	for _, stackCall := range s.callStack {
		if stackCall.templateName == s.templateName {
			cycle = stackCall
		}
	}
	return cycle
}

// inner creates a new scope "inside" the current scope
// The new scope has all the same state, but a new set of variables
// is created so assignments don't escape up the stack.
func (s *scope) inner() *scope {
	out := &scope{
		registry:     s.registry,
		templateName: s.templateName,
		callStack:    nil,
		parameters:   s.parameters,
		variables:    make(map[string][]*Param),
	}

	for _, template := range s.callStack {
		out.callStack = append(out.callStack, template)
	}

	for name, params := range s.variables {
		out.variables[name] = params
	}
	return out
}

// call creates a child scope as a result of a call
func (s *scope) call(templateName string) *scope {
	out := &scope{
		registry:     s.registry,
		templateName: templateName,
		parameters:   s.parameters,
		variables:    make(map[string][]*Param),
	}

	for _, template := range s.callStack {
		out.callStack = append(out.callStack, template)
	}
	out.callStack = append(out.callStack, s)
	return out
}

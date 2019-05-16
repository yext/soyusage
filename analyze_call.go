package soyusage

import (
	"github.com/robfig/soy/ast"
)

func analyzeCall(
	s *scope,
	call *ast.CallNode,
) error {
	var scopes []*scope
	template, found := s.registry.Template(call.Name)
	if !found {
		return newErrorf(s, call, "template not found: %s", call.Name)
	}

	callScope := s.call(call.Name)
	if !call.AllData {
		callScope.parameters = make(Params)
	}

	if call.Data != nil {
		variables, err := extractVariables(s, call.Data)
		if err != nil {
			return wrapError(s, call.Data, err)
		}
		for _, param := range variables {
			dataScope := s.call(call.Name)
			dataScope.parameters = param.Children
			scopes = append(scopes, dataScope)
		}
	} else {
		scopes = append(scopes, callScope)
	}

	if callScope.callCycles() > s.config.RecursionDepth {
		for _, param := range s.parameters {
			param.Usage[callScope.templateName] = append(param.Usage[callScope.templateName], Usage{
				Type:     UsageFull,
				Template: callScope.templateName,
				node:     call,
			})
		}
		for _, variables := range s.variables {
			for _, variable := range variables {
				variable.addUsageToLeaves(Usage{
					Type:     UsageFull,
					Template: callScope.templateName,
					node:     call,
				})
			}
		}
		return nil
	}

	for _, scope := range scopes {
		for _, parameter := range call.Params {
			switch v := parameter.(type) {
			case *ast.CallParamContentNode:
				err := analyzeNode(s, UsageFull, v.Content)
				if err != nil {
					return wrapError(s, parameter, err)
				}
				constants, err := extractConstantVariables(s, v.Content)
				if err != nil {
					return wrapError(s, parameter, err)
				}
				scope.variables[v.Key] = append(scope.variables[v.Key], constants...)
			case *ast.CallParamValueNode:
				variables, err := extractVariables(s, v.Value)
				if err != nil {
					return wrapError(s, parameter, err)
				}
				scope.variables[v.Key] = append(scope.variables[v.Key], variables...)
			}
		}
		if err := analyzeNode(scope, usageUndefined, template.Node); err != nil {
			return wrapError(s, template.Node, err)
		}
	}
	return nil
}

package soyusage

import "github.com/robfig/soy/ast"

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
	if callScope.isRecursive() {
		return analyzeRecursiveCall(s, call)
	}
	if !call.AllData {
		callScope.parameters = make(Params)
	}

	scopes = []*scope{
		callScope,
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

func analyzeRecursiveCall(
	s *scope,
	call *ast.CallNode,
) error {
	var scopes []*scope
	callScope := s.call(call.Name)
	s = callScope.finalCycle()
	if call.AllData {
		expectedParams := templateParams(s)
		for _, expected := range expectedParams {
			for _, param := range s.variables[expected] {
				dataScope := s.call(call.Name)
				for _, expected := range expectedParams {
					if param, hasParameter := s.parameters[expected]; hasParameter {
						p := newParam(expected)
						p.RecursesTo = param
						dataScope.parameters[expected] = p
					}
				}
				for _, value := range param.Children {
					p := newParam(value.name)
					p.RecursesTo = value
					dataScope.parameters[value.name] = p
				}
				scopes = append(scopes, dataScope)
			}
		}
	}

	if call.Data != nil {
		variables, err := extractVariables(s, call.Data)
		if err != nil {
			return wrapError(s, call.Data, err)
		}
		expectedParams := templateParams(s)
		for _, param := range variables {
			dataScope := s.call(call.Name)
			for _, expected := range expectedParams {
				if param, hasParameter := s.parameters[expected]; hasParameter {
					p := newParam(expected)
					p.RecursesTo = param
					dataScope.parameters[expected] = p
				}
			}
			for _, value := range param.Children {
				p := newParam(value.name)
				p.RecursesTo = value
				dataScope.parameters[value.name] = p
			}
			scopes = append(scopes, dataScope)
		}
	}

	for _, scope := range scopes {
		for _, parameter := range call.Params {
			switch v := parameter.(type) {
			case *ast.CallParamContentNode:
				err := analyzeNode(s, UsageFull, v.Content)
				if err != nil {
					return wrapError(s, parameter, err)
				}
			case *ast.CallParamValueNode:
				variables, err := extractVariables(s, v.Value)
				if err != nil {
					return wrapError(s, parameter, err)
				}
				for _, variable := range variables {
					if variable.isConstant() {
						continue
					}
					p := newParam(variable.name)
					p.RecursesTo = variable
					scope.parameters[variable.name] = p
				}
			}
		}
	}
	return nil
}

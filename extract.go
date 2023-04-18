package soyusage

import (
	"github.com/yext/soy/data"
)

// Extract returns a version of the input data containing only
// the values specified in the provided usage analysis.
func Extract(in data.Value, params Params) data.Value {
	var (
		out          = make(data.Map)
		inMap, isMap = in.(data.Map)
	)
	if !isMap {
		return in
	}

	for paramName, param := range params {
		inValue := inMap[paramName.String()]
		if (MapIndex{}) == paramName {
			values := extractMap(param, inMap)
			for name, value := range values {
				out[name] = value
			}
		}

		var outVal data.Value
		outVal = extractParam(param, inValue)
		if outVal != nil {
			out[paramName.String()] = outVal
		}
	}
	return out
}

func extractMap(param *Param, in data.Value) data.Map {
	var (
		out          = make(data.Map)
		inMap, isMap = in.(data.Map)
	)
	if !isMap {
		return nil
	}
	for name, value := range inMap {
		out[name] = Extract(value, param.Children)
	}
	return out
}

func extractParam(param *Param, in data.Value) data.Value {
	if in == nil {
		return nil
	}
	if listValue, isList := in.(data.List); isList {
		var outList data.List
		for _, value := range listValue {
			outList = append(outList, extractParam(param, value))
		}
		return outList
	}
	var (
		isFull   bool
		isExists bool
	)
	for _, usage := range param.Usage {
		switch usage.Type {
		case UsageFull, UsageUnknown, UsageMeta:
			isFull = true
		case UsageExists:
			isExists = true
		}
	}
	if isFull {
		return in
	}
	if isExists && len(param.Children) == 0 {
		return data.String("")
	}
	return Extract(in, param.Children)
}

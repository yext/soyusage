package soyusage

import (
	"github.com/robfig/soy/data"
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

	for name, param := range params {
		var outVal data.Value
		outVal = extractParam(param, inMap[name])
		if outVal != nil {
			out[name] = outVal
		}
	}
	return out
}

func extractParam(param *Param, in data.Value) data.Value {
	if in == nil {
		return nil
	}
	for _, usages := range param.Usage {
		for _, usage := range usages {
			switch usage.Type {
			case UsageFull, UsageUnknown:
				return in
			}
		}
	}
	return Extract(in, param.Children)
}

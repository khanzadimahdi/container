package resolver

import (
	"reflect"
)

// paramsOptions is an internal struct used to hold parameters for resolve-time injection.
type paramsOptions struct {
	Params []reflect.Value
}

// DefaultParamsOptions returns a new instance of paramsOptions with default values.
func DefaultParamsOptions() *paramsOptions {
	return &paramsOptions{}
}

// ParamsOption is a functional option for resolve-time parameters.
type ParamsOption func(*paramsOptions)

// ResolveParams is a functional option for resolve-time parameters.
func ResolveParams(params ...any) ParamsOption {
	return func(o *paramsOptions) {
		for _, param := range params {
			o.Params = append(o.Params, reflect.ValueOf(param))
		}
	}
}

// resolveOptions holds runtime parameters for resolve-time injection.
type resolveOptions struct {
	Name   string
	Params []reflect.Value
}

// DefaultResolveOptions returns a new instance of resolveOptions with default values.
func DefaultResolveOptions() *resolveOptions {
	return &resolveOptions{}
}

// ResolveOption is a functional option for resolve-time parameters.
type ResolveOption func(*resolveOptions)

// WithName specifies the name of the binding to resolve.
func WithName(name string) ResolveOption {
	return func(o *resolveOptions) {
		o.Name = name
	}
}

// WithParams provides runtime values used to satisfy resolver arguments.
func WithParams(params ...any) ResolveOption {
	return func(o *resolveOptions) {
		for _, param := range params {
			o.Params = append(o.Params, reflect.ValueOf(param))
		}
	}
}

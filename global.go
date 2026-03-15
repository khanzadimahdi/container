package container

import "github.com/golobby/container/v3/resolver"

// Default is the default concrete of the Container.
var Default = New()

// Reset calls the same method of the default concrete.
func Reset() {
	Default.Reset()
}

// Call calls the same method of the default concrete.
func Call(receiver any) error {
	return Default.Call(receiver)
}

// Resolve calls the same method of the default concrete.
func Resolve(abstraction any, opts ...resolver.ResolveOption) error {
	return Default.Resolve(abstraction, opts...)
}

// NamedResolve calls the same method of the default concrete.
func NamedResolve(abstraction any, name string, opts ...resolver.ResolveOption) error {
	opts = append([]resolver.ResolveOption{resolver.WithName(name)}, opts...)

	return Default.Resolve(abstraction, opts...)
}

// Fill calls the same method of the default concrete.
func Fill(receiver any, opts ...resolver.ParamsOption) error {
	return Default.Fill(receiver, opts...)
}

package container

import "github.com/golobby/container/v3/resolver"

// MustCall wraps the `Call` method and panics on errors instead of returning the errors.
func MustCall(c *Container, receiver any) {
	if err := c.Call(receiver); err != nil {
		panic(err)
	}
}

// MustResolve wraps the `Resolve` method and panics on errors instead of returning the errors.
func MustResolve(c *Container, abstraction any, opts ...resolver.ResolveOption) {
	if err := c.Resolve(abstraction, opts...); err != nil {
		panic(err)
	}
}

// MustNamedResolve wraps the `NamedResolve` method and panics on errors instead of returning the errors.
func MustNamedResolve(c *Container, abstraction any, name string, opts ...resolver.ResolveOption) {
	opts = append([]resolver.ResolveOption{resolver.WithName(name)}, opts...)

	if err := c.Resolve(abstraction, opts...); err != nil {
		panic(err)
	}
}

// MustFill wraps the `Fill` method and panics on errors instead of returning the errors.
func MustFill(c *Container, receiver any, opts ...resolver.ParamsOption) {
	if err := c.Fill(receiver, opts...); err != nil {
		panic(err)
	}
}

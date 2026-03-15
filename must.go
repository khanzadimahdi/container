package container

import "github.com/golobby/container/v3/resolve"

// MustCall wraps the `Call` method and panics on errors instead of returning the errors.
func MustCall(c *Container, receiver any, opts ...resolve.ResolveOption) {
	if err := c.Call(receiver, opts...); err != nil {
		panic(err)
	}
}

// MustResolve wraps the `Resolve` method and panics on errors instead of returning the errors.
func MustResolve(c *Container, abstraction any, opts ...resolve.ResolveOption) {
	if err := c.Resolve(abstraction, opts...); err != nil {
		panic(err)
	}
}

// MustNamedResolve wraps the `NamedResolve` method and panics on errors instead of returning the errors.
func MustNamedResolve(c *Container, abstraction any, name string, opts ...resolve.ResolveOption) {
	opts = append([]resolve.ResolveOption{resolve.WithName(name)}, opts...)

	if err := c.Resolve(abstraction, opts...); err != nil {
		panic(err)
	}
}

// MustFill wraps the `Fill` method and panics on errors instead of returning the errors.
func MustFill(c *Container, receiver any, opts ...resolve.ResolveOption) {
	if err := c.Fill(receiver, opts...); err != nil {
		panic(err)
	}
}

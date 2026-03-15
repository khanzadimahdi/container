package bind

// bindOptions holds the configuration options for a binding.
type bindOptions struct {
	Name      string
	Singleton bool
	Lazy      bool
}

func DefaultBindOptions() *bindOptions {
	return &bindOptions{}
}

// BindOption is a functional option for configuring a binding.
type BindOption func(*bindOptions)

// WithName sets a name for the binding, enabling multiple concretes per abstraction.
func WithName(name string) BindOption {
	return func(o *bindOptions) {
		o.Name = name
	}
}

// Singleton marks the binding as a singleton (one shared instance).
func Singleton() BindOption {
	return func(o *bindOptions) {
		o.Singleton = true
	}
}

// Lazy defers the resolver invocation until the first time the binding is resolved.
func Lazy() BindOption {
	return func(o *bindOptions) {
		o.Lazy = true
	}
}

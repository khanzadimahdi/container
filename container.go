// Package container is a lightweight yet powerful IoC container for Go projects.
// It provides an easy-to-use interface and performance-in-mind container to be your ultimate requirement.
package container

import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/golobby/container/v3/binder"
	"github.com/golobby/container/v3/resolver"
)

var (
	// errNonFunctionResolver is returned when the resolver is not a function.
	errNonFunctionResolver = errors.New("container: the resolver must be a function")

	// errInvalidResolver is returned when the resolver function signature is invalid.
	errInvalidResolver = errors.New("container: resolver function signature is invalid - it must return abstract, or abstract and error")

	// errResolverDependsOnAbstract is returned when the resolver function depends on the abstract it returns.
	errResolverDependsOnAbstract = errors.New("container: resolver function signature is invalid - depends on abstract it returns")

	// errInvalidAbstraction is returned when the abstraction provided to Resolve is invalid.
	errInvalidAbstraction = errors.New("container: invalid abstraction")

	// errEncounteredError is returned when an error is encountered while making a concrete.
	errEncounteredError = errors.New("container: encountered error while making concrete")

	// errNoConcreteFound is returned when no concrete is found for the given abstraction.
	errNoConcreteFound = errors.New("container: no concrete found for the given abstraction")

	// errInvalidFunction is returned when the function provided to Call is invalid.
	errInvalidFunction = errors.New("container: invalid function")

	// errInvalidFunctionSignature is returned when the function signature is invalid.
	errInvalidFunctionSignature = errors.New("container: receiver function signature is invalid")

	// errInvalidStructure is returned when the structure provided to Fill is invalid.
	errInvalidStructure = errors.New("container: invalid structure")

	// errInvalidStructTag is returned when a struct field has an invalid struct tag.
	errInvalidStructTag = errors.New("container: invalid struct tag")

	// errCannotMakeField is returned when a field with the `container` tag cannot be made.
	errCannotMakeField = errors.New("container: cannot make field")
)

// binding holds a resolver and a concrete (if already resolved).
// It is the break for the Container wall!
type binding struct {
	resolver    any  // resolver is the function that is responsible for making the concrete.
	concrete    any  // concrete is the stored instance for singleton bindings.
	isSingleton bool // isSingleton is true if the binding is a singleton.
}

// make resolves the binding if needed and returns the resolved concrete.
func (b *binding) make(c *Container, params []reflect.Value) (any, error) {
	if b.concrete != nil {
		return b.concrete, nil
	}

	retVal, err := c.invoke(b.resolver, params)
	if err != nil {
		return nil, err
	}

	if b.isSingleton {
		b.concrete = retVal
	}

	return retVal, nil
}

// registerar is a map that holds the bindings for each abstraction and name.
type registerar map[reflect.Type]map[string]*binding

// Container holds the bindings and provides methods to interact with them.
// It is the entry point in the package.
type Container struct {
	bindings registerar
}

// New creates a new concrete of the Container.
func New() *Container {
	return &Container{
		bindings: make(registerar),
	}
}

// Reset deletes all the existing bindings and empties the container.
func (c *Container) Reset() {
	clear(c.bindings)
}

// Bind maps an abstraction to concrete and instantiates if it is a singleton binding.
func (c *Container) Bind(resolver any, opts ...binder.BindOption) error {
	options := binder.DefaultBindOptions()
	for _, o := range opts {
		o(options)
	}

	reflectedResolver := reflect.TypeOf(resolver)
	if reflectedResolver.Kind() != reflect.Func {
		return errNonFunctionResolver
	}

	if reflectedResolver.NumOut() > 0 {
		if _, exist := c.bindings[reflectedResolver.Out(0)]; !exist {
			c.bindings[reflectedResolver.Out(0)] = make(map[string]*binding)
		}
	}

	if err := c.validateResolverFunction(reflectedResolver); err != nil {
		return err
	}

	var concrete any
	if !options.Lazy {
		var err error
		concrete, err = c.invoke(resolver, nil)
		if err != nil {
			return err
		}
	}

	if options.Singleton {
		c.bindings[reflectedResolver.Out(0)][options.Name] = &binding{resolver: resolver, concrete: concrete, isSingleton: options.Singleton}
	} else {
		c.bindings[reflectedResolver.Out(0)][options.Name] = &binding{resolver: resolver, isSingleton: options.Singleton}
	}

	return nil
}

// Resolve takes an abstraction (reference of an interface type) and fills it with the related concrete.
func (c *Container) Resolve(abstraction any, opts ...resolver.ResolveOption) error {
	options := resolver.DefaultResolveOptions()
	for _, o := range opts {
		o(options)
	}

	receiverType := reflect.TypeOf(abstraction)
	if receiverType == nil {
		return errInvalidAbstraction
	}

	if receiverType.Kind() == reflect.Ptr {
		elem := receiverType.Elem()

		if concrete, exist := c.bindings[elem][options.Name]; exist {
			if instance, err := concrete.make(c, options.Params); err == nil {
				reflect.ValueOf(abstraction).Elem().Set(reflect.ValueOf(instance))
				return nil
			} else {
				return fmt.Errorf("%w for: %s. Error encountered: %w", errEncounteredError, elem.String(), err)
			}
		}

		return fmt.Errorf("%w; the abstraction is: %s", errNoConcreteFound, elem.String())
	}

	return errInvalidAbstraction
}

// Call takes a receiver function with one or more arguments of the abstractions (interfaces).
// It invokes the receiver function and passes the related concretes.
func (c *Container) Call(function any) error {
	receiverType := reflect.TypeOf(function)
	if receiverType == nil || receiverType.Kind() != reflect.Func {
		return errInvalidFunction
	}

	arguments, err := c.arguments(function, nil)
	if err != nil {
		return err
	}

	result := reflect.ValueOf(function).Call(arguments)

	if len(result) == 0 {
		return nil
	} else if len(result) == 1 && result[0].CanInterface() {
		if result[0].IsNil() {
			return nil
		}
		if err, ok := result[0].Interface().(error); ok {
			return err
		}
	}

	return errInvalidFunctionSignature
}

// Fill takes a struct and resolves the fields with the tag `container:"inject"`
func (c *Container) Fill(structure any, opts ...resolver.ParamsOption) error {
	receiverType := reflect.TypeOf(structure)
	if receiverType == nil {
		return errInvalidStructure
	}

	if receiverType.Kind() == reflect.Ptr {
		elem := receiverType.Elem()
		if elem.Kind() == reflect.Struct {
			s := reflect.ValueOf(structure).Elem()

			options := resolver.DefaultParamsOptions()
			for _, o := range opts {
				o(options)
			}

			for i := 0; i < s.NumField(); i++ {
				f := s.Field(i)

				if t, exist := s.Type().Field(i).Tag.Lookup("container"); exist {
					var name string

					switch t {
					case "type":
						name = ""
					case "name":
						name = s.Type().Field(i).Name
					default:
						return fmt.Errorf("%w; the field is: %s", errInvalidStructTag, s.Type().Field(i).Name)
					}

					if concrete, exist := c.bindings[f.Type()][name]; exist {
						instance, err := concrete.make(c, options.Params)
						if err != nil {
							return err
						}

						ptr := reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
						ptr.Set(reflect.ValueOf(instance))

						continue
					}

					return fmt.Errorf("%w; the field is: %s", errCannotMakeField, s.Type().Field(i).Name)
				}
			}

			return nil
		}
	}

	return errInvalidStructure
}

// validateResolverFunction checks if the resolver function signature is valid.
func (c *Container) validateResolverFunction(funcType reflect.Type) error {
	retCount := funcType.NumOut()

	if retCount == 0 || retCount > 2 {
		return errInvalidResolver
	}

	resolveType := funcType.Out(0)
	for i := 0; i < funcType.NumIn(); i++ {
		if funcType.In(i) == resolveType {
			return errResolverDependsOnAbstract
		}
	}

	return nil
}

// invoke calls the provided function with the given parameters and returns the result or an error if it occurs.
func (c *Container) invoke(function any, params []reflect.Value) (any, error) {
	arguments, err := c.arguments(function, params)
	if err != nil {
		return nil, err
	}

	values := reflect.ValueOf(function).Call(arguments)
	if len(values) == 2 && values[1].CanInterface() {
		if err, ok := values[1].Interface().(error); ok {
			return values[0].Interface(), err
		}
	}

	return values[0].Interface(), nil
}

// arguments returns the list of resolved arguments for a function.
func (c *Container) arguments(function any, params []reflect.Value) ([]reflect.Value, error) {
	reflectedFunction := reflect.TypeOf(function)
	argumentsCount := reflectedFunction.NumIn()
	arguments := make([]reflect.Value, argumentsCount)
	usedParams := make([]bool, len(params))

	for i := 0; i < argumentsCount; i++ {
		abstraction := reflectedFunction.In(i)

		if value, ok := takeParam(abstraction, params, usedParams); ok {
			arguments[i] = value
			continue
		}

		if concrete, exist := c.concrete(abstraction); exist {
			instance, err := concrete.make(c, params)
			if err != nil {
				return nil, err
			}

			arguments[i] = reflect.ValueOf(instance)
		} else {
			return nil, fmt.Errorf("%w; the abstraction is: %s", errNoConcreteFound, abstraction.String())
		}
	}

	return arguments, nil
}

// takeParam checks if any of the provided parameters can be used to satisfy the given abstraction.
// It returns the first matching parameter and a boolean indicating if a match was found.
func takeParam(abstraction reflect.Type, params []reflect.Value, usedParams []bool) (reflect.Value, bool) {
	for i, param := range params {
		if usedParams[i] {
			continue
		}

		if param.Type().AssignableTo(abstraction) {
			usedParams[i] = true
			if param.Type() == abstraction {
				return param, true
			}

			return param.Convert(abstraction), true
		}
	}

	return reflect.Value{}, false
}

// concrete retrieves the binding for the given abstraction and name, checking for direct matches first and then for implementations.
func (c *Container) concrete(abstraction reflect.Type) (*binding, bool) {
	if concrete, exist := c.bindings[abstraction][""]; exist {
		return concrete, true
	}

	for boundAbstraction, namedConcretes := range c.bindings {
		if boundAbstraction.Implements(abstraction) {
			concrete, exists := namedConcretes[""]
			return concrete, exists
		}
	}

	return nil, false
}

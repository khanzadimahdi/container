package container_test

import (
	"errors"
	"testing"

	"github.com/golobby/container/v3"
	"github.com/golobby/container/v3/binder"
	"github.com/golobby/container/v3/resolver"
	"github.com/stretchr/testify/assert"
)

type Shape interface {
	SetArea(int)
	GetArea() int
}

type ReadOnlyShape interface {
	GetArea() int
}

type Circle struct {
	a int
}

// Ensure Circle implements Shape interface.
var _ Shape = &Circle{}

func (c *Circle) SetArea(a int) {
	c.a = a
}

func (c Circle) GetArea() int {
	return c.a
}

type Database interface {
	Connect() bool
}

type MySQL struct{}

// Ensure MySQL implements Database interface.
var _ Database = MySQL{}

func (m MySQL) Connect() bool {
	return true
}

type PostgreSQL struct {
	ready bool
}

// Ensure PostgreSQL implements Database interface.
var _ Database = PostgreSQL{}

func (d PostgreSQL) Connect() bool {
	return d.ready
}

var instance = container.New()

func TestContainer_Singleton(t *testing.T) {
	err := instance.Bind(func() Shape {
		return &Circle{a: 13}
	}, binder.Singleton())
	assert.NoError(t, err)

	err = instance.Call(func(s1 Shape) {
		s1.SetArea(666)
	})
	assert.NoError(t, err)

	err = instance.Call(func(s2 Shape) {
		a := s2.GetArea()
		assert.Equal(t, a, 666)
	})
	assert.NoError(t, err)
}

func TestContainer_SingletonLazy(t *testing.T) {
	err := instance.Bind(func() Shape {
		return &Circle{a: 13}
	}, binder.Singleton(), binder.Lazy())
	assert.NoError(t, err)

	err = instance.Call(func(s1 Shape) {
		s1.SetArea(666)
	})
	assert.NoError(t, err)

	err = instance.Call(func(s2 Shape) {
		a := s2.GetArea()
		assert.Equal(t, a, 666)
	})
	assert.NoError(t, err)
}

func TestContainer_Singleton_With_Missing_Dependency_Resolve(t *testing.T) {
	err := instance.Bind(func(db Database) Shape {
		return &Circle{a: 13}
	}, binder.Singleton())
	assert.EqualError(t, err, "container: no concrete found for the given abstraction; the abstraction is: container_test.Database")
}

func TestContainer_Singleton_With_Resolve_That_Returns_Nothing(t *testing.T) {
	err := instance.Bind(func() {}, binder.Singleton())
	assert.Error(t, err, "container: resolver function signature is invalid")
}

func TestContainer_SingletonLazy_With_Resolve_That_Returns_Nothing(t *testing.T) {
	err := instance.Bind(func() {}, binder.Singleton(), binder.Lazy())
	assert.Error(t, err, "container: resolver function signature is invalid")
}

func TestContainer_Singleton_With_Resolve_That_Returns_Error(t *testing.T) {
	err := instance.Bind(func() (Shape, error) {
		return nil, errors.New("app: error")
	}, binder.Singleton())
	assert.Error(t, err, "app: error")
}

func TestContainer_SingletonLazy_With_Resolve_That_Returns_Error(t *testing.T) {
	err := instance.Bind(func() (Shape, error) {
		return nil, errors.New("app: error")
	}, binder.Singleton(), binder.Lazy())
	assert.NoError(t, err)

	var s Shape
	err = instance.Resolve(&s)
	assert.Error(t, err, "app: error")
}

func TestContainer_SingletonLazy_With_Resolve_That_Returns_Error_Should_Not_Cache_Nil(t *testing.T) {
	resolvable := false
	err := instance.Bind(func() (Shape, error) {
		if resolvable {
			return &Circle{a: 5}, nil
		}
		return nil, errors.New("app: not ready")
	}, binder.Singleton(), binder.Lazy())
	assert.NoError(t, err)

	var s Shape
	err = instance.Resolve(&s)
	assert.Error(t, err)

	// Simulate meeting the expectation and resolving again.
	resolvable = true
	err = instance.Resolve(&s)
	assert.NoError(t, err)
	assert.Equal(t, 5, s.GetArea())
}

func TestContainer_Singleton_With_NonFunction_Resolver_It_Should_Fail(t *testing.T) {
	err := instance.Bind("STRING!", binder.Singleton())
	assert.EqualError(t, err, "container: the resolver must be a function")
}

func TestContainer_SingletonLazy_With_NonFunction_Resolver_It_Should_Fail(t *testing.T) {
	err := instance.Bind("STRING!", binder.Singleton(), binder.Lazy())
	assert.EqualError(t, err, "container: the resolver must be a function")
}

func TestContainer_Singleton_With_Resolvable_Arguments(t *testing.T) {
	err := instance.Bind(func() Shape {
		return &Circle{a: 666}
	}, binder.Singleton())
	assert.NoError(t, err)

	err = instance.Bind(func(s Shape) Database {
		assert.Equal(t, s.GetArea(), 666)
		return &MySQL{}
	}, binder.Singleton())
	assert.NoError(t, err)
}

func TestContainer_SingletonLazy_With_Resolvable_Arguments(t *testing.T) {
	err := instance.Bind(func() Shape {
		return &Circle{a: 666}
	}, binder.Singleton(), binder.Lazy())
	assert.NoError(t, err)

	err = instance.Bind(func(s Shape) Database {
		assert.Equal(t, s.GetArea(), 666)
		return &MySQL{}
	}, binder.Singleton(), binder.Lazy())
	assert.NoError(t, err)

	var s Shape
	err = instance.Resolve(&s)
	assert.NoError(t, err)
}

func TestContainer_Singleton_With_Non_Resolvable_Arguments(t *testing.T) {
	instance.Reset()

	err := instance.Bind(func(s Shape) Shape {
		return &Circle{a: s.GetArea()}
	}, binder.Singleton())
	assert.EqualError(t, err, "container: resolver function signature is invalid - depends on abstract it returns")
}

func TestContainer_SingletonLazy_With_Non_Resolvable_Arguments(t *testing.T) {
	instance.Reset()

	err := instance.Bind(func(s Shape) Shape {
		return &Circle{a: s.GetArea()}
	}, binder.Singleton(), binder.Lazy())
	assert.EqualError(t, err, "container: resolver function signature is invalid - depends on abstract it returns")
}

func TestContainer_NamedSingleton(t *testing.T) {
	err := instance.Bind(func() Shape {
		return &Circle{a: 13}
	}, binder.WithName("theCircle"), binder.Singleton())
	assert.NoError(t, err)

	var sh Shape
	err = instance.Resolve(&sh, resolver.WithName("theCircle"))
	assert.NoError(t, err)
	assert.Equal(t, sh.GetArea(), 13)
}

func TestContainer_NamedSingletonLazy(t *testing.T) {
	err := instance.Bind(func() Shape {
		return &Circle{a: 13}
	}, binder.WithName("theCircle"), binder.Singleton(), binder.Lazy())
	assert.NoError(t, err)

	var sh Shape
	err = instance.Resolve(&sh, resolver.WithName("theCircle"))
	assert.NoError(t, err)
	assert.Equal(t, sh.GetArea(), 13)
}

func TestContainer_Transient(t *testing.T) {
	err := instance.Bind(func() Shape {
		return &Circle{a: 666}
	})
	assert.NoError(t, err)

	err = instance.Call(func(s1 Shape) {
		s1.SetArea(13)
	})
	assert.NoError(t, err)

	err = instance.Call(func(s2 Shape) {
		a := s2.GetArea()
		assert.Equal(t, a, 666)
	})
	assert.NoError(t, err)
}

func TestContainer_TransientLazy(t *testing.T) {
	err := instance.Bind(func() Shape {
		return &Circle{a: 666}
	}, binder.Lazy())
	assert.NoError(t, err)

	err = instance.Call(func(s1 Shape) {
		s1.SetArea(13)
	})
	assert.NoError(t, err)

	err = instance.Call(func(s2 Shape) {
		a := s2.GetArea()
		assert.Equal(t, a, 666)
	})
	assert.NoError(t, err)
}

func TestContainer_Transient_With_Resolve_That_Returns_Nothing(t *testing.T) {
	err := instance.Bind(func() {})
	assert.Error(t, err, "container: resolver function signature is invalid")
}

func TestContainer_TransientLazy_With_Resolve_That_Returns_Nothing(t *testing.T) {
	err := instance.Bind(func() {}, binder.Lazy())
	assert.Error(t, err, "container: resolver function signature is invalid")
}

func TestContainer_Transient_With_Resolve_That_Returns_Error(t *testing.T) {
	err := instance.Bind(func() (Shape, error) {
		return nil, errors.New("app: error")
	})
	assert.Error(t, err, "app: error")

	firstCall := true
	err = instance.Bind(func() (Database, error) {
		if firstCall {
			firstCall = false
			return &MySQL{}, nil
		}
		return nil, errors.New("app: second call error")
	})
	assert.NoError(t, err)

	var db Database
	err = instance.Resolve(&db)
	assert.Error(t, err, "app: second call error")
}

func TestContainer_TransientLazy_With_Resolve_That_Returns_Error(t *testing.T) {
	err := instance.Bind(func() (Shape, error) {
		return nil, errors.New("app: error")
	}, binder.Lazy())
	assert.NoError(t, err)

	var s Shape
	err = instance.Resolve(&s)
	assert.Error(t, err, "app: error")

	firstCall := true
	err = instance.Bind(func() (Database, error) {
		if firstCall {
			firstCall = false
			return &MySQL{}, nil
		}
		return nil, errors.New("app: second call error")
	}, binder.Lazy())
	assert.NoError(t, err)

	var db Database
	err = instance.Resolve(&db)
	assert.NoError(t, err)

	err = instance.Resolve(&db)
	assert.Error(t, err, "app: second call error")
}

func TestContainer_Transient_With_Resolve_With_Invalid_Signature_It_Should_Fail(t *testing.T) {
	err := instance.Bind(func() (Shape, Database, error) {
		return nil, nil, nil
	})
	assert.Error(t, err, "container: resolver function signature is invalid")
}

func TestContainer_TransientLazy_With_Resolve_With_Invalid_Signature_It_Should_Fail(t *testing.T) {
	err := instance.Bind(func() (Shape, Database, error) {
		return nil, nil, nil
	}, binder.Lazy())
	assert.Error(t, err, "container: resolver function signature is invalid")
}

func TestContainer_Transient_With_Resolve_That_Returns_NonError_Second_Value(t *testing.T) {
	instance.Reset()

	err := instance.Bind(func() (Shape, int) {
		return &Circle{a: 1}, 42
	})
	assert.Error(t, err, "container: resolver function signature is invalid")
}

func TestContainer_NamedTransient(t *testing.T) {
	err := instance.Bind(func() Shape {
		return &Circle{a: 13}
	}, binder.WithName("theCircle"))
	assert.NoError(t, err)

	var sh Shape
	err = instance.Resolve(&sh, resolver.WithName("theCircle"))
	assert.NoError(t, err)
	assert.Equal(t, sh.GetArea(), 13)
}

func TestContainer_NamedTransientLazy(t *testing.T) {
	err := instance.Bind(func() Shape {
		return &Circle{a: 13}
	}, binder.WithName("theCircle"), binder.Lazy())
	assert.NoError(t, err)

	var sh Shape
	err = instance.Resolve(&sh, resolver.WithName("theCircle"))
	assert.NoError(t, err)
	assert.Equal(t, sh.GetArea(), 13)
}

func TestContainer_Call_With_Multiple_Resolving(t *testing.T) {
	err := instance.Bind(func() Shape {
		return &Circle{a: 5}
	}, binder.Singleton())
	assert.NoError(t, err)

	err = instance.Bind(func() Database {
		return &MySQL{}
	}, binder.Singleton())
	assert.NoError(t, err)

	err = instance.Call(func(s Shape, m Database) {
		if _, ok := s.(*Circle); !ok {
			t.Error("Expected Circle")
		}

		if _, ok := m.(*MySQL); !ok {
			t.Error("Expected MySQL")
		}
	})
	assert.NoError(t, err)
}

func TestContainer_Call_With_Dependency_Missing_In_Chain(t *testing.T) {
	var instance = container.New()
	err := instance.Bind(func() (Database, error) {
		var s Shape
		if err := instance.Resolve(&s); err != nil {
			return nil, err
		}
		return &MySQL{}, nil
	}, binder.Singleton(), binder.Lazy())
	assert.NoError(t, err)

	err = instance.Call(func(m Database) {
		if _, ok := m.(*MySQL); !ok {
			t.Error("Expected MySQL")
		}
	})
	assert.EqualError(t, err, "container: no concrete found for the given abstraction; the abstraction is: container_test.Shape")
}

func TestContainer_Call_With_Unsupported_Receiver_It_Should_Fail(t *testing.T) {
	err := instance.Call("STRING!")
	assert.EqualError(t, err, "container: invalid function")
}

func TestContainer_Call_With_Second_UnBounded_Argument(t *testing.T) {
	instance.Reset()

	err := instance.Bind(func() Shape {
		return &Circle{}
	}, binder.Singleton())
	assert.NoError(t, err)

	err = instance.Call(func(s Shape, d Database) {})
	assert.EqualError(t, err, "container: no concrete found for the given abstraction; the abstraction is: container_test.Database")
}

func TestContainer_Call_With_A_Returning_Error(t *testing.T) {
	instance.Reset()

	err := instance.Bind(func() Shape {
		return &Circle{}
	}, binder.Singleton())
	assert.NoError(t, err)

	err = instance.Call(func(s Shape) error {
		return errors.New("app: some context error")
	})
	assert.EqualError(t, err, "app: some context error")
}

func TestContainer_Call_With_A_Returning_Nil_Error(t *testing.T) {
	instance.Reset()

	err := instance.Bind(func() Shape {
		return &Circle{}
	}, binder.Singleton())
	assert.NoError(t, err)

	err = instance.Call(func(s Shape) error {
		return nil
	})
	assert.Nil(t, err)
}

func TestContainer_Call_With_Invalid_Signature(t *testing.T) {
	instance.Reset()

	err := instance.Bind(func() Shape {
		return &Circle{}
	}, binder.Singleton())
	assert.NoError(t, err)

	err = instance.Call(func(s Shape) (int, error) {
		return 13, errors.New("app: some context error")
	})
	assert.EqualError(t, err, "container: receiver function signature is invalid")
}

func TestContainer_Resolve_With_Reference_As_Resolver(t *testing.T) {
	err := instance.Bind(func() Shape {
		return &Circle{a: 5}
	}, binder.Singleton())
	assert.NoError(t, err)

	err = instance.Bind(func() Database {
		return &MySQL{}
	}, binder.Singleton())
	assert.NoError(t, err)

	var (
		s Shape
		d Database
	)

	err = instance.Resolve(&s)
	assert.NoError(t, err)
	if _, ok := s.(*Circle); !ok {
		t.Error("Expected Circle")
	}

	err = instance.Resolve(&d)
	assert.NoError(t, err)
	if _, ok := d.(*MySQL); !ok {
		t.Error("Expected MySQL")
	}
}

func TestContainer_Resolve_With_Unsupported_Receiver_It_Should_Fail(t *testing.T) {
	err := instance.Resolve("STRING!")
	assert.EqualError(t, err, "container: invalid abstraction")
}

func TestContainer_Resolve_With_NonReference_Receiver_It_Should_Fail(t *testing.T) {
	var s Shape
	err := instance.Resolve(s)
	assert.EqualError(t, err, "container: invalid abstraction")
}

func TestContainer_Resolve_With_UnBounded_Reference_It_Should_Fail(t *testing.T) {
	instance.Reset()

	var s Shape
	err := instance.Resolve(&s)
	assert.EqualError(t, err, "container: no concrete found for the given abstraction; the abstraction is: container_test.Shape")
}

func TestContainer_Resolve_With_Runtime_Params(t *testing.T) {
	instance.Reset()

	err := instance.Bind(func(x int, s Shape) Database {
		return PostgreSQL{ready: x+s.GetArea() == 12}
	}, binder.Singleton(), binder.Lazy())
	assert.NoError(t, err)

	var db Database
	err = instance.Resolve(&db, resolver.WithParams(10, &Circle{a: 2}))
	assert.NoError(t, err)
	assert.True(t, db.Connect())
}

func TestContainer_Resolve_With_Runtime_Params_And_Container_Fallback(t *testing.T) {
	instance.Reset()

	err := instance.Bind(func() Shape {
		return &Circle{a: 2}
	}, binder.Singleton(), binder.Lazy())
	assert.NoError(t, err)

	err = instance.Bind(func(x int, s Shape) Database {
		return PostgreSQL{ready: x+s.GetArea() == 12}
	}, binder.Singleton(), binder.Lazy())
	assert.NoError(t, err)

	var db Database
	err = instance.Resolve(&db, resolver.WithParams(10))
	assert.NoError(t, err)
	assert.True(t, db.Connect())
}

func TestContainer_Resolve_With_Runtime_Params_Takes_Precedence_Over_Container(t *testing.T) {
	instance.Reset()

	// Container has Shape bound to Circle{a: 99} — runtime param must win.
	err := instance.Bind(func() Shape {
		return &Circle{a: 99}
	}, binder.Singleton(), binder.Lazy())
	assert.NoError(t, err)

	// Transient so the resolver is called fresh each time (no cached concrete).
	err = instance.Bind(func(s Shape) Database {
		return PostgreSQL{ready: s.GetArea() == 2}
	})
	assert.NoError(t, err)

	var db Database
	err = instance.Resolve(&db, resolver.WithParams(&Circle{a: 2}))
	assert.NoError(t, err)
	assert.True(t, db.Connect()) // would be false if container binding (a:99) was used
}

func TestContainer_Resolve_With_Runtime_Params_Missing_And_No_Fallback(t *testing.T) {
	instance.Reset()

	err := instance.Bind(func(x int, s Shape) Database {
		return PostgreSQL{ready: x+s.GetArea() == 12}
	}, binder.Singleton(), binder.Lazy())
	assert.NoError(t, err)

	var db Database
	err = instance.Resolve(&db, resolver.WithParams(10))
	assert.EqualError(t, err, "container: encountered error while making concrete for: container_test.Database. Error encountered: container: no concrete found for the given abstraction; the abstraction is: container_test.Shape")
}

func TestContainer_NamedResolve_With_Runtime_Params(t *testing.T) {
	instance.Reset()

	err := instance.Bind(func(x int, s Shape) Database {
		return PostgreSQL{ready: x+s.GetArea() == 12}
	}, binder.WithName("runtime"), binder.Lazy())
	assert.NoError(t, err)

	var db Database
	err = instance.Resolve(&db, resolver.WithName("runtime"), resolver.WithParams(10, &Circle{a: 2}))
	assert.NoError(t, err)
	assert.True(t, db.Connect())
}

func TestContainer_Fill_With_Struct_Pointer(t *testing.T) {
	err := instance.Bind(func() Shape {
		return &Circle{a: 5}
	}, binder.Singleton(), binder.Lazy())
	assert.NoError(t, err)

	err = instance.Bind(func() Shape {
		return &Circle{a: 5}
	}, binder.WithName("C"), binder.Singleton(), binder.Lazy())
	assert.NoError(t, err)

	err = instance.Bind(func() Database {
		return &MySQL{}
	}, binder.Singleton(), binder.Lazy())
	assert.NoError(t, err)

	myApp := struct {
		S Shape    `container:"type"`
		D Database `container:"type"`
		C Shape    `container:"name"`
		X string
	}{}

	err = instance.Fill(&myApp)
	assert.NoError(t, err)

	assert.IsType(t, &Circle{}, myApp.S)
	assert.IsType(t, &MySQL{}, myApp.D)
}

func TestContainer_Fill_Unexported_With_Struct_Pointer(t *testing.T) {
	err := instance.Bind(func() Shape {
		return &Circle{a: 5}
	}, binder.Singleton(), binder.Lazy())
	assert.NoError(t, err)

	err = instance.Bind(func() Database {
		return &MySQL{}
	}, binder.Singleton(), binder.Lazy())
	assert.NoError(t, err)

	myApp := struct {
		s Shape    `container:"type"`
		d Database `container:"type"`
		y int
	}{}

	err = instance.Fill(&myApp)
	assert.NoError(t, err)

	assert.IsType(t, &Circle{}, myApp.s)
	assert.IsType(t, &MySQL{}, myApp.d)
}

func TestContainer_Fill_With_Invalid_Field_It_Should_Fail(t *testing.T) {
	err := instance.Bind(func() Shape {
		return &Circle{a: 5}
	}, binder.WithName("C"), binder.Singleton(), binder.Lazy())
	assert.NoError(t, err)

	type App struct {
		S string `container:"name"`
	}

	myApp := App{}

	err = instance.Fill(&myApp)
	assert.EqualError(t, err, "container: cannot make field; the field is: S")
}

func TestContainer_Fill_With_Invalid_Tag_It_Should_Fail(t *testing.T) {
	type App struct {
		S string `container:"invalid"`
	}

	myApp := App{}

	err := instance.Fill(&myApp)
	assert.EqualError(t, err, "container: invalid struct tag; the field is: S")
}

func TestContainer_Fill_With_Invalid_Field_Name_It_Should_Fail(t *testing.T) {
	type App struct {
		S string `container:"name"`
	}

	myApp := App{}

	err := instance.Fill(&myApp)
	assert.EqualError(t, err, "container: cannot make field; the field is: S")
}

func TestContainer_Fill_With_Invalid_Struct_It_Should_Fail(t *testing.T) {
	invalidStruct := 0
	err := instance.Fill(&invalidStruct)
	assert.EqualError(t, err, "container: invalid structure")
}

func TestContainer_Fill_With_Invalid_Pointer_It_Should_Fail(t *testing.T) {
	var s Shape
	err := instance.Fill(s)
	assert.EqualError(t, err, "container: invalid structure")
}

func TestContainer_Fill_With_Dependency_Missing_In_Chain(t *testing.T) {
	var instance = container.New()

	err := instance.Bind(func() Shape {
		return &Circle{a: 5}
	}, binder.Singleton())
	assert.NoError(t, err)

	err = instance.Bind(func() (Shape, error) {
		var s Shape
		if err := instance.Resolve(&s, resolver.WithName("foo")); err != nil {
			return nil, err
		}
		return &Circle{a: 5}, nil
	}, binder.WithName("C"), binder.Singleton(), binder.Lazy())
	assert.NoError(t, err)

	err = instance.Bind(func() Database {
		return &MySQL{}
	}, binder.Singleton())
	assert.NoError(t, err)

	myApp := struct {
		S Shape    `container:"type"`
		D Database `container:"type"`
		C Shape    `container:"name"`
		X string
	}{}

	err = instance.Fill(&myApp)
	assert.EqualError(t, err, "container: no concrete found for the given abstraction; the abstraction is: container_test.Shape")
}

func TestContainer_Fill_With_Runtime_Params(t *testing.T) {
	instance.Reset()

	err := instance.Bind(func(x int, s Shape) Database {
		return PostgreSQL{ready: x+s.GetArea() == 12}
	}, binder.Singleton(), binder.Lazy())
	assert.NoError(t, err)

	myApp := struct {
		D Database `container:"type"`
	}{}

	err = instance.Fill(&myApp, resolver.ResolveParams(10, &Circle{a: 2}))
	assert.NoError(t, err)
	assert.True(t, myApp.D.Connect())
}

func TestContainer_Fill_With_Runtime_Params_And_Container_Fallback(t *testing.T) {
	instance.Reset()

	err := instance.Bind(func() Shape {
		return &Circle{a: 2}
	}, binder.Singleton(), binder.Lazy())
	assert.NoError(t, err)

	err = instance.Bind(func(x int, s Shape) Database {
		return PostgreSQL{ready: x+s.GetArea() == 12}
	}, binder.Singleton(), binder.Lazy())
	assert.NoError(t, err)

	myApp := struct {
		D Database `container:"type"`
	}{}

	err = instance.Fill(&myApp, resolver.ResolveParams(10))
	assert.NoError(t, err)
	assert.True(t, myApp.D.Connect())
}

func TestContainer_Fill_With_Runtime_Params_Missing_And_No_Fallback(t *testing.T) {
	instance.Reset()

	err := instance.Bind(func(x int, s Shape) Database {
		return PostgreSQL{ready: (x + s.GetArea()) == 12}
	}, binder.Singleton(), binder.Lazy())
	assert.NoError(t, err)

	myApp := struct {
		D Database `container:"type"`
	}{}

	err = instance.Fill(&myApp, resolver.ResolveParams(10))
	assert.EqualError(t, err, "container: no concrete found for the given abstraction; the abstraction is: container_test.Shape")
}

func TestContainer_Singleton_Bind_As_Concrete_But_Resolve_By_Interface(t *testing.T) {
	instance := container.New()

	err := instance.Bind(func() *Circle {
		return &Circle{a: 13}
	}, binder.Singleton())
	assert.NoError(t, err)

	err = instance.Call(func(s Shape) {
		a := s.GetArea()
		assert.Equal(t, 13, a)
		s.SetArea(666)
	})
	assert.NoError(t, err)

	err = instance.Call(func(s ReadOnlyShape) {
		a := s.GetArea()
		assert.Equal(t, 666, a)
	})
	assert.NoError(t, err)
}

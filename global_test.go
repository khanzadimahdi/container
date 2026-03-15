package container_test

import (
	"testing"

	"github.com/golobby/container/v3"
	"github.com/golobby/container/v3/bind"
	"github.com/golobby/container/v3/resolve"
	"github.com/stretchr/testify/assert"
)

func TestGlobalBindModes(t *testing.T) {
	tests := []struct {
		name        string
		bindOpts    []bind.BindOption
		resolveOpts []resolve.ResolveOption
	}{
		{name: "singleton", bindOpts: []bind.BindOption{bind.Singleton()}},
		{name: "singleton lazy", bindOpts: []bind.BindOption{bind.Singleton(), bind.Lazy()}},
		{name: "named singleton", bindOpts: []bind.BindOption{bind.WithName("rounded"), bind.Singleton()}, resolveOpts: []resolve.ResolveOption{resolve.WithName("rounded")}},
		{name: "named singleton lazy", bindOpts: []bind.BindOption{bind.WithName("rounded"), bind.Singleton(), bind.Lazy()}, resolveOpts: []resolve.ResolveOption{resolve.WithName("rounded")}},
		{name: "transient"},
		{name: "transient lazy", bindOpts: []bind.BindOption{bind.Lazy()}},
		{name: "named transient", bindOpts: []bind.BindOption{bind.WithName("rounded")}, resolveOpts: []resolve.ResolveOption{resolve.WithName("rounded")}},
		{name: "named transient lazy", bindOpts: []bind.BindOption{bind.WithName("rounded"), bind.Lazy()}, resolveOpts: []resolve.ResolveOption{resolve.WithName("rounded")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container.Reset()

			err := container.Bind(func() Shape {
				return &Circle{a: 13}
			}, tt.bindOpts...)
			assert.NoError(t, err)

			var s Shape
			err = container.Resolve(&s, tt.resolveOpts...)
			assert.NoError(t, err)
			assert.Equal(t, 13, s.GetArea())
		})
	}
}

func TestCall(t *testing.T) {
	container.Reset()

	err := container.Call(func() {})
	assert.NoError(t, err)
}

func TestResolve(t *testing.T) {
	container.Reset()

	var s Shape

	err := container.Bind(func() Shape {
		return &Circle{a: 13}
	}, bind.Singleton())
	assert.NoError(t, err)

	err = container.Resolve(&s)
	assert.NoError(t, err)
}

func TestNamedResolve(t *testing.T) {
	container.Reset()

	var s Shape

	err := container.Bind(func() Shape {
		return &Circle{a: 13}
	}, bind.WithName("rounded"), bind.Singleton())
	assert.NoError(t, err)

	err = container.Resolve(&s, resolve.WithName("rounded"))
	assert.NoError(t, err)
}

func TestResolve_With_Runtime_Params(t *testing.T) {
	container.Reset()

	err := container.Bind(func(x int, s Shape) Database {
		return PostgreSQL{ready: x+s.GetArea() == 12}
	}, bind.Singleton(), bind.Lazy())
	assert.NoError(t, err)

	var db Database
	err = container.Resolve(&db, resolve.WithParams(10, &Circle{a: 2}))
	assert.NoError(t, err)
	assert.True(t, db.Connect())
}

func TestNamedResolve_With_Runtime_Params(t *testing.T) {
	container.Reset()

	err := container.Bind(func(x int, s Shape) Database {
		return PostgreSQL{ready: x+s.GetArea() == 12}
	}, bind.WithName("rounded"), bind.Lazy())
	assert.NoError(t, err)

	var db Database
	err = container.Resolve(&db, resolve.WithName("rounded"), resolve.WithParams(10, &Circle{a: 2}))
	assert.NoError(t, err)
	assert.True(t, db.Connect())
}

func TestResolve_With_Runtime_Params_And_Container_Fallback(t *testing.T) {
	container.Reset()

	err := container.Bind(func() Shape {
		return &Circle{a: 2}
	}, bind.Singleton())
	assert.NoError(t, err)

	err = container.Bind(func(x int, s Shape) Database {
		return PostgreSQL{ready: x+s.GetArea() == 12}
	}, bind.Singleton(), bind.Lazy())
	assert.NoError(t, err)

	var db Database
	err = container.Resolve(&db, resolve.WithParams(10))
	assert.NoError(t, err)
	assert.True(t, db.Connect())
}

func TestResolve_With_Runtime_Params_Takes_Precedence_Over_Container(t *testing.T) {
	container.Reset()

	err := container.Bind(func() Shape {
		return &Circle{a: 99}
	}, bind.Singleton())
	assert.NoError(t, err)

	err = container.Bind(func(s Shape) Database {
		return PostgreSQL{ready: s.GetArea() == 2}
	})
	assert.NoError(t, err)

	var db Database
	err = container.Resolve(&db, resolve.WithParams(&Circle{a: 2}))
	assert.NoError(t, err)
	assert.True(t, db.Connect())
}

func TestResolve_With_Runtime_Params_Missing_And_No_Fallback(t *testing.T) {
	container.Reset()

	err := container.Bind(func(x int, s Shape) Database {
		return PostgreSQL{ready: x+s.GetArea() == 12}
	}, bind.Singleton(), bind.Lazy())
	assert.NoError(t, err)

	var db Database
	err = container.Resolve(&db, resolve.WithParams(10))
	assert.EqualError(t, err, "container: encountered error while making concrete for: container_test.Database. Error encountered: container: no concrete found for the given abstraction; the abstraction is: container_test.Shape")
}

func TestFill(t *testing.T) {
	container.Reset()

	err := container.Bind(func() Shape {
		return &Circle{a: 13}
	}, bind.Singleton())
	assert.NoError(t, err)

	myApp := struct {
		S Shape `container:"type"`
	}{}

	err = container.Fill(&myApp)
	assert.NoError(t, err)
	assert.IsType(t, &Circle{}, myApp.S)
}

func TestFill_With_Runtime_Params(t *testing.T) {
	container.Reset()

	err := container.Bind(func(x int, s Shape) Database {
		return PostgreSQL{ready: x+s.GetArea() == 12}
	}, bind.Singleton(), bind.Lazy())
	assert.NoError(t, err)

	myApp := struct {
		D Database `container:"type"`
	}{}

	err = container.Fill(&myApp, resolve.WithParams(10, &Circle{a: 2}))
	assert.NoError(t, err)
	assert.True(t, myApp.D.Connect())
}

func TestFill_With_Runtime_Params_And_Container_Fallback(t *testing.T) {
	container.Reset()

	err := container.Bind(func() Shape {
		return &Circle{a: 2}
	}, bind.Singleton())
	assert.NoError(t, err)

	err = container.Bind(func(x int, s Shape) Database {
		return PostgreSQL{ready: x+s.GetArea() == 12}
	}, bind.Singleton(), bind.Lazy())
	assert.NoError(t, err)

	myApp := struct {
		D Database `container:"type"`
	}{}

	err = container.Fill(&myApp, resolve.WithParams(10))
	assert.NoError(t, err)
	assert.True(t, myApp.D.Connect())
}

func TestFill_With_Runtime_Params_Missing_And_No_Fallback(t *testing.T) {
	container.Reset()

	err := container.Bind(func(x int, s Shape) Database {
		return PostgreSQL{ready: x+s.GetArea() == 12}
	}, bind.Singleton(), bind.Lazy())
	assert.NoError(t, err)

	myApp := struct {
		D Database `container:"type"`
	}{}

	err = container.Fill(&myApp, resolve.WithParams(10))
	assert.EqualError(t, err, "container: no concrete found for the given abstraction; the abstraction is: container_test.Shape")
}

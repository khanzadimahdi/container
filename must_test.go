package container_test

import (
	"testing"

	"github.com/golobby/container/v3"
)

func TestMustCall_It_Should_Panic_On_Error(t *testing.T) {
	c := container.New()

	defer func() { recover() }()
	container.MustCall(c, func(s Shape) {
		s.GetArea()
	})
	t.Errorf("panic expcted.")
}

func TestMustResolve_It_Should_Panic_On_Error(t *testing.T) {
	c := container.New()

	var s Shape

	defer func() { recover() }()
	container.MustResolve(c, &s)
	t.Errorf("panic expcted.")
}

func TestMustNamedResolve_It_Should_Panic_On_Error(t *testing.T) {
	c := container.New()

	var s Shape

	defer func() { recover() }()
	container.MustNamedResolve(c, &s, "name")
	t.Errorf("panic expcted.")
}

func TestMustFill_It_Should_Panic_On_Error(t *testing.T) {
	c := container.New()

	myApp := struct {
		S Shape `container:"type"`
	}{}

	defer func() { recover() }()
	container.MustFill(c, &myApp)
	t.Errorf("panic expcted.")
}

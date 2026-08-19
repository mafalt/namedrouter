package namedrouter_test

import (
	"errors"
	"testing"

	"github.com/mafalt/namedrouter"
)

func TestRegistry_RegisterAndGetRoute(t *testing.T) {
	registry := namedrouter.NewRegistry(&testParser{}, &testApplier{})

	route := namedrouter.RouteDefinition{
		Name:    "users.show",
		Pattern: "/users/{id}",
	}

	if err := registry.Register(route); err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}

	got, ok := registry.GetRoute("users.show")
	if !ok {
		t.Fatal("GetRoute() expected route to be present")
	}
	if got.Name != route.Name || got.Pattern != route.Pattern {
		t.Fatalf("GetRoute() returned unexpected route: %+v", got)
	}
}

func TestRegistry_MustRegisterPanicsOnInvalidRoute(t *testing.T) {
	registry := namedrouter.NewRegistry(&testParser{}, &testApplier{})

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("MustRegister() expected panic")
		}
	}()

	registry.MustRegister(namedrouter.RouteDefinition{Name: ""})
}

func TestRegistry_URL(t *testing.T) {
	registry := namedrouter.NewRegistry(&testParser{}, &testApplier{})

	route := namedrouter.RouteDefinition{
		Name:    "users.show",
		Pattern: "/users/{id}",
	}

	if err := registry.Register(route); err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}

	url, err := registry.URL("users.show", namedrouter.RouteParams{"id": 123})
	if err != nil {
		t.Fatalf("URL() unexpected error: %v", err)
	}
	if url != "/users/123" {
		t.Fatalf("URL() expected '/users/123', got %q", url)
	}
}

func TestRegistry_URLRouteNotFound(t *testing.T) {
	registry := namedrouter.NewRegistry(&testParser{}, &testApplier{})

	_, err := registry.URL("missing", nil)
	if err == nil {
		t.Fatal("URL() expected error for unknown route")
	}

	var routeErr *namedrouter.RouteNotFoundError
	if !errors.As(err, &routeErr) {
		t.Fatalf("URL() expected RouteNotFoundError, got %T", err)
	}
}

func TestRegistry_RejectsEmptyRouteName(t *testing.T) {
	registry := namedrouter.NewRegistry(&testParser{}, &testApplier{})

	err := registry.Register(namedrouter.RouteDefinition{Name: ""})
	if !errors.Is(err, namedrouter.ErrRouteNameEmpty) {
		t.Fatalf("Register() expected ErrRouteNameEmpty, got %v", err)
	}
}

func TestRegistry_RejectsDuplicateRouteName(t *testing.T) {
	registry := namedrouter.NewRegistry(&testParser{}, &testApplier{})

	first := namedrouter.RouteDefinition{Name: "users.show", Pattern: "/users/{id}"}
	second := namedrouter.RouteDefinition{Name: "users.show", Pattern: "/users/other"}

	if err := registry.Register(first); err != nil {
		t.Fatalf("Register(first) unexpected error: %v", err)
	}

	err := registry.Register(second)
	if err == nil {
		t.Fatal("Register(second) expected duplicate route error")
	}

	var duplicateErr *namedrouter.DuplicateRouteParameterError
	if !errors.As(err, &duplicateErr) {
		t.Fatalf("Register(second) expected DuplicateRouteParameterError, got %T", err)
	}
}

func TestRegistry_RejectsMissingRouteParameters(t *testing.T) {
	registry := namedrouter.NewRegistry(&testParser{}, &testApplier{})

	route := namedrouter.RouteDefinition{Name: "users.show", Pattern: "/users/{id}"}
	if err := registry.Register(route); err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}

	_, err := registry.URL("users.show", nil)
	if err == nil {
		t.Fatal("URL() expected missing parameter error")
	}

	var missingErr *namedrouter.RouteParameterNotProvidedError
	if !errors.As(err, &missingErr) {
		t.Fatalf("URL() expected RouteParameterNotProvidedError, got %T", err)
	}
}

func TestRegistry_RejectsExtraRouteParameters(t *testing.T) {
	registry := namedrouter.NewRegistry(&testParser{}, &testApplier{})

	route := namedrouter.RouteDefinition{Name: "users.show", Pattern: "/users"}
	if err := registry.Register(route); err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}

	_, err := registry.URL("users.show", namedrouter.RouteParams{"extra": "param"})
	if err == nil {
		t.Fatal("URL() expected extra parameter error")
	}

	var extraErr *namedrouter.RouteParameterDoesNotExistError
	if !errors.As(err, &extraErr) {
		t.Fatalf("URL() expected RouteParameterDoesNotExistError, got %T", err)
	}
}

func TestRegistry_RejectsNilRouteParameter(t *testing.T) {
	registry := namedrouter.NewRegistry(&testParser{}, &testApplier{})

	route := namedrouter.RouteDefinition{Name: "users.show", Pattern: "/users/{id}"}
	if err := registry.Register(route); err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}

	_, err := registry.URL("users.show", namedrouter.RouteParams{"id": nil})
	if err == nil {
		t.Fatal("URL() expected nil parameter error")
	}

	var nilErr *namedrouter.NilRouteParameterError
	if !errors.As(err, &nilErr) {
		t.Fatalf("URL() expected NilRouteParameterError, got %T", err)
	}
}

func TestRegistry_RejectsInvalidParameterFormat(t *testing.T) {
	registry := namedrouter.NewRegistry(&testParser{}, &testApplier{})

	route := namedrouter.RouteDefinition{Name: "users.show", Pattern: "/users/{"}
	err := registry.Register(route)
	if err == nil {
		t.Fatal("Register() expected invalid parameter format error")
	}

	var invalidErr *namedrouter.InvalidRouteParameterFormatError
	if !errors.As(err, &invalidErr) {
		t.Fatalf("Register() expected InvalidRouteParameterFormatError, got %T", err)
	}
}

func TestRegistry_RejectsEmptyParameterName(t *testing.T) {
	registry := namedrouter.NewRegistry(&testParser{}, &testApplier{})

	route := namedrouter.RouteDefinition{Name: "users.show", Pattern: "/users/{}"}
	err := registry.Register(route)
	if err == nil {
		t.Fatal("Register() expected empty parameter name error")
	}

	var emptyErr *namedrouter.EmptyRouteParameterNameError
	if !errors.As(err, &emptyErr) {
		t.Fatalf("Register() expected EmptyRouteParameterNameError, got %T", err)
	}
}

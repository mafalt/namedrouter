package namedrouter_test

import (
	"testing"

	"github.com/mafalt/namedrouter"
)

func TestRouteAlreadyExistsError(t *testing.T) {
	err := &namedrouter.RouteAlreadyExistsError{
		RouteName: "test.route",
	}

	expected := "route already exists: test.route"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestRouteNotFoundError(t *testing.T) {
	err := &namedrouter.RouteNotFoundError{
		RouteName: "missing.route",
	}

	expected := "route not found: missing.route"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestRouteParameterDoesNotExistError(t *testing.T) {
	err := &namedrouter.RouteParameterDoesNotExistError{
		RouteName:     "test.route",
		ParameterName: "id",
	}

	expected := "route parameter does not exist: id for route: test.route"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestRouteParameterNotProvidedError(t *testing.T) {
	err := &namedrouter.RouteParameterNotProvidedError{
		RouteName:     "users.show",
		ParameterName: "userId",
	}

	expected := "route parameter not provided: userId for route: users.show"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestDuplicateRouteParameterError(t *testing.T) {
	err := &namedrouter.DuplicateRouteParameterError{
		ParameterName: "duplicate.route",
	}

	expected := "duplicate route parameter: duplicate.route"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestEmptyRouteParameterNameError(t *testing.T) {
	err := &namedrouter.EmptyRouteParameterNameError{
		RouteName: "test.route",
	}

	expected := "empty route parameter name for route: test.route"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestInvalidRouteParameterFormatError(t *testing.T) {
	err := &namedrouter.InvalidRouteParameterFormatError{
		RouteName: "bad.route",
	}

	expected := "invalid route parameter format for route: bad.route"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestNilRouteParameterError(t *testing.T) {
	err := &namedrouter.NilRouteParameterError{
		RouteName: "test.route",
		ParamName: "id",
	}

	expected := "nil route parameters for route: test.route, parameter: id"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestErrRouteNameEmpty(t *testing.T) {
	expected := "route name cannot be empty"
	if namedrouter.ErrRouteNameEmpty.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, namedrouter.ErrRouteNameEmpty.Error())
	}
}

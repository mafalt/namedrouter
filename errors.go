package namedrouter

import "errors"

// RouteAlreadyExistsError is an error type that indicates a route with the same name already exists in the registry.
type RouteAlreadyExistsError struct {
	RouteName string
}

// Error implements the error interface for RouteAlreadyExistsError.
func (e *RouteAlreadyExistsError) Error() string {
	return "route already exists: " + e.RouteName
}

// RouteNotFoundError is an error type that indicates a route with the specified name was not found in the registry.
type RouteNotFoundError struct {
	RouteName string
}

// Error implements the error interface for RouteNotFoundError.
func (e *RouteNotFoundError) Error() string {
	return "route not found: " + e.RouteName
}

// RouteParameterDoesNotExistError is an error type that indicates a route parameter does not exist for the specified route.
type RouteParameterDoesNotExistError struct {
	RouteName     string
	ParameterName string
}

// Error implements the error interface for RouteParameterDoesNotExistError.
func (e *RouteParameterDoesNotExistError) Error() string {
	return "route parameter does not exist: " + e.ParameterName + " for route: " + e.RouteName
}

// RouteParameterNotProvidedError is an error type that indicates a required route parameter was not provided for the specified route.
type RouteParameterNotProvidedError struct {
	RouteName     string
	ParameterName string
}

// Error implements the error interface for RouteParameterNotProvidedError.
func (e *RouteParameterNotProvidedError) Error() string {
	return "route parameter not provided: " + e.ParameterName + " for route: " + e.RouteName
}

// DuplicateRouteParameterError is an error type that indicates a duplicate route parameter was found for the specified route.
type DuplicateRouteParameterError struct {
	ParameterName string
}

// Error implements the error interface for DuplicateRouteParameterError.
func (e *DuplicateRouteParameterError) Error() string {
	return "duplicate route parameter: " + e.ParameterName
}

// EmptyRouteParameterNameError is an error type that indicates an empty route parameter name was found for the specified route.
type EmptyRouteParameterNameError struct {
	RouteName string
}

// Error implements the error interface for EmptyRouteParameterNameError.
func (e *EmptyRouteParameterNameError) Error() string {
	return "empty route parameter name for route: " + e.RouteName
}

// InvalidRouteParameterFormatError is an error type that indicates an invalid route parameter format was found for the specified route.
type InvalidRouteParameterFormatError struct {
	RouteName string
}

// Error implements the error interface for InvalidRouteParameterFormatError.
func (e *InvalidRouteParameterFormatError) Error() string {
	return "invalid route parameter format for route: " + e.RouteName
}

// NilRouteParameterError is an error type that indicates nil route parameters were provided for the specified route.
type NilRouteParameterError struct {
	RouteName string
	ParamName string
}

// Error implements the error interface for NilRouteParameterError.
func (e *NilRouteParameterError) Error() string {
	return "nil route parameters for route: " + e.RouteName + ", parameter: " + e.ParamName
}

var ErrRouteNameEmpty = errors.New("route name cannot be empty")

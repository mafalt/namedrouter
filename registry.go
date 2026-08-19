package namedrouter

// RouteParams represents a single route parameter with its name and value.
type RouteParams map[string]any

// Contains checks if the given parameter name exists in the RouteParams map.
func (r RouteParams) Contains(name string) bool {
	_, exists := r[name]
	return exists
}

// ParameterParser is an interface that defines methods to parse parameters from incoming RouteDefinition.
type ParameterParser interface {
	Parse(route RouteDefinition) (RouteParameterNames, error)
}

// ParameterApplier is an interface that defines methods to apply parameter values to specified RouteDefinition.
type ParameterApplier interface {
	Apply(pattern string, params RouteParams) string
}

// RoutesRegistry is an interface that defines methods for managing route definitions in a registry.
type RoutesRegistry interface {
	GetRoute(name string) (RouteDefinition, bool)
	Register(route RouteDefinition) error
	MustRegister(route RouteDefinition)
	URL(name string, params RouteParams) (string, error)
}

// routesRegistry is a struct that holds a map of route definitions,
// allowing for the retrieval of route definitions by their names.
type routesRegistry struct {
	routes        map[string]RouteDefinition
	paramsParser  ParameterParser
	paramsApplier ParameterApplier
}

func NewRegistry(parser ParameterParser, applier ParameterApplier) RoutesRegistry {
	return &routesRegistry{
		routes:        make(map[string]RouteDefinition),
		paramsParser:  parser,
		paramsApplier: applier,
	}
}

// GetRoute retrieves a route definition by its name from the route registry.
func (r *routesRegistry) GetRoute(name string) (RouteDefinition, bool) {
	route, ok := r.routes[name]
	return route, ok
}

// Register registers a new route definition to the route registry.
func (r *routesRegistry) Register(route RouteDefinition) error {
	if err := r.validateRoute(route); err != nil {
		return err
	}

	routeParams, err := r.extractParametersFromPattern(route)
	if err != nil {
		return err
	}
	route.Parameters = routeParams

	r.routes[route.Name] = route
	return nil
}

// MustRegister registers a new route definition to the route registry. It panics if an error occurs.
func (r *routesRegistry) MustRegister(route RouteDefinition) {
	if err := r.Register(route); err != nil {
		panic(err)
	}
}

// URL generates a URL for a named route, replacing any route parameters with the provided values.
func (r *routesRegistry) URL(name string, params RouteParams) (string, error) {
	route, ok := r.GetRoute(name)
	if !ok {
		return "", &RouteNotFoundError{RouteName: name}
	}

	if err := r.validateRouteParameters(route, params); err != nil {
		return "", err
	}

	return r.replaceRouteParameters(route.Pattern, params), nil
}

// replaceRouteParameters replaces route parameters in the given pattern with the provided values from the params map.
func (r *routesRegistry) replaceRouteParameters(pattern string, params RouteParams) string {
	return r.paramsApplier.Apply(pattern, params)
}

// validateRouteParameters checks if the provided route parameters match the expected parameters for the specified route definition.
func (r *routesRegistry) validateRouteParameters(route RouteDefinition, params RouteParams) error {
	if len(route.Parameters) == 0 && len(params) > 0 {
		return &RouteParameterDoesNotExistError{
			RouteName:     route.Name,
			ParameterName: "unknown",
		}
	}

	// Check if all required parameters are provided
	for paramName := range route.Parameters {
		if !params.Contains(paramName) {
			return &RouteParameterNotProvidedError{
				RouteName:     route.Name,
				ParameterName: paramName,
			}
		}
	}

	for paramName := range params {
		// Check if the parameter exists in the route definition
		if !route.Parameters.Contains(paramName) {
			return &RouteParameterDoesNotExistError{
				RouteName:     route.Name,
				ParameterName: paramName,
			}
		}

		// Check if the parameter value is nil
		if params[paramName] == nil {
			return &NilRouteParameterError{
				RouteName: route.Name,
				ParamName: paramName,
			}
		}
	}

	return nil
}

// validateRoute checks if the route definition is valid before registering it in the route registry.
func (r *routesRegistry) validateRoute(route RouteDefinition) error {
	if err := r.validateName(route.Name); err != nil {
		return err
	}

	return nil
}

// validateName checks if the route name is valid and not already registered in the route registry.
func (r *routesRegistry) validateName(name string) error {
	if name == "" {
		return ErrRouteNameEmpty
	}

	if _, exists := r.routes[name]; exists {
		return &RouteAlreadyExistsError{
			RouteName: name,
		}
	}

	return nil
}

// extractParametersFromPattern extracts route parameters from the route pattern and returns them as a RouteParams map.
func (r *routesRegistry) extractParametersFromPattern(route RouteDefinition) (RouteParameterNames, error) {
	return r.paramsParser.Parse(route)
}

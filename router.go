package namedrouter

import (
	"fmt"
	"net/http"
)

// Middleware represents a function that takes an http.Handler and returns a new http.Handler.
type Middleware = func(http.Handler) http.Handler

// Adapter is an interface that defines methods that every Adapter implementation need to implement.
type Adapter interface {
	http.Handler

	Subrouter(prefix string, middlewares ...Middleware) Adapter

	Register(method, pattern string, handler http.HandlerFunc)
	ApplyMiddlewares(pattern string, middlewares ...Middleware) Adapter
	JoinPath(prefix, pattern string) string
	Use(middlewares ...Middleware)
	Static(pattern, root string)

	URLParam(r *http.Request, key string) string

	ParameterParser() ParameterParser
	ParameterApplier() ParameterApplier

	Walk()
}

// NamedRouter is an interface that defines methods
// for registering named routes with different HTTP methods (GET, POST, PUT, DELETE)
// and grouping routes under a common prefix.
// It allows for the organization and management of routes in a web application.
type NamedRouter interface {
	http.Handler

	Subrouter(prefix string, middlewares ...Middleware) NamedRouter

	Use(middlewares ...Middleware)
	Static(pattern, root string)

	RegisterGet(pattern, name string, handler http.HandlerFunc, middlewares ...Middleware)
	RegisterPost(pattern, name string, handler http.HandlerFunc, middlewares ...Middleware)
	RegisterPut(pattern, name string, handler http.HandlerFunc, middlewares ...Middleware)
	RegisterDelete(pattern, name string, handler http.HandlerFunc, middlewares ...Middleware)
	RegisterPatch(pattern, name string, handler http.HandlerFunc, middlewares ...Middleware)
	RegisterTrace(pattern, name string, handler http.HandlerFunc, middlewares ...Middleware)
	RegisterHead(pattern, name string, handler http.HandlerFunc, middlewares ...Middleware)
	RegisterOptions(pattern, name string, handler http.HandlerFunc, middlewares ...Middleware)
	Register(method, pattern, name string, handler http.HandlerFunc, middlewares ...Middleware)

	URL(name string, params RouteParams) (string, error)
	MustURL(name string, params RouteParams) string

	URLParam(r *http.Request, key string) string

	Walk()
}

// RouteMethod represents the HTTP method for a route.
type RouteMethod string

type RouteParameterNames map[string]struct{}

// Contains checks if the given parameter name exists in the RouteParameterNames map.
func (r RouteParameterNames) Contains(name string) bool {
	_, exists := r[name]
	return exists
}

const (
	MethodGET     RouteMethod = "GET"
	MethodPOST    RouteMethod = "POST"
	MethodPUT     RouteMethod = "PUT"
	MethodDELETE  RouteMethod = "DELETE"
	MethodPATCH   RouteMethod = "PATCH"
	MethodTRACE   RouteMethod = "TRACE"
	MethodHEAD    RouteMethod = "HEAD"
	MethodOPTIONS RouteMethod = "OPTIONS"
)

// RouteDefinition represents a single route definition with its name, pattern, HTTP method, handler,
// and optional middlewares.
type RouteDefinition struct {
	Name        string
	Pattern     string
	Method      string
	Handler     http.HandlerFunc
	Middlewares []Middleware
	Parameters  RouteParameterNames
}

// namedRouter is a concrete implementation of the NamedRouter interface.
type namedRouter struct {
	adapter  Adapter
	registry RoutesRegistry
	prefix   string
}

// New creates a new instance of NamedRouter with the provided adapter,
// parameter parser and parameter applier.
func New(adapter Adapter) NamedRouter {
	return &namedRouter{
		adapter:  adapter,
		registry: NewRegistry(adapter.ParameterParser(), adapter.ParameterApplier()),
	}
}

// ServeHTTP implements the http.Handler interface for the namedRouter.
func (n *namedRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	n.adapter.ServeHTTP(w, r)
}

// Subrouter creates a new sub router with the specified prefix and optional middlewares.
func (n *namedRouter) Subrouter(prefix string, middlewares ...Middleware) NamedRouter {
	return &namedRouter{
		adapter:  n.adapter.Subrouter(prefix, middlewares...),
		registry: n.registry,
		prefix:   n.adapter.JoinPath(n.prefix, prefix),
	}
}

// RegisterGet registers a new GET route with the specified pattern, name, handler, and optional middlewares.
func (n *namedRouter) RegisterGet(pattern, name string, handler http.HandlerFunc, middlewares ...Middleware) {
	n.Register(string(MethodGET), pattern, name, handler, middlewares...)
}

// RegisterPost registers a new POST route with the specified pattern, name, handler, and optional middlewares.
func (n *namedRouter) RegisterPost(pattern, name string, handler http.HandlerFunc, middlewares ...Middleware) {
	n.Register(string(MethodPOST), pattern, name, handler, middlewares...)
}

// RegisterPut registers a new PUT route with the specified pattern, name, handler, and optional middlewares.
func (n *namedRouter) RegisterPut(pattern, name string, handler http.HandlerFunc, middlewares ...Middleware) {
	n.Register(string(MethodPUT), pattern, name, handler, middlewares...)
}

// RegisterDelete registers a new DELETE route with the specified pattern, name, handler, and optional middlewares.
func (n *namedRouter) RegisterDelete(pattern, name string, handler http.HandlerFunc, middlewares ...Middleware) {
	n.Register(string(MethodDELETE), pattern, name, handler, middlewares...)
}

// RegisterPatch registers a new PATCH route with the specified pattern, name, handler, and optional middlewares.
func (n *namedRouter) RegisterPatch(pattern, name string, handler http.HandlerFunc, middlewares ...Middleware) {
	n.Register(string(MethodPATCH), pattern, name, handler, middlewares...)
}

// RegisterTrace registers a new TRACE route with the specified pattern, name, handler, and optional middlewares.
func (n *namedRouter) RegisterTrace(pattern, name string, handler http.HandlerFunc, middlewares ...Middleware) {
	n.Register(string(MethodTRACE), pattern, name, handler, middlewares...)
}

// RegisterHead registers a new HEAD route with the specified pattern, name, handler, and optional middlewares.
func (n *namedRouter) RegisterHead(pattern, name string, handler http.HandlerFunc, middlewares ...Middleware) {
	n.Register(string(MethodHEAD), pattern, name, handler, middlewares...)
}

// RegisterOptions registers a new OPTIONS route with the specified pattern, name, handler, and optional middlewares.
func (n *namedRouter) RegisterOptions(pattern, name string, handler http.HandlerFunc, middlewares ...Middleware) {
	n.Register(string(MethodOPTIONS), pattern, name, handler, middlewares...)
}

// Register registers a new route with a custom HTTP method, specified pattern, name, handler, and optional middlewares.
func (n *namedRouter) Register(method, pattern, name string, handler http.HandlerFunc, middlewares ...Middleware) {
	n.mustRegister(RouteDefinition{
		Name:        name,
		Pattern:     n.joinPath(n.prefix, pattern),
		Handler:     handler,
		Middlewares: middlewares,
		Method:      method,
	})
}

// URL generates a URL for a named route, replacing any route parameters with the provided values.
func (n *namedRouter) URL(name string, params RouteParams) (string, error) {
	return n.registry.URL(name, params)
}

// MustURL generates a URL for a named route, replacing any route parameters with the provided values.
// If an error occurs during URL generation, it panics.
func (n *namedRouter) MustURL(name string, params RouteParams) string {
	url, err := n.URL(name, params)
	if err != nil {
		panic(fmt.Errorf("failed to generate URL for route '%s': %w", name, err))
	}

	return url
}

// joinPath joins the prefix and pattern paths, ensuring that the resulting path is properly formatted.
func (n *namedRouter) joinPath(prefix, pattern string) string {
	return n.adapter.JoinPath(prefix, pattern)
}

// register registers a new route definition in the registry and sets up the route in the router.
func (n *namedRouter) register(routeDef RouteDefinition) error {
	if err := n.registry.Register(routeDef); err != nil {
		return err
	}

	adapter := n.adapter
	if len(routeDef.Middlewares) > 0 {
		adapter = adapter.ApplyMiddlewares(routeDef.Pattern, routeDef.Middlewares...)
	}

	adapter.Register(routeDef.Method, routeDef.Pattern, routeDef.Handler)

	return nil
}

// mustRegister registers a new route definition in the registry and sets up the route in the router.
func (n *namedRouter) mustRegister(routeDef RouteDefinition) {
	if err := n.register(routeDef); err != nil {
		panic(err)
	}
}

// Walk walks the adapter's routes.
func (n *namedRouter) Walk() {
	n.adapter.Walk()
}

// Use applies the specified middlewares to the router.
func (n *namedRouter) Use(middlewares ...Middleware) {
	n.adapter.Use(middlewares...)
}

// Static serves static files from the specified root directory for the given pattern.
func (n *namedRouter) Static(pattern, root string) {
	n.adapter.Static(pattern, root)
}

// URLParam retrieves the value of a URL parameter from the request using the adapter's URLParam method.
func (n *namedRouter) URLParam(r *http.Request, key string) string {
	return n.adapter.URLParam(r, key)
}

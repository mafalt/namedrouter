[![CI](https://github.com/mafalt/namedrouter/actions/workflows/ci.yml/badge.svg)](https://github.com/mafalt/namedrouter/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/mafalt/namedrouter.svg)](https://pkg.go.dev/github.com/mafalt/namedrouter)
[![License](https://img.shields.io/github/license/mafalt/namedrouter)](LICENSE)

# NamedRouter

A router abstraction for Go applications that provides named routes and
reverse routing while keeping application code independent of the underlying
HTTP router.

NamedRouter separates route management from the HTTP router implementation
through a pluggable adapter.

## Features

- Named routes
- Reverse routing
- Route parameters
- Nested subrouters
- Route-specific middleware
- Subrouter middleware
- Global middleware
- Static file serving
- Support for standard HTTP methods
- Pluggable HTTP router adapters
- Implements `http.Handler`

## Why NamedRouter?

Go HTTP routers provide different APIs and routing semantics. This is useful
when an application is built around a specific router, but it can make the
application tightly coupled to that implementation.

NamedRouter introduces a small abstraction between the application and the
underlying router:
```
Application
     │
     ▼
NamedRouter
     │
     ▼
  Adapter
     │
     ▼
HTTP Router
```

The application works with the `NamedRouter` API, while an adapter translates
those operations to the underlying router.

This allows the same application-level routing code to work with different
HTTP routers.

## Installation

Install NamedRouter together with an adapter for the HTTP router you want to
use (e.g. for [go-chi](https://github.com/go-chi/chi)).

```bash
go get github.com/mafalt/namedrouter
go get github.com/mafalt/namedrouter-chi
```

## Quick Start

Create an adapter and pass it to `namedrouter.New()`:
```go
adapter := chiadapter.New()
router := namedrouter.New(adapter)
```

`NamedRouter` obtains the parameter parser and parameter applier from the
adapter. Application code does not need to know how the underlying router
represents route parameters.

Register a named route:
```go
router.RegisterGet(
	"/users",
	"users",
	getUsers,
)
```

Because `NamedRouter` implements `http.Handler`, it can be passed directly to
an HTTP server:
```go
http.ListenAndServe(":8080", router)
```

## Named Routes

Every route registered with NamedRouter has a unique name.

```go
router.RegisterGet(
	"/users/{id}",
	"user",
	getUser,
)
```

The route name can then be used to generate a URL:
```go
url, err := router.URL(
	"user",
	namedrouter.RouteParams{
		"id": 42,
	},
)
```

For example, this can produce:
```
/users/42
```

Using route names instead of hard-coded URLs allows route patterns to change
without having to update every place where the URL is generated.

## Reverse Routing

`URL` generates a URL for a registered route and validates the supplied
parameters.
```go
url, err := router.URL(
	"user",
	namedrouter.RouteParams{
		"id": 42,
	},
)
```

If the route does not exist, or required parameters are missing or invalid,
URL returns an error.

When an invalid route or parameter represents a programming error,
MustURL can be used instead:

```go
url := router.MustURL(
	"user",
	namedrouter.RouteParams{
		"id": 42,
	},
)
```

`MustURL` panics when URL generation fails.

## Route Parameters

Route parameters are defined by the underlying adapter.

For example, an adapter may support a pattern such as:

```
/users/{id}
```

The parameter values are supplied through RouteParams:

```go
namedrouter.RouteParams{
	"id": 42,
}
```

NamedRouter does not impose a router-specific parameter syntax. Parameter
parsing and parameter application are responsibilities of the adapter.

This allows different adapters to use the parameter syntax and semantics of
their underlying router.

## HTTP Methods

NamedRouter provides convenience methods for the standard HTTP methods:

```go
router.RegisterGet(...)
router.RegisterPost(...)
router.RegisterPut(...)
router.RegisterDelete(...)
router.RegisterPatch(...)
router.RegisterTrace(...)
router.RegisterHead(...)
router.RegisterOptions(...)
```

The generic Register method can be used when the HTTP method is determined
dynamically:

```go
router.Register(
	http.MethodGet,
	"/users",
	"users",
	getUsers,
)
```

The route name is always part of the NamedRouter API, regardless of the
underlying router.

## Middleware

NamedRouter supports middleware at three levels:

1. Global middleware
2. Subrouter middleware
3. Route-specific middleware

### Global Middleware

Use Use to apply middleware to the router:

```go
router.Use(
	loggingMiddleware,
	recoveryMiddleware,
)
```

### Subrouter Middleware

Middleware can be applied when creating a subrouter:

```go
r := router.Subrouter(
	"/admin",
	authenticationMiddleware,
	authorizationMiddleware,
)
```

All routes registered through this subrouter are handled within that
middleware context.

### Route Middleware

Middleware can also be supplied directly to a route registration method:

```go
r.RegisterGet(
	"/users",
	"admin-users",
	getUsers,
	requireUserReadPermission,
)
```

This makes it possible to compose middleware naturally:

```
Router
└── global middleware
    │
    └── /admin
        ├── subrouter middleware
        │
        └── GET /users
            └── route middleware
```

## Subrouters

Subrouters allow routes to be organized under a common path prefix and
middleware context.

```go
tenants := router.Subrouter(
	"/tenants",
	authorizeMiddleware,
	requiresTenantPermissionsMiddleware,
)
```

Routes registered on the subrouter automatically include the prefix:

```go
tenants.RegisterGet(
	"/",
	"tenants",
	getTenants,
)
```

Subrouters can be nested:

```go
tenants := router.Subrouter("/tenants")

dialogs := tenants.Subrouter("/dialog")

dialogs.RegisterGet(
	"/create",
	"tenant-dialog-create",
	getTenantCreateDialog,
)

dialogs.RegisterGet(
	"/edit/{tenantId}",
	"tenant-dialog-edit",
	getTenantEditDialog,
)
```

This results in routes such as:

```
/tenants/
/tenants/dialog/create
/tenants/dialog/edit/{tenantId}
```

A more complete example:

```go
r := router.Subrouter(
	"/tenants",
	authorizeMiddleware,
	requiresTenantPermissionsMiddleware,
)

r.RegisterGet(
	"/",
	"tenants",
	getTenants,
	requireTenantView,
)

dr := r.Subrouter("/dialog")

dr.RegisterGet(
	"/create",
	"tenant-dialog-create",
	getTenantCreateDialog,
	requireTenantCreate,
)

dr.RegisterGet(
	"/edit/{tenantId}",
	"tenant-dialog-edit",
	getTenantEditDialog,
	requireTenantEdit,
)

dr.RegisterGet(
	"/delete/{tenantId}",
	"tenant-dialog-delete",
	getTenantDeleteDialog,
	requireTenantDelete,
)
```

This allows route trees to reflect the structure of the application while
keeping common prefixes and middleware in one place.

## Static Files

Static files can be registered through the common API:

```go
router.Static(
	"/static",
	"./static",
)
```

The adapter is responsible for translating this operation to the underlying
HTTP router.

## Adapters

NamedRouter does not implement HTTP routing itself. Router-specific behavior
is provided by an adapter.

An adapter is responsible for translating NamedRouter operations to the
underlying router.

```
NamedRouter
     │
     ▼
Adapter
     │
     ├── route registration
     ├── subrouters
     ├── middleware
     ├── path handling
     ├── static files
     └── route parameters
     │
     ▼
HTTP Router
```

The adapter also provides the ParameterParser and ParameterApplier
implementations required by NamedRouter.

This keeps router-specific details out of application code.

### Available Adapters

The following adapters are available:

* [Chi Adapter](https://github.com/mafalt/namedrouter-chi)

Additional adapters can be developed independently of NamedRouter.

### Implementing an Adapter

NamedRouter is designed so that adapters can be implemented independently.

An adapter must implement the `Adapter` interface:

```go
type Adapter interface {
	http.Handler


	Subrouter(prefix string, middlewares ...Middleware) Adapter


	Register(method, pattern string, handler http.Handler)
	ApplyMiddlewares(pattern string, middlewares ...Middleware) Adapter
	JoinPath(prefix, pattern string) string
	Use(middlewares ...Middleware)
	Static(pattern, root string)


	ParameterParser() ParameterParser
	ParameterApplier() ParameterApplier


	Walk()
}
```

The adapter owns its router-specific implementation.

For example, the Chi adapter creates and manages its own Chi router:

```go
func New() namedrouter.Adapter {
	return &chiAdapter{
		router: chi.NewRouter(),
	}
}
```

Application code does not need to interact with the underlying router
directly.

For information about implementing adapters, see
[CONTRIBUTING.md](https://github.com/mafalt/namedrouter/CONTRIBUTING.md).


## Route Registry

NamedRouter maintains a route registry internally.

The registry associates route names with their `RouteDefinition`:

```
Route name
    │
    ▼
RouteDefinition
├── Name
├── Pattern
├── Method
├── Handler
├── Middlewares
└── Parameters
```

The registry is responsible for:

validating route names,
detecting duplicate routes,
extracting route parameters through the adapter,
validating parameters used for reverse routing,
generating URLs.

The registry is an internal implementation detail of `NamedRouter`.
Adapters do not need to know about the registry.

## Walking Routes

`Walk` delegates route traversal to the underlying adapter:

```go
router.Walk()
```

The exact behavior and output depend on the adapter implementation.

This is primarily useful for inspecting registered routes, debugging, and
testing.

## Errors

`NamedRouter` provides typed errors for common route and parameter failures.

Examples include:

* RouteAlreadyExistsError
* RouteNotFoundError
* RouteParameterDoesNotExistError
* RouteParameterNotProvidedError
* DuplicateRouteParameterError
* EmptyRouteParameterNameError
* InvalidRouteParameterFormatError
* NilRouteParameterError
* ErrRouteNameEmpty

Typed errors can be inspected with `errors.As` or compared with
`errors.Is` where applicable.

```go
var errRouteNotFound *namedrouter.RouteNotFoundError

if errors.As(err, &errRouteNotFound) {
	// Handle missing route.
}
```

## Design Goals

`NamedRouter` is intentionally small.

The goal is not to implement another HTTP router, but to provide a common
application-facing API for named routes and route management.

In particular:

* `NamedRouter` should not depend on a specific HTTP router.
* Router-specific behavior belongs in adapters.
* The route registry should remain independent of router implementations.
* Parameter parsing and application should be provided by the adapter.
* Application code should not need to know about adapter internals.

## Versioning

NamedRouter follows [Semantic Versioning](https://semver.org/).

The 0.x releases are considered experimental and the public API may change
before 1.0.0.

## Contributing

Contributions are welcome.

If you want to implement an adapter or contribute to NamedRouter itself, see
[CONTRIBUTING.md](https://github.com/mafalt/namedrouter/CONTRIBUTING.md).

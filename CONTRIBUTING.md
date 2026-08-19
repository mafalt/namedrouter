# Contributing to NamedRouter

Thank you for your interest in contributing to NamedRouter.

NamedRouter is intentionally designed as a small abstraction layer between
application code and HTTP router implementations. The most important part of
contributing is therefore preserving the separation between the public
NamedRouter API, the route registry, and router-specific adapters.

## Architecture

NamedRouter consists of three main parts:

```text
Application
     │
     ▼
NamedRouter
     │
     ├── Route Registry
     │
     ▼
  Adapter
     │
     ▼
HTTP Router
```

## NamedRouter

`NamedRouter` provides the application-facing API.

It is responsible for:

* registering named routes,
* maintaining the route registry,
* generating URLs for named routes,
* validating route parameters,
* managing subrouters,
* delegating router-specific operations to the adapter.

`NamedRouter` must not depend on a specific HTTP router implementation.

## Route Registry

The route registry is owned by `NamedRouter`.

The registry is responsible for managing `RouteDefinition` instances and
providing the information required for reverse routing.

Adapters must not know about the route registry.

In particular, an adapter must never:

* register routes in the NamedRouter registry,
* query the registry,
* generate route names,
* perform NamedRouter-specific route validation.

The adapter only implements the operations required to communicate with the
underlying HTTP router.

## Adapter

An adapter translates the NamedRouter API to a specific HTTP router.

For example:

```
NamedRouter
     │
     ▼
ChiAdapter
     │
     ▼
chi.Router
```

or:

```
NamedRouter
     │
     ▼
StdLibAdapter
     │
     ▼
net/http
```

An adapter owns its underlying router instance and is responsible for all
router-specific behavior.

## Adapter Interface

An adapter must implement the following interface:

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

The adapter should expose a constructor that creates and owns its underlying
router.

For example:

```go
func New() namedrouter.Adapter {
	return &chiAdapter{
		router: chi.NewRouter(),
	}
}
```

The application should not have to create the underlying router itself.

## ParameterParser

Route parameter syntax differs between HTTP routers.

`NamedRouter` therefore does not attempt to understand the syntax of a
particular router.

Instead, the adapter provides a `ParameterParser`:

```go
type ParameterParser interface {
	Parse(route RouteDefinition) (RouteParameterNames, error)
}
```

The parser is responsible for extracting parameter names from a
`RouteDefinition` using the syntax supported by the underlying router.

For example, an adapter may support:

```
/users/{id}
/users/{userId}/posts/{postId}
```

and return:

```go
RouteParameterNames{
	"id": {},
}
```

or:

```go
RouteParameterNames{
	"userId": {},
	"postId": {},
}
```

### Important

`ParameterParser` must understand the syntax of the underlying router.

Do not implement a generic parser in `NamedRouter` based on the syntax of
another router.

For example, the Chi adapter should understand Chi's parameter syntax, while
a Gin adapter should understand Gin's parameter syntax.

This keeps router-specific knowledge inside the adapter.

## ParameterApplier

`ParameterApplier` performs the reverse operation:

```go
type ParameterApplier interface {
	Apply(pattern string, params RouteParams) string
}
```

It replaces route parameters in a pattern with the values supplied by
`NamedRouter`.

For example:

```go
pattern := "/users/{id}"

params := namedrouter.RouteParams{
	"id": 42,
}
```

The result should be:

```
/users/42
```

The exact implementation is adapter-specific because the parameter syntax is
adapter-specific.

## Parser and Applier Should Usually Be Adapter-Owned

An adapter exposes both components:

```go
ParameterParser() ParameterParser
ParameterApplier() ParameterApplier
```

This keeps them out of the public construction API of `NamedRouter`.

Application code only needs to do:

```go
adapter := chiadapter.New()
router := namedrouter.New(adapter)
```

It does not need to know that parameter parsing and application are separate
components.

An adapter may reuse an implementation from another package when the
semantics are genuinely identical, but this should be an implementation
detail of the adapter.

Do not introduce dependencies between adapters merely to share small pieces
of implementation unless there is a clear architectural reason.

## Path Handling

Path handling is also adapter-specific.

The adapter provides:

```go
JoinPath(prefix, pattern string) string
```

`NamedRouter` uses this operation when constructing the full route pattern for
the route registry.

The result of `JoinPath` must represent the path as understood by the
underlying router.

The adapter should therefore not assume that path semantics are identical
to those of another router.

## Middleware

`NamedRouter` supports middleware at three levels:

* global middleware,
* subrouter middleware,
* route middleware.

The adapter is responsible for translating these operations to its
underlying router.

### Global middleware

```go
Use(middlewares ...Middleware)
```

This modifies the router's global middleware configuration.

### Subrouter middleware

```go
Subrouter(prefix string, middlewares ...Middleware) Adapter
```

The returned adapter represents the new routing context.

### Route middleware

```go
ApplyMiddlewares(pattern string, middlewares ...Middleware) Adapter
```

This is applied during route registration.

The important distinction is that `ApplyMiddlewares` modifies only the
current route context used for registration. It must not modify the global
router middleware configuration.

## Static Files

Adapters must implement:

```go
Static(pattern, root string)
```

The adapter is responsible for translating this operation to the underlying
router.

`NamedRouter` should not contain router-specific static file handling.

## Route Registration

The adapter receives a fully constructed route pattern from `NamedRouter`:

```go
Register(method, pattern string, handler http.Handler)
```

The adapter should register the route with the underlying router and should
not modify the NamedRouter route registry.

Route names are intentionally not part of the adapter API.

The underlying HTTP router does not need to know that a route has a
`NamedRouter` name.

## HTTP Methods

`NamedRouter` provides convenience methods such as:

```go
RegisterGet(...)
RegisterPost(...)
RegisterPut(...)
RegisterDelete(...)
RegisterPatch(...)
RegisterTrace(...)
RegisterHead(...)
RegisterOptions(...)
```

All of them eventually use:

```go
Register(method, pattern, name, handler, middlewares...)
```

The adapter receives only the HTTP method as a string.

It is responsible for translating the method to the underlying router.

## Subrouters and Prefixes

Subrouters are an important part of the abstraction.

Given:

```go
users := router.Subrouter("/users")
```

and:

```go
posts := users.Subrouter("/posts")
```

a route registered as:

```go
posts.RegisterGet(
	"/",
	"user-posts",
	handler,
)
```

should be registered by the adapter under the resulting path:

```
/users/posts/
```

The adapter must preserve the routing semantics of the underlying router
when creating subrouters.

Do not implement subrouter behavior by inspecting or manipulating another
router's internal representation.

## Do Not Leak the Underlying Router

A key design goal is that `NamedRouter` should not know the concrete router
implementation.

Avoid adding methods such as:

```go
ChiRouter() chi.Router
GinEngine() *gin.Engine
```

to `NamedRouter`.

Likewise, adapter-specific types should not appear in the `NamedRouter`
package.

The abstraction exists specifically to prevent application code from
depending on a particular HTTP router.

## Keep Router-Specific Logic in the Adapter

When implementing a new feature, ask:

Is this behavior required by `NamedRouter` itself, or is it required because
of the underlying HTTP router?

If it is router-specific, it belongs in the adapter.

For example:

|Responsibility|Location|
|--------------|--------|
|Route names|NamedRouter / Registry|
|Duplicate route names|Registry|
|Reverse routing validation|Registry|
|Parameter syntax|Adapter|
|Parameter parsing|Adapter|
|Parameter application|Adapter|
|Path joining|Adapter|
|Middleware application|Adapter|
|Static files|Adapter|
|HTTP method registration|Adapter|
|Subrouter creation|Adapter|

## Adding a New Adapter

A new adapter should normally live in its own repository.

For example:

```
namedrouter
namedrouter-chi
namedrouter-stdlib
namedrouter-gin
```

The adapter repository should contain:

```
LICENSE
README.md
go.mod
...
```

The adapter should depend on NamedRouter rather than the other way around.

```
NamedRouter
    ▲
    │
    ├── ChiAdapter
    ├── StdLibAdapter
    └── GinAdapter
```

NamedRouter must not import an adapter package.

## Adapter README

Every adapter should document:

* which router it adapts,
* how to install it,
* how to create the adapter,
* supported route parameter syntax,
* any router-specific limitations,
* middleware behavior,
* static file behavior,
* examples.

A minimal example should look similar to:

```go
adapter := chiadapter.New()
router := namedrouter.New(adapter)


router.RegisterGet(
	"/users/{id}",
	"user",
	getUser,
)
```

## Testing

Adapters should have tests covering at least:

* route registration,
* HTTP method handling,
* route parameters,
* reverse routing,
* nested subrouters,
* prefixes,
* global middleware,
* subrouter middleware,
* route middleware,
* static files,
* path joining,
* route walking.

Tests should verify observable behavior rather than implementation details of
the underlying router.

For example, prefer testing:

```
GET /users/42
```

instead of inspecting internal fields of a router.

## Backward Compatibility

`NamedRouter` is currently in the 0.1 development phase.

The API may change before 1.0.0.

Nevertheless, changes to the public API should be deliberate and should
preserve the architectural boundaries described in this document.

When changing an interface, consider whether the change belongs to:

* NamedRouter,
* the adapter contract,
* the registry,
* or a specific adapter.

Avoid adding router-specific functionality to the core package simply because
one adapter needs it.

## Pull Requests

Before submitting a pull request:

1. Run the complete test suite.
2. Run `go vet ./...`.
3. Run `go fmt`.
4. Make sure public APIs have appropriate Go documentation.
4. Verify that the change does not introduce a dependency on a concrete
router implementation.
5. Add or update tests for changed behavior.
6. Update documentation when the public API changes.

## Architectural Principle

The most important rule when contributing to `NamedRouter` is:

> **NamedRouter defines the common API. Adapters know the router. The registry
knows the routes.**

Keep those responsibilities separate.

If implementing a feature requires `NamedRouter` to know about Chi, Gin,
`net/http`, or another concrete router, stop and reconsider whether that
responsibility belongs in the adapter instead.

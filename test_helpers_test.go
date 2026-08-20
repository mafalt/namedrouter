package namedrouter_test

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/mafalt/namedrouter"
)

type routeEntry struct {
	method  string
	pattern string
	handler http.Handler
}

type testAdapter struct {
	prefix  string
	parser  namedrouter.ParameterParser
	applier namedrouter.ParameterApplier

	parent             *testAdapter
	children           []*testAdapter
	routes             []routeEntry
	middlewares        []namedrouter.Middleware
	pendingMiddlewares []namedrouter.Middleware
}

type testParser struct{}

func (p *testParser) Parse(route namedrouter.RouteDefinition) (namedrouter.RouteParameterNames, error) {
	params := namedrouter.RouteParameterNames{}
	for i := 0; i < len(route.Pattern); i++ {
		if route.Pattern[i] != '{' {
			continue
		}
		end := strings.IndexByte(route.Pattern[i+1:], '}')
		if end < 0 {
			return nil, &namedrouter.InvalidRouteParameterFormatError{RouteName: route.Name}
		}
		name := route.Pattern[i+1 : i+1+end]
		if name == "" {
			return nil, &namedrouter.EmptyRouteParameterNameError{RouteName: route.Name}
		}
		params[name] = struct{}{}
		i += end + 1
	}
	return params, nil
}

type testApplier struct{}

func (a *testApplier) Apply(pattern string, params namedrouter.RouteParams) string {
	result := pattern
	for name, value := range params {
		placeholder := fmt.Sprintf("{%s}", name)
		if !strings.Contains(result, placeholder) {
			continue
		}
		result = strings.ReplaceAll(result, placeholder, url.PathEscape(fmt.Sprint(value)))
	}
	return result
}

func newTestAdapter() namedrouter.Adapter {
	return &testAdapter{
		prefix:  "",
		parser:  &testParser{},
		applier: &testApplier{},
	}
}

func newNamedRouter() namedrouter.NamedRouter {
	return namedrouter.New(newTestAdapter())
}

func (a *testAdapter) Subrouter(prefix string, middlewares ...namedrouter.Middleware) namedrouter.Adapter {
	child := &testAdapter{
		prefix:  a.JoinPath(a.prefix, prefix),
		parser:  a.parser,
		applier: a.applier,
		parent:  a,
	}
	if len(middlewares) > 0 {
		child.middlewares = append(child.middlewares, middlewares...)
	}
	a.children = append(a.children, child)
	return child
}

func (a *testAdapter) Register(method, pattern string, handler http.HandlerFunc) {
	wrapped := http.Handler(handler)
	for i := len(a.pendingMiddlewares) - 1; i >= 0; i-- {
		wrapped = a.pendingMiddlewares[i](wrapped)
	}
	for i := len(a.middlewares) - 1; i >= 0; i-- {
		wrapped = a.middlewares[i](wrapped)
	}
	a.routes = append(a.routes, routeEntry{method: method, pattern: pattern, handler: wrapped})
	a.pendingMiddlewares = nil
}

func (a *testAdapter) ApplyMiddlewares(pattern string, middlewares ...namedrouter.Middleware) namedrouter.Adapter {
	a.pendingMiddlewares = append([]namedrouter.Middleware(nil), middlewares...)
	return a
}

func (a *testAdapter) Use(middlewares ...namedrouter.Middleware) {
	a.middlewares = append(a.middlewares, middlewares...)
}

func (a *testAdapter) Static(pattern, root string) {
}

func (a *testAdapter) URLParam(r *http.Request, key string) string {
	if r == nil || key == "" {
		return ""
	}
	return r.URL.Query().Get(key)
}

func (a *testAdapter) Walk() {
}

func (a *testAdapter) ParameterParser() namedrouter.ParameterParser {
	return a.parser
}

func (a *testAdapter) ParameterApplier() namedrouter.ParameterApplier {
	return a.applier
}

func (a *testAdapter) JoinPath(prefix, pattern string) string {
	if prefix == "" && pattern == "" {
		return "/"
	}

	return path.Join(prefix, pattern)
}

func (a *testAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if a.serve(w, r) {
		return
	}
	http.NotFound(w, r)
}

func (a *testAdapter) serve(w http.ResponseWriter, r *http.Request) bool {
	for _, route := range a.routes {
		if route.method == r.Method && route.pattern == r.URL.Path {
			route.handler.ServeHTTP(w, r)
			return true
		}
	}
	for _, child := range a.children {
		if child.serve(w, r) {
			return true
		}
	}
	return false
}

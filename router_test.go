package namedrouter_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mafalt/namedrouter"
)

func TestNew(t *testing.T) {
	nr := newNamedRouter()

	if nr == nil {
		t.Fatal("Expected New to return a non-nil NamedRouter")
	}
}

func TestNamedRouter_RegisterGet(t *testing.T) {
	nr := newNamedRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	nr.RegisterGet("/test", "test.get", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	nr.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "test" {
		t.Errorf("Expected body 'test', got %s", w.Body.String())
	}
}

func TestNamedRouter_RegisterPost(t *testing.T) {
	nr := newNamedRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	nr.RegisterPost("/create", "create.post", handler)

	req := httptest.NewRequest("POST", "/create", nil)
	w := httptest.NewRecorder()
	nr.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}
}

func TestNamedRouter_RegisterPut(t *testing.T) {
	nr := newNamedRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	nr.RegisterPut("/update", "update.put", handler)

	req := httptest.NewRequest("PUT", "/update", nil)
	w := httptest.NewRecorder()
	nr.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestNamedRouter_RegisterDelete(t *testing.T) {
	nr := newNamedRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	nr.RegisterDelete("/delete", "delete.delete", handler)

	req := httptest.NewRequest("DELETE", "/delete", nil)
	w := httptest.NewRecorder()
	nr.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestNamedRouter_RegisterPatch(t *testing.T) {
	nr := newNamedRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	nr.RegisterPatch("/patch", "patch.patch", handler)

	req := httptest.NewRequest("PATCH", "/patch", nil)
	w := httptest.NewRecorder()
	nr.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestNamedRouter_RegisterHead(t *testing.T) {
	nr := newNamedRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	nr.RegisterHead("/head", "head.head", handler)

	req := httptest.NewRequest("HEAD", "/head", nil)
	w := httptest.NewRecorder()
	nr.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestNamedRouter_RegisterOptions(t *testing.T) {
	nr := newNamedRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	nr.RegisterOptions("/options", "options.options", handler)

	req := httptest.NewRequest("OPTIONS", "/options", nil)
	w := httptest.NewRecorder()
	nr.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestNamedRouter_RegisterTrace(t *testing.T) {
	nr := newNamedRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	nr.RegisterTrace("/trace", "trace.trace", handler)

	req := httptest.NewRequest("TRACE", "/trace", nil)
	w := httptest.NewRecorder()
	nr.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestNamedRouter_URL_Simple(t *testing.T) {
	nr := newNamedRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	nr.RegisterGet("/test", "test.simple", handler)

	url, err := nr.URL("test.simple", nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if url != "/test" {
		t.Errorf("Expected URL '/test', got %s", url)
	}
}

func TestNamedRouter_URL_WithParameter(t *testing.T) {
	nr := newNamedRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	nr.RegisterGet("/users/{id}", "users.show", handler)

	url, err := nr.URL("users.show", namedrouter.RouteParams{"id": 123})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if url != "/users/123" {
		t.Errorf("Expected URL '/users/123', got %s", url)
	}
}

func TestNamedRouter_URL_WithMultipleParameters(t *testing.T) {
	nr := newNamedRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	nr.RegisterGet("/tenants/{tenantId}/users/{userId}", "tenants.users.show", handler)

	url, err := nr.URL("tenants.users.show", namedrouter.RouteParams{
		"tenantId": 42,
		"userId":   99,
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if url != "/tenants/42/users/99" {
		t.Errorf("Expected URL '/tenants/42/users/99', got %s", url)
	}
}

func TestNamedRouter_URL_RouteNotFound(t *testing.T) {
	nr := newNamedRouter()

	_, err := nr.URL("nonexistent.route", nil)
	if err == nil {
		t.Fatal("Expected error for nonexistent route")
	}

	var routeNotFoundErr *namedrouter.RouteNotFoundError
	if !errors.As(err, &routeNotFoundErr) {
		t.Errorf("Expected RouteNotFoundError, got %T", err)
	}
}

func TestNamedRouter_URL_MissingParameter(t *testing.T) {
	nr := newNamedRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	nr.RegisterGet("/users/{id}", "users.show", handler)

	_, err := nr.URL("users.show", nil)
	if err == nil {
		t.Fatal("Expected error for missing parameter")
	}

	var paramNotProvidedErr *namedrouter.RouteParameterNotProvidedError
	if !errors.As(err, &paramNotProvidedErr) {
		t.Errorf("Expected RouteParameterNotProvidedError, got %T", err)
	}
}

func TestNamedRouter_URL_ExtraParameter(t *testing.T) {
	nr := newNamedRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	nr.RegisterGet("/test", "test.route", handler)

	_, err := nr.URL("test.route", namedrouter.RouteParams{"extra": "param"})
	if err == nil {
		t.Fatal("Expected error for extra parameter")
	}

	var paramDoesNotExistErr *namedrouter.RouteParameterDoesNotExistError
	if !errors.As(err, &paramDoesNotExistErr) {
		t.Errorf("Expected RouteParameterDoesNotExistError, got %T", err)
	}
}

func TestNamedRouter_URL_NilParameter(t *testing.T) {
	nr := newNamedRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	nr.RegisterGet("/users/{id}", "users.show", handler)

	_, err := nr.URL("users.show", namedrouter.RouteParams{"id": nil})
	if err == nil {
		t.Fatal("Expected error for nil parameter")
	}

	var nilParamErr *namedrouter.NilRouteParameterError
	if !errors.As(err, &nilParamErr) {
		t.Errorf("Expected NilRouteParameterError, got %T", err)
	}
}

func TestNamedRouter_URL_SpecialCharacters(t *testing.T) {
	nr := newNamedRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	nr.RegisterGet("/users/{name}", "users.byname", handler)

	url, err := nr.URL("users.byname", namedrouter.RouteParams{"name": "john doe"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if url != "/users/john%20doe" {
		t.Errorf("Expected URL '/users/john%%20doe', got %s", url)
	}
}

func TestNamedRouter_MustURL_Success(t *testing.T) {
	nr := newNamedRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	nr.RegisterGet("/test", "test.route", handler)

	url := nr.MustURL("test.route", nil)
	if url != "/test" {
		t.Errorf("Expected URL '/test', got %s", url)
	}
}

func TestNamedRouter_MustURL_Panic(t *testing.T) {
	nr := newNamedRouter()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected MustURL to panic for nonexistent route")
		}
	}()

	nr.MustURL("nonexistent.route", nil)
}

func TestNamedRouter_URLParam_DelegatesToAdapter(t *testing.T) {
	nr := newNamedRouter()
	req := httptest.NewRequest("GET", "/users?id=42", nil)

	value := nr.URLParam(req, "id")
	if value != "42" {
		t.Fatalf("Expected URLParam to return '42', got %q", value)
	}
}

func TestNamedRouter_URLParam_UnknownKeyReturnsEmptyString(t *testing.T) {
	nr := newNamedRouter()
	req := httptest.NewRequest("GET", "/users?id=42", nil)

	value := nr.URLParam(req, "missing")
	if value != "" {
		t.Fatalf("Expected URLParam to return empty string for unknown key, got %q", value)
	}
}

func TestNamedRouter_URLParam_NilRequestReturnsEmptyString(t *testing.T) {
	nr := newNamedRouter()

	value := nr.URLParam(nil, "id")
	if value != "" {
		t.Fatalf("Expected URLParam to return empty string for nil request, got %q", value)
	}
}

func TestNamedRouter_Subrouter(t *testing.T) {
	nr := newNamedRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	subrouter := nr.Subrouter("/api")
	subrouter.RegisterGet("/test", "api.test", handler)

	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	nr.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	url, err := nr.URL("api.test", nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if url != "/api/test" {
		t.Errorf("Expected URL '/api/test', got %s", url)
	}
}

func TestNamedRouter_SubrouterWithMiddleware(t *testing.T) {
	nr := newNamedRouter()

	middlewareCalled := false
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewareCalled = true
			next.ServeHTTP(w, r)
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	subrouter := nr.Subrouter("/protected", middleware)
	subrouter.RegisterGet("/test", "protected.test", handler)

	req := httptest.NewRequest("GET", "/protected/test", nil)
	w := httptest.NewRecorder()
	nr.ServeHTTP(w, req)

	if !middlewareCalled {
		t.Error("Expected middleware to be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestNamedRouter_NestedSubrouter(t *testing.T) {
	nr := newNamedRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	api := nr.Subrouter("/api")
	v1 := api.Subrouter("/v1")
	v1.RegisterGet("/test", "api.v1.test", handler)

	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	w := httptest.NewRecorder()
	nr.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	url, err := nr.URL("api.v1.test", nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if url != "/api/v1/test" {
		t.Errorf("Expected URL '/api/v1/test', got %s", url)
	}
}

func TestNamedRouter_RegisterWithMiddleware(t *testing.T) {
	nr := newNamedRouter()

	middlewareCalled := false
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewareCalled = true
			next.ServeHTTP(w, r)
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	nr.RegisterGet("/test", "test.middleware", handler, middleware)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	nr.ServeHTTP(w, req)

	if !middlewareCalled {
		t.Error("Expected middleware to be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestNamedRouter_RegisterWithMultipleMiddlewares(t *testing.T) {
	nr := newNamedRouter()

	middleware1Called := false
	middleware2Called := false

	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middleware1Called = true
			next.ServeHTTP(w, r)
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middleware2Called = true
			next.ServeHTTP(w, r)
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	nr.RegisterGet("/test", "test.multimiddleware", handler, middleware1, middleware2)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	nr.ServeHTTP(w, req)

	if !middleware1Called {
		t.Error("Expected middleware1 to be called")
	}
	if !middleware2Called {
		t.Error("Expected middleware2 to be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRouteParameterNames_Contains(t *testing.T) {
	params := namedrouter.RouteParameterNames{
		"id":   struct{}{},
		"name": struct{}{},
	}

	if !params.Contains("id") {
		t.Error("Expected Contains('id') to return true")
	}
	if !params.Contains("name") {
		t.Error("Expected Contains('name') to return true")
	}
	if params.Contains("missing") {
		t.Error("Expected Contains('missing') to return false")
	}
}

func TestRouteParams_Contains(t *testing.T) {
	params := namedrouter.RouteParams{
		"id":   42,
		"name": "jane",
	}

	if !params.Contains("id") {
		t.Error("Expected Contains('id') to return true")
	}
	if !params.Contains("name") {
		t.Error("Expected Contains('name') to return true")
	}
	if params.Contains("missing") {
		t.Error("Expected Contains('missing') to return false")
	}
}

package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type testBot struct {
	started bool
}

func (b *testBot) StartBot() error {
	b.started = true
	return nil
}

func (b *testBot) StopBot() error {
	b.started = false
	return nil
}

func (b *testBot) BotStatus() BotStatus {
	return BotStatus{Running: b.started}
}

func TestRouterExposesOnlyConfigAndBotControls(t *testing.T) {
	router := createRouter(nil, &testBot{})

	allowed := map[string]bool{
		"GET /api/v1/configs":         true,
		"GET /api/v1/configs/:key":    true,
		"PUT /api/v1/configs/:key":    true,
		"DELETE /api/v1/configs/:key": true,
		"GET /api/v1/bot":             true,
		"POST /api/v1/start":          true,
		"POST /api/v1/stop":           true,
	}

	for _, route := range router.Routes() {
		key := route.Method + " " + route.Path
		if !allowed[key] {
			t.Fatalf("unexpected route exposed: %s", key)
		}
		delete(allowed, key)
	}

	for route := range allowed {
		t.Fatalf("expected route missing: %s", route)
	}
}

func TestAccountRoutesAreNotExposed(t *testing.T) {
	router := createRouter(nil, &testBot{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected accounts route to be unavailable, got status %d", rec.Code)
	}
}

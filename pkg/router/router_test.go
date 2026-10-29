package router_test

import (
	"net/http"
	"testing"

	"feldrise.com/balade/pkg/testharness"
)

func TestLegacyRegistrationAndGroupRoutesAreGone(t *testing.T) {
	h := testharness.New(t)
	admin := h.Token(h.CreateUser("admin@example.com", "admin"))

	for _, route := range []struct{ method, path string }{
		{http.MethodPost, "/registrations"},
		{http.MethodGet, "/registrations"},
		{http.MethodGet, "/registrations/1"},
		{http.MethodPut, "/registrations/1/confirm"},
		{http.MethodPut, "/registrations/1/cancel"},
		{http.MethodGet, "/registrations/ramble/1"},
		{http.MethodGet, "/registrations/1/payments"},
		{http.MethodGet, "/registrations/groups/1"},
		{http.MethodGet, "/registrations/groups/ramble/1"},
		{http.MethodPut, "/registrations/groups/1/confirm"},
		{http.MethodPut, "/registrations/groups/1/cancel"},
		{http.MethodGet, "/registrations/admin"},
		{http.MethodPut, "/registrations/admin/1"},
		{http.MethodPut, "/registrations/admin/1/status"},
		{http.MethodDelete, "/registrations/admin/1"},
		{http.MethodPost, "/registrations/admin/bulk-action"},
		{http.MethodGet, "/registrations/admin/groups"},
		{http.MethodPut, "/registrations/admin/groups/1/status"},
		{http.MethodDelete, "/registrations/admin/groups/1"},
		{http.MethodPost, "/payments"},
		{http.MethodGet, "/payments/groups/1"},
	} {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			if response := h.Do(route.method, route.path, map[string]any{}, admin); response.Code != http.StatusNotFound {
				t.Errorf("expected 404, got %d: %s", response.Code, response.Body)
			}
		})
	}
}

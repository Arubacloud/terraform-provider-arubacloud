package testutil

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestNewMockServer(t *testing.T) {
	t.Run("token endpoint returns canned bearer token", func(t *testing.T) {
		called := false
		s := NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		c := NewClient(t, s.URL)
		resp, err := c.DoRequest(context.Background(), http.MethodGet, "/anything", nil, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resp.Body.Close()
		if !called {
			t.Error("handler was not called")
		}
	})

	t.Run("NewBrokenClient returns network error", func(t *testing.T) {
		s := NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})

		c := NewBrokenClient(t, s.URL)
		_, err := c.DoRequest(context.Background(), http.MethodGet, "/anything", nil, nil, nil)
		if err == nil {
			t.Fatal("expected an error from broken client, got nil")
		}
	})

	t.Run("ErrorBodyJSON produces parseable ErrorResponse", func(t *testing.T) {
		body := ErrorBodyJSON("Not Found", "resource does not exist", 404)
		if body == "" {
			t.Fatal("expected non-empty JSON")
		}

		var e types.ErrorResponse
		if err := json.Unmarshal([]byte(body), &e); err != nil {
			t.Fatalf("ErrorBodyJSON produced invalid JSON: %v", err)
		}
		if e.Title == nil || *e.Title != "Not Found" {
			t.Errorf("unexpected title: %v", e.Title)
		}
		if e.Status == nil || *e.Status != 404 {
			t.Errorf("unexpected status: %v", e.Status)
		}
	})
}

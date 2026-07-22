package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCommands(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		method     string
		path       string
		wantStatus int
		wantBody   map[string]string
	}{
		{name: "set password", args: []string{"set-password", "alice", "secret"}, method: http.MethodPut, path: "/user/password", wantStatus: http.StatusNoContent, wantBody: map[string]string{"screen_name": "alice", "password": "secret"}},
		{name: "list users", args: []string{"user", "list"}, method: http.MethodGet, path: "/user", wantStatus: http.StatusOK},
		{name: "show user", args: []string{"user", "show", "alice smith"}, method: http.MethodGet, path: "/user/alice%20smith/account", wantStatus: http.StatusOK},
		{name: "create user", args: []string{"user", "create", "alice", "secret"}, method: http.MethodPost, path: "/user", wantStatus: http.StatusCreated, wantBody: map[string]string{"screen_name": "alice", "password": "secret"}},
		{name: "delete user", args: []string{"user", "delete", "alice"}, method: http.MethodDelete, path: "/user", wantStatus: http.StatusNoContent, wantBody: map[string]string{"screen_name": "alice"}},
		{name: "list sessions", args: []string{"session", "list"}, method: http.MethodGet, path: "/session", wantStatus: http.StatusOK},
		{name: "show session", args: []string{"session", "show", "alice"}, method: http.MethodGet, path: "/session/alice", wantStatus: http.StatusOK},
		{name: "disconnect session", args: []string{"session", "disconnect", "alice"}, method: http.MethodDelete, path: "/session/alice", wantStatus: http.StatusNoContent},
		{name: "send message", args: []string{"message", "send", "admin", "alice", "hello", "there"}, method: http.MethodPost, path: "/instant-message", wantStatus: http.StatusOK, wantBody: map[string]string{"from": "admin", "to": "alice", "text": "hello there"}},
		{name: "version", args: []string{"version"}, method: http.MethodGet, path: "/version", wantStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tt.method || r.URL.EscapedPath() != tt.path {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.EscapedPath())
				}
				if tt.wantBody != nil {
					var body map[string]string
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Fatal(err)
					}
					for key, value := range tt.wantBody {
						if body[key] != value {
							t.Errorf("unexpected %s: %q", key, body[key])
						}
					}
				}
				w.WriteHeader(tt.wantStatus)
				if tt.wantStatus == http.StatusOK {
					_, _ = w.Write([]byte(`{}`))
				}
			}))
			defer server.Close()

			var args []string
			if tt.args[0] == "set-password" || tt.args[0] == "version" {
				args = append([]string{tt.args[0], "-url", server.URL}, tt.args[1:]...)
			} else {
				args = append([]string{tt.args[0], tt.args[1], "-url", server.URL}, tt.args[2:]...)
			}
			var output strings.Builder
			if err := run(args, strings.NewReader(""), &output, server.Client()); err != nil {
				t.Fatal(err)
			}
		})
	}
}

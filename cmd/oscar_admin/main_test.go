package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSetPassword(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/user/password" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["screen_name"] != "alice" || body["password"] != "secret" {
			t.Errorf("unexpected body: %#v", body)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	var output strings.Builder
	err := run([]string{"set-password", "-url", server.URL, "alice", "secret"}, strings.NewReader(""), &output, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	if output.String() != "password updated for alice\n" {
		t.Errorf("unexpected output: %q", output.String())
	}
}

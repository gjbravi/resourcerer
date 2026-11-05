package decision

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAskSendsModelAndToken(t *testing.T) {
	var got struct {
		Model string `json:"model"`
	}
	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		_ = json.NewDecoder(r.Body).Decode(&got)
		_, _ = w.Write([]byte(`{"model":"decider-4b-v2","answers":{}}`))
	}))
	defer srv.Close()

	resp, err := Ask(context.Background(), srv.Client(), srv.URL, "secret", time.Second, Request{Model: "decider-4b-v2"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Model != "decider-4b-v2" || auth != "Bearer secret" || resp.Model != "decider-4b-v2" {
		t.Fatalf("model %q, auth %q, answered by %q", got.Model, auth, resp.Model)
	}
}

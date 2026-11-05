package decision

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Ask posts one request to a /v1/systemone endpoint. token is empty for a self-hosted decider.
func Ask(ctx context.Context, c *http.Client, endpoint, token string, timeout time.Duration, req Request) (Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return Response{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	hr, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return Response{}, err
	}
	hr.Header.Set("Content-Type", "application/json")
	if token != "" {
		hr.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := c.Do(hr)
	if err != nil {
		return Response{}, fmt.Errorf("ask the decision model: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	b, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode != http.StatusOK {
		return Response{}, fmt.Errorf("decision model answered %s: %s", res.Status, bytes.TrimSpace(b[:min(len(b), 300)]))
	}
	var out Response
	if err := json.Unmarshal(b, &out); err != nil {
		return Response{}, fmt.Errorf("read the decision model's answer: %w", err)
	}
	return out, nil
}

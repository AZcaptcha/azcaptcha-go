package azcaptcha_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	azcaptcha "github.com/azcaptcha/azcaptcha-go"
)

type call struct {
	kind    string
	path    string
	payload any
}

type fakeTransport struct {
	responses []map[string]any
	calls     []call
}

func (f *fakeTransport) next(kind, path string, payload any) (map[string]any, error) {
	f.calls = append(f.calls, call{kind, path, payload})
	if len(f.responses) == 0 {
		return nil, fmt.Errorf("unexpected call: %s %s", kind, path)
	}
	response := f.responses[0]
	f.responses = f.responses[1:]
	return response, nil
}

func (f *fakeTransport) PostForm(_ context.Context, path string, form url.Values) (map[string]any, error) {
	flat := map[string]any{}
	for key := range form {
		flat[key] = form.Get(key)
	}
	return f.next("form", path, flat)
}

func (f *fakeTransport) PostJSON(_ context.Context, path string, payload any) (map[string]any, error) {
	return f.next("json", path, payload)
}

func (f *fakeTransport) Get(_ context.Context, path string, query url.Values) (map[string]any, error) {
	flat := map[string]any{}
	for key := range query {
		flat[key] = query.Get(key)
	}
	return f.next("get", path, flat)
}

func fixture(t *testing.T, name string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile("testdata/fixtures/" + name + ".json")
	if err != nil {
		t.Fatalf("fixture %s: %v", name, err)
	}
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("fixture %s: %v", name, err)
	}
	return data
}

func fixtureJSON(t *testing.T, name string) map[string]any {
	t.Helper()
	nested, ok := fixture(t, name)["response"].(map[string]any)
	if !ok {
		t.Fatalf("fixture %s has no response", name)
	}
	body, ok := nested["json"].(map[string]any)
	if !ok {
		t.Fatalf("fixture %s has no json body", name)
	}
	return body
}

func newClient(t *testing.T, transport *fakeTransport, softID int) (*azcaptcha.Client, *[]string) {
	t.Helper()
	var sleeps []string
	client, err := azcaptcha.New("test-key", azcaptcha.Config{
		Transport: transport,
		SoftID:    softID,
		Sleep: func(d time.Duration) {
			sleeps = append(sleeps, d.String())
		},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return client, &sleeps
}

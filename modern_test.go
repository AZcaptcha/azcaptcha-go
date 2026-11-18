package azcaptcha_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	azcaptcha "github.com/azcaptcha/azcaptcha-go"
)

func TestCreateTaskFixture(t *testing.T) {
	fx := fixture(t, "modern-create-task")
	response := fixtureJSON(t, "modern-create-task")
	transport := &fakeTransport{responses: []map[string]any{response}}
	client, _ := newClient(t, transport, 0)
	id, err := client.CreateTask(context.Background(), "RecaptchaV2TaskProxyless", map[string]any{
		"websiteURL": "https://example.com/login",
		"websiteKey": "6Le-wvkSAAAAAPBMRTvw0Q4Muexq9bi0DJwx_mJ-",
	})
	if err != nil || id != 1234567892 {
		t.Fatalf("create = %d, %v", id, err)
	}
	payload := transport.calls[0].payload.(map[string]any)
	expected := fx["request"].(map[string]any)["json"].(map[string]any)["task"]
	if !reflect.DeepEqual(payload["task"], expected) {
		t.Fatalf("task = %+v", payload["task"])
	}
	if payload["clientKey"] != "test-key" {
		t.Fatalf("payload = %+v", payload)
	}
}

func TestGetTaskResultFixture(t *testing.T) {
	transport := &fakeTransport{responses: []map[string]any{fixtureJSON(t, "modern-task-ready")}}
	client, _ := newClient(t, transport, 0)
	body, err := client.GetTaskResult(context.Background(), 1234567892)
	if err != nil || body["status"] != "ready" {
		t.Fatalf("body = %+v, %v", body, err)
	}
	solution := body["solution"].(map[string]any)
	if solution["gRecaptchaResponse"] != "03AGdBq26..." {
		t.Fatalf("solution = %+v", solution)
	}
}

func TestWaitTaskResult(t *testing.T) {
	transport := &fakeTransport{responses: []map[string]any{
		{"errorId": float64(0), "status": "processing"},
		{"errorId": float64(0), "status": "ready", "solution": map[string]any{"token": "abc"}},
	}}
	client, sleeps := newClient(t, transport, 0)
	solution, err := client.WaitTaskResult(context.Background(), 1, 0)
	if err != nil || solution["token"] != "abc" {
		t.Fatalf("solution = %+v, %v", solution, err)
	}
	if len(*sleeps) != 1 || (*sleeps)[0] != "1s" {
		t.Fatalf("sleeps = %v", *sleeps)
	}
}

func TestModernErrorMapping(t *testing.T) {
	transport := &fakeTransport{responses: []map[string]any{
		{"errorId": float64(12), "errorCode": "ERROR_ZERO_BALANCE", "errorDescription": "no funds"},
	}}
	client, _ := newClient(t, transport, 0)
	_, err := client.CreateTask(context.Background(), "RecaptchaV2TaskProxyless", nil)
	if !errors.Is(err, azcaptcha.ErrBalance) {
		t.Fatalf("err = %v", err)
	}
}

func TestWaitTaskResultTimeout(t *testing.T) {
	transport := &fakeTransport{responses: []map[string]any{
		{"errorId": float64(0), "status": "processing"},
	}}
	client, _ := newClient(t, transport, 0)
	_, err := client.WaitTaskResult(context.Background(), 1, 1)
	if !errors.Is(err, azcaptcha.ErrResult) {
		t.Fatalf("err = %v", err)
	}
}

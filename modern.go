package azcaptcha

import (
	"context"
	"time"
)

// CreateTask submits a modern-API task and returns its numeric id.
func (c *Client) CreateTask(ctx context.Context, taskType string, task map[string]any) (int, error) {
	merged := make(map[string]any, len(task)+1)
	merged["type"] = taskType
	for key, value := range task {
		if value != nil {
			merged[key] = value
		}
	}
	payload := map[string]any{"clientKey": c.apiKey, "task": merged}
	if c.softID > 0 {
		payload["softId"] = c.softID
	}
	body, err := c.http.PostJSON(ctx, "/createTask", payload)
	if err != nil {
		return 0, err
	}
	if err := raiseModern(body); err != nil {
		return 0, err
	}
	id, _ := body["taskId"].(float64)
	return int(id), nil
}

// GetTaskResult polls the modern API once.
func (c *Client) GetTaskResult(ctx context.Context, taskID int) (map[string]any, error) {
	body, err := c.http.PostJSON(ctx, "/getTaskResult", map[string]any{
		"clientKey": c.apiKey, "taskId": taskID,
	})
	if err != nil {
		return nil, err
	}
	if err := raiseModern(body); err != nil {
		return nil, err
	}
	return body, nil
}

// WaitTaskResult polls until the task is ready or the timeout passes.
func (c *Client) WaitTaskResult(ctx context.Context, taskID int, timeout time.Duration) (map[string]any, error) {
	if timeout == 0 {
		timeout = 300 * time.Second
	}
	deadline := time.Now().Add(timeout)
	delay := firstDelay
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		body, err := c.GetTaskResult(ctx, taskID)
		if err != nil {
			return nil, err
		}
		if body["status"] == "ready" {
			if solution, ok := body["solution"].(map[string]any); ok {
				return solution, nil
			}
			return map[string]any{}, nil
		}
		if time.Now().Add(delay).After(deadline) {
			return nil, ErrFor("ERROR_TIMEOUT")
		}
		c.sleep(delay)
		delay = pollInterval
	}
}

func raiseModern(body map[string]any) error {
	errorID, _ := body["errorId"].(float64)
	if errorID == 0 {
		return nil
	}
	code, _ := body["errorCode"].(string)
	if code == "" {
		code = "ERROR_UNKNOWN"
	}
	err := ErrFor(code)
	if description, _ := body["errorDescription"].(string); description != "" {
		err.Message = description
	}
	return err
}

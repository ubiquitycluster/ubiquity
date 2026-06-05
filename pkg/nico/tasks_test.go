package nico

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestPollTaskSucceedsAfterPendingStates(t *testing.T) {
	statuses := []string{"pending", "running", "succeeded"}
	var calls atomic.Int64
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idx := int(calls.Add(1)) - 1
		if idx >= len(statuses) {
			idx = len(statuses) - 1
		}
		_, _ = w.Write([]byte(`{"id":"task-1","status":"` + statuses[idx] + `"}`))
	}))
	defer ts.Close()
	c, err := NewClient(Config{BaseURL: ts.URL, Org: "acme", Token: "tok"}.WithDefaults())
	if err != nil {
		t.Fatal(err)
	}
	task, err := c.PollTask(context.Background(), "task-1", PollOptions{Interval: time.Millisecond, Timeout: time.Second})
	if err != nil {
		t.Fatalf("PollTask: %v", err)
	}
	if task.Status != TaskSucceeded || calls.Load() != 3 {
		t.Fatalf("task=%#v calls=%d", task, calls.Load())
	}
}

func TestPollTaskFailsOnTerminalFailedOrCancelled(t *testing.T) {
	for _, status := range []TaskStatus{TaskFailed, TaskCancelled} {
		t.Run(string(status), func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"id":"task-1","status":"` + string(status) + `","error":"bad"}`))
			}))
			defer ts.Close()
			c, err := NewClient(Config{BaseURL: ts.URL, Org: "acme", Token: "tok"}.WithDefaults())
			if err != nil {
				t.Fatal(err)
			}
			_, err = c.PollTask(context.Background(), "task-1", PollOptions{Interval: time.Millisecond, Timeout: time.Second})
			if err == nil || !strings.Contains(err.Error(), string(status)) {
				t.Fatalf("err = %v, want status %s", err, status)
			}
		})
	}
}

func TestPollTaskTimesOut(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":"task-1","status":"running"}`))
	}))
	defer ts.Close()
	c, err := NewClient(Config{BaseURL: ts.URL, Org: "acme", Token: "tok"}.WithDefaults())
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.PollTask(context.Background(), "task-1", PollOptions{Interval: time.Millisecond, Timeout: 5 * time.Millisecond})
	if !errors.Is(err, ErrTaskPollTimeout) {
		t.Fatalf("err = %v, want ErrTaskPollTimeout", err)
	}
}

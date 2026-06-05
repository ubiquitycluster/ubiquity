package nico

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type TaskStatus string

const (
	TaskPending   TaskStatus = "pending"
	TaskRunning   TaskStatus = "running"
	TaskSucceeded TaskStatus = "succeeded"
	TaskFailed    TaskStatus = "failed"
	TaskCancelled TaskStatus = "cancelled"
)

type Task struct {
	ID        string     `json:"id,omitempty"`
	Status    TaskStatus `json:"status,omitempty"`
	Error     string     `json:"error,omitempty"`
	MachineID string     `json:"machineId,omitempty"`
	Action    string     `json:"action,omitempty"`
}

type PollOptions struct {
	Interval time.Duration
	Timeout  time.Duration
}

var ErrTaskPollTimeout = errors.New("task poll timeout")

func (c *Client) GetTask(ctx context.Context, id string) (Task, error) {
	var out Task
	return out, c.get(ctx, c.resourcePath("task", id), &out)
}

func (c *Client) PollTask(ctx context.Context, id string, opts PollOptions) (Task, error) {
	if opts.Interval <= 0 {
		opts.Interval = time.Second
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 5 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	for {
		task, err := c.GetTask(ctx, id)
		if err != nil {
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return task, ErrTaskPollTimeout
			}
			return Task{}, err
		}
		switch task.Status {
		case TaskSucceeded:
			return task, nil
		case TaskFailed, TaskCancelled:
			if task.Error != "" {
				return task, fmt.Errorf("task %s %s: %s", id, task.Status, task.Error)
			}
			return task, fmt.Errorf("task %s %s", id, task.Status)
		}

		timer := time.NewTimer(opts.Interval)
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return task, ErrTaskPollTimeout
			}
			return task, ctx.Err()
		case <-timer.C:
		}
	}
}

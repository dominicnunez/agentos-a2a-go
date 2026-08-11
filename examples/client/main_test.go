package main

import (
	"testing"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

func TestSummarizeResult(t *testing.T) {
	tests := []struct {
		name   string
		result a2a.SendMessageResult
		want   string
	}{
		{
			name:   "task",
			result: &a2a.Task{ID: "task-1", Status: a2a.TaskStatus{State: a2a.TaskStateCompleted}},
			want:   "task task-1: TASK_STATE_COMPLETED",
		},
		{
			name:   "message",
			result: &a2a.Message{ID: "message-1"},
			want:   "message message-1 received",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := summarizeResult(test.result)
			if err != nil {
				t.Fatalf("summarizeResult() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("summarizeResult() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestSummarizeResultRejectsNil(t *testing.T) {
	for _, result := range []a2a.SendMessageResult{(*a2a.Task)(nil), (*a2a.Message)(nil), nil} {
		if _, err := summarizeResult(result); err == nil {
			t.Fatalf("summarizeResult(%T) error = nil, want error", result)
		}
	}
}

package main

import (
	"errors"
	"net/http"
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

func TestBearerTransportAcceptsEquivalentOrigin(t *testing.T) {
	origin, err := parseOrigin("https://AGENT.example:0443/card")
	if err != nil {
		t.Fatalf("parseOrigin() error = %v", err)
	}
	called := false
	transport := bearerTransport{
		origin: origin,
		token:  "secret-token",
		base: roundTripperFunc(func(request *http.Request) (*http.Response, error) {
			called = true
			if got := request.Header.Get("Authorization"); got != "Bearer secret-token" {
				t.Fatalf("Authorization = %q, want bearer token", got)
			}
			return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
		}),
	}
	request, err := http.NewRequest(http.MethodPost, "https://agent.example/rpc", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	if _, err := transport.RoundTrip(request); err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	if !called {
		t.Fatal("base transport was not called")
	}
}

func TestBearerTransportRejectsDifferentEffectivePort(t *testing.T) {
	origin, err := parseOrigin("https://agent.example")
	if err != nil {
		t.Fatalf("parseOrigin() error = %v", err)
	}
	transport := bearerTransport{
		origin: origin,
		token:  "secret-token",
		base: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("base transport must not be called")
		}),
	}
	request, err := http.NewRequest(http.MethodPost, "https://agent.example:444/rpc", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	if _, err := transport.RoundTrip(request); err == nil {
		t.Fatal("RoundTrip() error = nil, want origin rejection")
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (function roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

package intentconfirmation

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

func TestRoundTripAndIdempotentSet(t *testing.T) {
	fingerprint := strings.Repeat("a", 64)
	message := boundMessage()
	if err := Set(message, Confirmation{Fingerprint: fingerprint}); err != nil {
		t.Fatal(err)
	}
	if err := Set(message, Confirmation{Fingerprint: fingerprint}); err != nil {
		t.Fatal(err)
	}
	wire, err := json.Marshal(message)
	if err != nil {
		t.Fatal(err)
	}
	var decoded a2a.Message
	if err := json.Unmarshal(wire, &decoded); err != nil {
		t.Fatal(err)
	}
	got, present, err := Get(&decoded)
	if err != nil || !present || got.Fingerprint != fingerprint || len(decoded.Extensions) != 1 {
		t.Fatalf("Get()=(%+v,%v,%v)", got, present, err)
	}
}

func TestRejectsMalformedState(t *testing.T) {
	fingerprint := strings.Repeat("b", 64)
	tests := []struct {
		message *a2a.Message
		want    error
	}{
		{nil, ErrNilMessage},
		{messageWith(nil, confirmationMetadata(fingerprint)), ErrMissingDeclaration},
		{messageWith([]string{URI}, nil), ErrMissingMetadata},
		{messageWith([]string{URI, URI}, confirmationMetadata(fingerprint)), ErrDuplicateDeclaration},
		{messageWith([]string{URI}, map[string]any{URI: map[string]any{"action": "APPROVE", "fingerprint": fingerprint}}), ErrMalformedMetadata},
		{messageWith([]string{URI}, confirmationMetadata("short")), ErrInvalidFingerprint},
		{messageWith([]string{URI}, map[string]any{URI: map[string]any{"action": "CONFIRM", "fingerprint": fingerprint, "approval": true}}), ErrMalformedMetadata},
	}
	for _, test := range tests {
		if _, _, err := Get(test.message); !errors.Is(err, test.want) {
			t.Fatalf("Get(%#v) error=%v want=%v", test.message, err, test.want)
		}
	}
}

func TestAbsentAndInvalidSet(t *testing.T) {
	if confirmation, present, err := Get(&a2a.Message{}); err != nil || present || confirmation.Fingerprint != "" {
		t.Fatalf("Get()=(%+v,%v,%v)", confirmation, present, err)
	}
	if err := Set(nil, Confirmation{Fingerprint: strings.Repeat("a", 64)}); !errors.Is(err, ErrNilMessage) {
		t.Fatal(err)
	}
	if err := Set(boundMessage(), Confirmation{Fingerprint: strings.Repeat("A", 64)}); !errors.Is(err, ErrInvalidFingerprint) {
		t.Fatal(err)
	}
}

func TestRequiresDurableTaskBinding(t *testing.T) {
	fingerprint := strings.Repeat("c", 64)
	for _, message := range []*a2a.Message{
		{TaskID: "task-1", ContextID: "context-1"},
		{ID: "message-1", ContextID: "context-1"},
		{ID: "message-1", TaskID: "task-1"},
		{ID: " ", TaskID: "task-1", ContextID: "context-1"},
	} {
		if err := Set(message, Confirmation{Fingerprint: fingerprint}); !errors.Is(err, ErrMissingBinding) {
			t.Fatalf("Set(%#v) error=%v want=%v", message, err, ErrMissingBinding)
		}
		message.Extensions = []string{URI}
		message.Metadata = confirmationMetadata(fingerprint)
		if _, present, err := Get(message); present || !errors.Is(err, ErrMissingBinding) {
			t.Fatalf("Get(%#v) present=%v error=%v want=%v", message, present, err, ErrMissingBinding)
		}
	}
}

func boundMessage() *a2a.Message {
	message := a2a.NewMessage(a2a.MessageRoleUser, a2a.NewTextPart("Confirm the reviewed Intent."))
	message.TaskID = "task-1"
	message.ContextID = "context-1"
	return message
}

func messageWith(extensions []string, metadata map[string]any) *a2a.Message {
	message := boundMessage()
	message.Extensions = extensions
	message.Metadata = metadata
	return message
}

func confirmationMetadata(fingerprint string) map[string]any {
	return map[string]any{URI: map[string]any{"action": "CONFIRM", "fingerprint": fingerprint}}
}

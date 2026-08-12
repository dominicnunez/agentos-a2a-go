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
	message := a2a.NewMessage(a2a.MessageRoleUser, a2a.NewTextPart("Confirm the reviewed Intent."))
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
		{&a2a.Message{Metadata: map[string]any{URI: map[string]any{"action": "CONFIRM", "fingerprint": fingerprint}}}, ErrMissingDeclaration},
		{&a2a.Message{Extensions: []string{URI}}, ErrMissingMetadata},
		{&a2a.Message{Extensions: []string{URI, URI}, Metadata: map[string]any{URI: map[string]any{"action": "CONFIRM", "fingerprint": fingerprint}}}, ErrDuplicateDeclaration},
		{&a2a.Message{Extensions: []string{URI}, Metadata: map[string]any{URI: map[string]any{"action": "APPROVE", "fingerprint": fingerprint}}}, ErrMalformedMetadata},
		{&a2a.Message{Extensions: []string{URI}, Metadata: map[string]any{URI: map[string]any{"action": "CONFIRM", "fingerprint": "short"}}}, ErrInvalidFingerprint},
		{&a2a.Message{Extensions: []string{URI}, Metadata: map[string]any{URI: map[string]any{"action": "CONFIRM", "fingerprint": fingerprint, "approval": true}}}, ErrMalformedMetadata},
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
	if err := Set(&a2a.Message{}, Confirmation{Fingerprint: strings.Repeat("A", 64)}); !errors.Is(err, ErrInvalidFingerprint) {
		t.Fatal(err)
	}
}

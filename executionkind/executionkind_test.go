package executionkind

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

func TestJSONRoundTrip(t *testing.T) {
	message := a2a.NewMessage(a2a.MessageRoleUser, a2a.NewTextPart("work"))
	if err := Set(message, KindDeterministic); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	wire, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	var decoded a2a.Message
	if err := json.Unmarshal(wire, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	kind, present, err := Get(&decoded)
	if err != nil || !present || kind != KindDeterministic {
		t.Fatalf("Get() = (%q, %v, %v), want (%q, true, nil)", kind, present, err, KindDeterministic)
	}
}

func TestRoundTripAndIdempotentSet(t *testing.T) {
	message := a2a.NewMessage(a2a.MessageRoleUser, a2a.NewTextPart("work"))
	if err := Set(message, KindAgent); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	wantExtensions := []string{URI}
	if !reflect.DeepEqual(message.Extensions, wantExtensions) {
		t.Fatalf("extensions = %#v, want %#v", message.Extensions, wantExtensions)
	}
	if err := Set(message, KindAgent); err != nil {
		t.Fatalf("second Set() error = %v", err)
	}
	if !reflect.DeepEqual(message.Extensions, wantExtensions) {
		t.Fatalf("extensions after second Set = %#v, want %#v", message.Extensions, wantExtensions)
	}
	kind, present, err := Get(message)
	if err != nil || !present || kind != KindAgent {
		t.Fatalf("Get() = (%q, %v, %v), want (%q, true, nil)", kind, present, err, KindAgent)
	}
}

func TestSetUpdatesValidMetadata(t *testing.T) {
	message := &a2a.Message{
		Extensions: []string{URI},
		Metadata:   map[string]any{URI: map[string]any{"kind": string(KindAgent)}},
	}
	if err := Set(message, KindHuman); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	kind, present, err := Get(message)
	if err != nil || !present || kind != KindHuman {
		t.Fatalf("Get() = (%q, %v, %v), want (%q, true, nil)", kind, present, err, KindHuman)
	}
}

func TestAbsent(t *testing.T) {
	kind, present, err := Get(&a2a.Message{})
	if err != nil || present || kind != "" {
		t.Fatalf("Get() = (%q, %v, %v), want absent", kind, present, err)
	}
}

func TestRejectsInvalidState(t *testing.T) {
	tests := []struct {
		name    string
		message *a2a.Message
		want    error
	}{
		{name: "nil", want: ErrNilMessage},
		{name: "missing declaration", message: &a2a.Message{Metadata: map[string]any{URI: map[string]any{"kind": "AGENT"}}}, want: ErrMissingDeclaration},
		{name: "missing metadata", message: &a2a.Message{Extensions: []string{URI}}, want: ErrMissingMetadata},
		{name: "duplicate declaration", message: &a2a.Message{Extensions: []string{URI, URI}, Metadata: map[string]any{URI: map[string]any{"kind": "AGENT"}}}, want: ErrDuplicateDeclaration},
		{name: "metadata scalar", message: &a2a.Message{Extensions: []string{URI}, Metadata: map[string]any{URI: "AGENT"}}, want: ErrMalformedMetadata},
		{name: "metadata extra field", message: &a2a.Message{Extensions: []string{URI}, Metadata: map[string]any{URI: map[string]any{"kind": "AGENT", "authority": true}}}, want: ErrMalformedMetadata},
		{name: "kind non-string", message: &a2a.Message{Extensions: []string{URI}, Metadata: map[string]any{URI: map[string]any{"kind": 1}}}, want: ErrMalformedMetadata},
		{name: "unknown kind", message: &a2a.Message{Extensions: []string{URI}, Metadata: map[string]any{URI: map[string]any{"kind": "TOOL"}}}, want: ErrUnknownKind},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := Get(test.message)
			if !errors.Is(err, test.want) {
				t.Fatalf("Get() error = %v, want errors.Is(%v)", err, test.want)
			}
		})
	}
}

func TestSetRejectsInvalidInput(t *testing.T) {
	if err := Set(nil, KindAgent); !errors.Is(err, ErrNilMessage) {
		t.Fatalf("Set(nil) error = %v", err)
	}
	if err := Set(&a2a.Message{}, Kind("TOOL")); !errors.Is(err, ErrUnknownKind) {
		t.Fatalf("Set(unknown) error = %v", err)
	}
	message := &a2a.Message{Extensions: []string{URI, URI}}
	if err := Set(message, KindAgent); !errors.Is(err, ErrDuplicateDeclaration) {
		t.Fatalf("Set(duplicate) error = %v", err)
	}
}

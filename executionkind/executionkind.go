// Package executionkind implements the Agent OS execution-kind A2A extension.
//
// Execution kind is an untrusted routing hint. It does not grant authority,
// approval, capability, completion status, or permission to perform effects.
package executionkind

import (
	"errors"
	"fmt"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

// URI is the canonical identifier for version 1 of the extension.
const URI = "https://github.com/dominicnunez/agentos-a2a-go/blob/main/spec/execution-kind-v1.md"

// Kind identifies a requested execution strategy.
type Kind string

const (
	KindDeterministic Kind = "DETERMINISTIC"
	KindAgent         Kind = "AGENT"
	KindHuman         Kind = "HUMAN"
)

var (
	ErrNilMessage           = errors.New("executionkind: message is nil")
	ErrUnknownKind          = errors.New("executionkind: unknown kind")
	ErrDuplicateDeclaration = errors.New("executionkind: duplicate extension declaration")
	ErrMissingDeclaration   = errors.New("executionkind: metadata is missing its extension declaration")
	ErrMissingMetadata      = errors.New("executionkind: declared extension has no metadata")
	ErrMalformedMetadata    = errors.New("executionkind: malformed metadata")
)

// Set declares the extension and writes its metadata. Repeating Set with the
// same message and kind is idempotent. Existing malformed or duplicate state
// is rejected instead of silently repaired.
func Set(message *a2a.Message, kind Kind) error {
	if message == nil {
		return ErrNilMessage
	}
	if err := validateKind(kind); err != nil {
		return err
	}
	declarations := declarationCount(message.Extensions)
	if declarations > 1 {
		return ErrDuplicateDeclaration
	}
	raw, hasMetadata := message.Metadata[URI]
	if declarations == 0 && hasMetadata {
		return ErrMissingDeclaration
	}
	if declarations == 1 && !hasMetadata {
		return ErrMissingMetadata
	}
	if hasMetadata {
		if _, err := decode(raw); err != nil {
			return err
		}
	}
	if declarations == 0 {
		message.Extensions = append(message.Extensions, URI)
	}
	if message.Metadata == nil {
		message.Metadata = make(map[string]any)
	}
	message.Metadata[URI] = map[string]any{"kind": string(kind)}
	return nil
}

// Get reads and validates the execution-kind extension.
func Get(message *a2a.Message) (kind Kind, present bool, err error) {
	if message == nil {
		return "", false, ErrNilMessage
	}
	declarations := declarationCount(message.Extensions)
	if declarations > 1 {
		return "", false, ErrDuplicateDeclaration
	}
	raw, hasMetadata := message.Metadata[URI]
	if declarations == 0 {
		if hasMetadata {
			return "", false, ErrMissingDeclaration
		}
		return "", false, nil
	}
	if !hasMetadata {
		return "", false, ErrMissingMetadata
	}
	kind, err = decode(raw)
	if err != nil {
		return "", false, err
	}
	return kind, true, nil
}

func declarationCount(extensions []string) int {
	count := 0
	for _, extension := range extensions {
		if extension == URI {
			count++
		}
	}
	return count
}

func decode(raw any) (Kind, error) {
	object, ok := raw.(map[string]any)
	if !ok || len(object) != 1 {
		return "", ErrMalformedMetadata
	}
	value, ok := object["kind"].(string)
	if !ok {
		return "", ErrMalformedMetadata
	}
	kind := Kind(value)
	if err := validateKind(kind); err != nil {
		return "", err
	}
	return kind, nil
}

func validateKind(kind Kind) error {
	switch kind {
	case KindDeterministic, KindAgent, KindHuman:
		return nil
	default:
		return fmt.Errorf("%w %q", ErrUnknownKind, kind)
	}
}

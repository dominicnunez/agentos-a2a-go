// Package intentconfirmation implements the Agent OS Intent-confirmation A2A
// extension. Confirmation admits an exact reviewed Intent to planning; it does
// not grant approval, capability, completion status, or effect permission.
package intentconfirmation

import (
	"errors"
	"fmt"
	"strings"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

const URI = "https://github.com/dominicnunez/agentos-a2a-go/blob/main/spec/intent-confirmation-v1.md"

type Confirmation struct {
	Fingerprint string
}

var (
	ErrNilMessage           = errors.New("intentconfirmation: message is nil")
	ErrInvalidFingerprint   = errors.New("intentconfirmation: invalid fingerprint")
	ErrDuplicateDeclaration = errors.New("intentconfirmation: duplicate extension declaration")
	ErrMissingDeclaration   = errors.New("intentconfirmation: metadata is missing its extension declaration")
	ErrMissingMetadata      = errors.New("intentconfirmation: declared extension has no metadata")
	ErrMalformedMetadata    = errors.New("intentconfirmation: malformed metadata")
)

func Set(message *a2a.Message, confirmation Confirmation) error {
	if message == nil {
		return ErrNilMessage
	}
	if err := validateFingerprint(confirmation.Fingerprint); err != nil {
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
	message.Metadata[URI] = map[string]any{"action": "CONFIRM", "fingerprint": confirmation.Fingerprint}
	return nil
}

func Get(message *a2a.Message) (Confirmation, bool, error) {
	if message == nil {
		return Confirmation{}, false, ErrNilMessage
	}
	declarations := declarationCount(message.Extensions)
	if declarations > 1 {
		return Confirmation{}, false, ErrDuplicateDeclaration
	}
	raw, hasMetadata := message.Metadata[URI]
	if declarations == 0 {
		if hasMetadata {
			return Confirmation{}, false, ErrMissingDeclaration
		}
		return Confirmation{}, false, nil
	}
	if !hasMetadata {
		return Confirmation{}, false, ErrMissingMetadata
	}
	confirmation, err := decode(raw)
	if err != nil {
		return Confirmation{}, false, err
	}
	return confirmation, true, nil
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

func decode(raw any) (Confirmation, error) {
	object, ok := raw.(map[string]any)
	if !ok || len(object) != 2 || object["action"] != "CONFIRM" {
		return Confirmation{}, ErrMalformedMetadata
	}
	fingerprint, ok := object["fingerprint"].(string)
	if !ok {
		return Confirmation{}, ErrMalformedMetadata
	}
	if err := validateFingerprint(fingerprint); err != nil {
		return Confirmation{}, err
	}
	return Confirmation{Fingerprint: fingerprint}, nil
}

func validateFingerprint(fingerprint string) error {
	if len(fingerprint) != 64 {
		return fmt.Errorf("%w: expected 64 lowercase hexadecimal characters", ErrInvalidFingerprint)
	}
	for _, character := range fingerprint {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return fmt.Errorf("%w: expected 64 lowercase hexadecimal characters", ErrInvalidFingerprint)
		}
	}
	return nil
}

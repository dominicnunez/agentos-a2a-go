# Agent OS Intent confirmation extension, version 1

Canonical URI:

```text
https://github.com/dominicnunez/agentos-a2a-go/blob/main/spec/intent-confirmation-v1.md
```

An authenticated A2A actor uses this extension only after Agent OS returns a
reviewable, versioned Intent and its SHA-256 fingerprint. The confirmation
message declares the URI in `Message.Extensions` and carries exactly:

```json
{
  "https://github.com/dominicnunez/agentos-a2a-go/blob/main/spec/intent-confirmation-v1.md": {
    "action": "CONFIRM",
    "fingerprint": "64 lowercase hexadecimal characters"
  }
}
```

The message must also carry the durable `taskId`, matching `contextId`, and a
new `messageId`. Missing declarations or metadata, duplicate declarations,
unknown fields, other actions, and malformed fingerprints are invalid.

The package validates that `taskId`, `contextId`, and `messageId` are
nonempty. The receiving server must compare the task and context identifiers
with its durable task state and reject a previously used message identifier;
those stateful checks cannot be established by extension metadata alone.

Confirmation means the actor accepts Agent OS's exact interpretation of the
requested work and permits planning to begin. It grants no approval,
capability, completion status, policy change, or permission to perform an
effect. Consequential effects remain subject to their independent Agent OS
authorization and exact-effect approval boundaries.

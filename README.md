# Agent OS A2A Go extensions

`agentos-a2a-go` contains small, independently usable extensions and examples
for connecting A2A-capable agents to Agent OS. It builds on the official
[`a2aproject/a2a-go/v2`](https://github.com/a2aproject/a2a-go) SDK and is not a
replacement SDK or an Agent OS client wrapper.

The initial `executionkind` package encodes an untrusted execution-routing hint
using A2A `Message.Extensions` and `Message.Metadata`. It never grants authority,
approval, capability, completion status, or effect permission.

```go
message := a2a.NewMessage(a2a.MessageRoleUser, a2a.NewTextPart("echo hello"))
if err := executionkind.Set(message, executionkind.KindDeterministic); err != nil {
    return err
}
```

The `intentconfirmation` package binds an authenticated actor's confirmation
to the exact fingerprint of a reviewed Agent OS Intent. It admits that Intent
to planning but grants no effect authority or approval.

See the [execution-kind](spec/execution-kind-v1.md) and
[Intent-confirmation](spec/intent-confirmation-v1.md) specifications plus the
[reference client](examples/client/main.go). The client reads its endpoint and
bearer credential only from `AGENTOS_A2A_URL` and `AGENTOS_A2A_TOKEN`, never
prints the credential, rejects redirects, and confines authorization to the
configured origin.

## Compatibility

- Go 1.25.0 or newer
- A2A Go SDK v2.4.0
- A2A protocol v1.0

This module has no dependency on the AGPL Agent OS module. CI enforces that
boundary across the complete Go module graph.

## License and contributions

Copyright 2026 Dominic Nunez. Licensed under Apache-2.0; see [LICENSE](LICENSE)
and [NOTICE](NOTICE).

Issues are welcome. External code contributions are not currently accepted;
see [CONTRIBUTING.md](CONTRIBUTING.md).

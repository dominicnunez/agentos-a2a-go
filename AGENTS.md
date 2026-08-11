# Repository guidance

This repository contains independently usable Agent OS extensions for the public A2A protocol.

- Keep the module Apache-2.0 and independent of `github.com/dominicnunez/agentos`.
- Build only on official A2A types and contracts; do not create a competing SDK or copy Agent OS internals.
- Treat extension metadata as untrusted input. Extensions never grant authority, approval, capability, completion, or effect permission.
- Keep examples credential-safe and source-only.

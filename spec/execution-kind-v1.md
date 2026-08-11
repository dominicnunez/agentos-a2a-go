# Execution Kind Extension v1

The execution-kind extension lets an A2A client express a preferred execution
strategy for a message. Its canonical URI is:

```text
https://github.com/dominicnunez/agentos-a2a-go/blob/main/spec/execution-kind-v1.md
```

A message using the extension declares the URI once in `Message.Extensions`
and stores this object at the matching `Message.Metadata` key:

```json
{
  "kind": "AGENT"
}
```

The only values are `DETERMINISTIC`, `AGENT`, and `HUMAN`. The declaration and
metadata must both be present, declarations must not be duplicated, and the
metadata object must contain exactly the string field `kind`.

Execution kind is untrusted routing input. It grants no authority, approval,
capability, completion status, or permission to perform an effect. A receiver
must independently authenticate the caller and enforce its own authorization,
policy, approval, execution, and completion controls. A receiver may reject or
ignore a valid hint.

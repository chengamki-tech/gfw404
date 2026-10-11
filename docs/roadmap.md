# Roadmap

## v0.1.0 — shipped

- Ultra security profile with TLS 1.3, certificate validation, PMTU policy,
  credential checks, authentication attempt limits, and redacted logs.
- Reconnect cleanup and authorized performance benchmarks.

## v0.2.0 — reliability

- `404-Brain` health-probe interface for explicitly configured endpoints.
- Bounded exponential backoff with jitter for reconnect attempts.
- Endpoint scoring from recent RTT, loss, and successful handshakes.
- Metrics for reconnect reason, handshake latency, and route changes.

## v0.3.0 — authorized relay operations

- Explicit HTTPS/WebSocket/gRPC relay adapters with certificate validation.
- Relay health checks and circuit breaking.
- Per-relay bandwidth and connection limits.

## Non-goals

- REALITY/Vision impersonation.
- Port hopping or censorship bypass.
- Simulating another website's traffic timing or packet distribution.
- A claim that the protocol is undetectable or unblockable.

# gfw404 Architecture

`gfw404` is a secure, high-availability transport stack for networks and
servers that the operator is authorized to manage. It preserves the standard
Hysteria 2 wire protocol and uses standard TLS 1.3 and QUIC.

## Components

| Component | Safe responsibility |
| --- | --- |
| `gfw404` | The complete client/server distribution |
| `404-Brain` | Explicit-endpoint health checks, RTT route selection, bounded reconnects, and session recovery |
| `404-Rush` | Standard QUIC/TLS 1.3 transport with PMTU and congestion-control tuning |
| `404-Relay` | Explicitly configured HTTPS/WebSocket/gRPC relay for authorized deployments |
| `404-Veil` | Fixed, documented padding for protocol framing; never site imitation or traffic simulation |

## Deliberate exclusions

This project does not implement REALITY/Vision impersonation, port hopping,
traffic-shape imitation, DPI evasion, or mechanisms intended to bypass a
campus, operator, or other network's controls. Relay protocols must be
identified as relays by their deployment configuration and must not pretend to
be unrelated websites.

## Request flow

1. `404-Brain` checks only explicitly configured endpoints using authenticated,
   rate-limited health probes; it never scans or discovers public endpoints.
2. The client chooses the lowest-latency healthy endpoint according to an
   explicit policy and keeps the active session until it fails.
3. A failure closes the old transport, applies bounded exponential backoff,
   and reconnects to another configured endpoint.
4. `404-Rush` carries application traffic using standard QUIC/TLS 1.3.
5. Metrics record latency, loss, throughput, reconnect count, and reason codes;
   credentials and certificate material are never logged.

The first selector implementation is in `core/brain`. It is transport
agnostic so callers can supply an authorized probe implementation without
coupling endpoint policy to packet handling.

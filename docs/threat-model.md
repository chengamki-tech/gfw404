# Threat Model

## In scope

- Passive observers learning endpoint metadata and traffic volume.
- Active network faults causing loss, delay, reordering, or connection resets.
- Credential guessing and repeated unauthenticated handshakes.
- Malicious or compromised relay endpoints.
- Accidental leakage of credentials through logs, URIs, process arguments, or
  insecure control-plane connections.

## Security properties

- TLS 1.3 is the minimum transport version.
- Certificate verification or an explicit certificate pin is required by the
  Ultra profile.
- Password checks use constant-time comparison.
- Ultra credentials require a minimum length and failed authentication is
  bounded per source IP and per connection.
- Authentication state and endpoint health state are bounded in memory.
- A failed connection is discarded before reconnecting.

## Out of scope

No protocol can guarantee immunity from detection, rate limiting, blocking,
or traffic analysis. The project does not attempt to hide that a configured
transport endpoint is being used, impersonate unrelated websites, or evade
network inspection.

## Operator responsibilities

- Protect private keys and authentication credentials.
- Use HTTPS for external authentication and control-plane services.
- Restrict relay endpoints and management APIs with network policy.
- Rotate certificates and pins using an overlap window.
- Collect only the metrics needed for operations and protect their storage.

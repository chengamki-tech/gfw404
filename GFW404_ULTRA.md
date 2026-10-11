# gfw404 Ultra

`gfw404 Ultra` is an opt-in security and performance profile built on the
upstream Hysteria 2 wire protocol. It is not a new cryptosystem and it does
not change the QUIC packet format.

## Security contract

- QUIC provides the encrypted transport and TLS 1.3 handshake.
- The core explicitly sets TLS 1.3 as the minimum TLS version on both client
  and server configurations.
- The profile uses the Go TLS verifier, an explicit CA, or a SHA-256
  certificate pin as the trust anchor.
- `profile: ultra` rejects `tls.insecure: true` unless `tls.pinSHA256` is
  configured. A custom CA is used with normal certificate verification and
  must not be combined with `tls.insecure: true` as a substitute for it.
- `profile: ultra` rejects explicit `quic.disablePathMTUDiscovery: true`.
- `profile: ultra` keeps QUIC stateless reset support enabled for prompt dead
  connection recovery.
- `profile: ultra` rejects insecure Realm rendezvous transport.
- `profile: ultra` rejects insecure external HTTP authentication transport.
- Ultra rejects command-based authentication because the credential is exposed
  through process arguments; external authentication must use an HTTPS URL.
- A single QUIC connection permits at most three authentication attempts before
  further attempts receive only the configured non-proxy response.
- Ultra also applies a bounded server-side failure limiter: three failed
  attempts per source IP per minute, with at most 4096 tracked source entries;
  successful authentication clears that source's failure state.
- Client logs redact authentication, obfuscation passwords, certificate pins,
  and ECH material; explicit `share` output remains intentionally complete.
- Passwords remain authentication credentials; they are not replacements for
  TLS certificate verification.
- Password and user-password credentials must be at least 16 characters in
  Ultra mode; legacy profiles keep their existing policy.
- Password and user-password authentication comparisons use constant-time
  equality for the secret portion.

Certificate pinning must be rotated deliberately. A stale pin causes a hard
failure, which is preferable to silently accepting an unexpected certificate.

## Performance contract

The profile keeps standard QUIC path-MTU discovery, BBR/Reno congestion
control, pacing, stream flow control, and Hysteria UDP fragmentation. Tune
receive windows and congestion settings only after measuring the actual path.
Larger windows are not automatically faster and can increase queueing and
memory use.

The reconnectable client clears a failed session before retrying, so a
transient handshake or path failure cannot leave later requests attached to a
known-closed connection.

Use the benchmark harness in `benchmarks/` and compare an unchanged upstream
build with the profile under the same CPU, route, MTU, latency, loss, and
server bandwidth conditions.

The endpoint-selection core in `core/brain` accepts only operator-provided
endpoints and records bounded health state; it does not scan networks or
discover alternate ports. Health probes use a bounded worker pool so a large
configured endpoint set cannot create an unbounded probe burst.

## Explicit non-goals

This project does not implement traffic camouflage, protocol impersonation,
DPI evasion, port hopping, or mechanisms intended to bypass a campus,
operator, or other network's rate limits and blocking controls. No transport
protocol can guarantee immunity from detection, rate limiting, or blocking.

## Configuration

The client and server accept:

```yaml
profile: ultra
```

The client share URI preserves this setting with `profile=ultra`. Existing
configs with no `profile` keep the upstream behavior.

## Verification

```bash
go test ./extras/security ./app/cmd
go test ./core/...
cd benchmarks && go test -run '^$' -bench . -benchmem
```

## VPS download

The first release publishes Linux binaries for common VPS architectures:

```bash
arch=$(uname -m)
case "$arch" in
  x86_64) asset=gfw404-linux-amd64 ;;
  aarch64|arm64) asset=gfw404-linux-arm64 ;;
  *) echo "unsupported architecture: $arch" >&2; exit 1 ;;
esac
base=https://github.com/chengamki-tech/gfw404/releases/download/v0.1.0
curl -fL "$base/$asset" -o /tmp/gfw404
curl -fL "$base/SHA256SUMS" -o /tmp/gfw404.SHA256SUMS
(cd /tmp && grep " $asset$" gfw404.SHA256SUMS | sha256sum -c -)
install -m 0755 /tmp/gfw404 /usr/local/bin/gfw404
```

Only download releases from a repository and tag you trust, and verify the
checksum before installation.

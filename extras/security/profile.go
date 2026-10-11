// Package security contains opt-in security profiles for the application layer.
package security

import (
	"fmt"
	"strings"
	"time"
)

// Profile selects application-level security defaults. The empty profile keeps
// the upstream configuration behavior for compatibility.
type Profile string

const (
	ProfileDefault      Profile = "default"
	ProfileUltra        Profile = "ultra"
	MinCredentialLength         = 16
)

// NormalizeProfile validates and canonicalizes a profile name.
func NormalizeProfile(value string) (Profile, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", string(ProfileDefault):
		return ProfileDefault, nil
	case string(ProfileUltra):
		return ProfileUltra, nil
	default:
		return "", fmt.Errorf("unsupported security profile %q", value)
	}
}

// ValidateClientTLS enforces the additional certificate-verification rule for
// Ultra. A pin is the only trust anchor allowed when normal verification is
// explicitly disabled; a custom CA is used with normal verification.
func ValidateClientTLS(profile Profile, insecure, pinned, customCA bool) error {
	if profile == ProfileUltra && insecure && customCA {
		return fmt.Errorf("ultra profile does not allow tls.insecure with tls.ca")
	}
	if profile == ProfileUltra && insecure && !pinned {
		return fmt.Errorf("ultra profile requires certificate pinning when tls.insecure is enabled")
	}
	return nil
}

// ValidateUltraQUIC validates explicit QUIC settings for the Ultra profile.
// Zero values remain valid because the core applies its normal defaults later.
func ValidateUltraQUIC(profile Profile, streamInit, streamMax, connInit, connMax uint64, idle time.Duration, streams int64, disablePMTU, disableStatelessReset bool) error {
	if profile != ProfileUltra {
		return nil
	}
	if disablePMTU {
		return fmt.Errorf("ultra profile requires path MTU discovery")
	}
	if disableStatelessReset {
		return fmt.Errorf("ultra profile requires stateless reset support")
	}
	if err := validateWindow(streamInit, 256<<20); err != nil {
		return fmt.Errorf("invalid stream receive window: %w", err)
	}
	if err := validateWindow(streamMax, 256<<20); err != nil {
		return fmt.Errorf("invalid stream receive window: %w", err)
	}
	if err := validateWindow(connInit, 1<<30); err != nil {
		return fmt.Errorf("invalid connection receive window: %w", err)
	}
	if err := validateWindow(connMax, 1<<30); err != nil {
		return fmt.Errorf("invalid connection receive window: %w", err)
	}
	if streamInit != 0 && streamMax != 0 && streamInit > streamMax {
		return fmt.Errorf("initial stream receive window cannot exceed maximum")
	}
	if connInit != 0 && connMax != 0 && connInit > connMax {
		return fmt.Errorf("initial connection receive window cannot exceed maximum")
	}
	if idle != 0 && (idle < 4*time.Second || idle > 120*time.Second) {
		return fmt.Errorf("idle timeout must be between 4s and 2m0s")
	}
	if streams != 0 && (streams < 8 || streams > 65536) {
		return fmt.Errorf("incoming streams must be between 8 and 65536")
	}
	return nil
}

func validateWindow(value, max uint64) error {
	if value != 0 && (value < 16384 || value > max) {
		return fmt.Errorf("must be between 16384 and %d", max)
	}
	return nil
}

// ValidateCertificatePin checks the format of a pin used by the Ultra profile.
// The value is the lowercase hexadecimal SHA-256 digest of the DER leaf cert;
// separators accepted by the CLI are ignored before validation.
func ValidateCertificatePin(profile Profile, value string) error {
	if profile != ProfileUltra || value == "" {
		return nil
	}
	normalized := strings.NewReplacer(":", "", "-", "").Replace(strings.ToLower(value))
	if len(normalized) != 64 {
		return fmt.Errorf("ultra profile requires a 64-character SHA-256 certificate pin")
	}
	for _, r := range normalized {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return fmt.Errorf("ultra profile requires a hexadecimal SHA-256 certificate pin")
		}
	}
	return nil
}

// ValidateCredential applies the Ultra minimum length to credentials while
// keeping legacy profiles compatible with existing deployments.
func ValidateCredential(profile Profile, value string) error {
	if profile == ProfileUltra && len(value) < MinCredentialLength {
		return fmt.Errorf("ultra profile requires credentials of at least %d characters", MinCredentialLength)
	}
	return nil
}

const (
	UltraAuthFailureWindow  = time.Minute
	UltraAuthFailureLimit   = 3
	UltraAuthFailureEntries = 4096
)

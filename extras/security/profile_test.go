package security

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNormalizeProfile(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want Profile
	}{
		{name: "empty keeps legacy behavior", in: "", want: ProfileDefault},
		{name: "default", in: "default", want: ProfileDefault},
		{name: "case and whitespace", in: "  ULTRA ", want: ProfileUltra},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeProfile(tt.in)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNormalizeProfileRejectsUnknown(t *testing.T) {
	_, err := NormalizeProfile("stealth")
	require.EqualError(t, err, `unsupported security profile "stealth"`)
}

func TestValidateClientTLS(t *testing.T) {
	require.NoError(t, ValidateClientTLS(ProfileDefault, true, false, false))
	require.NoError(t, ValidateClientTLS(ProfileUltra, false, false, false))
	require.NoError(t, ValidateClientTLS(ProfileUltra, true, true, false))
	require.EqualError(t,
		ValidateClientTLS(ProfileUltra, true, false, true),
		"ultra profile does not allow tls.insecure with tls.ca",
	)
	require.EqualError(t,
		ValidateClientTLS(ProfileUltra, true, false, false),
		"ultra profile requires certificate pinning when tls.insecure is enabled",
	)
}

func TestValidateCertificatePin(t *testing.T) {
	require.NoError(t, ValidateCertificatePin(ProfileDefault, "deadbeef"))
	require.NoError(t, ValidateCertificatePin(ProfileUltra, "00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff:00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff"))
	require.EqualError(t, ValidateCertificatePin(ProfileUltra, "deadbeef"), "ultra profile requires a 64-character SHA-256 certificate pin")
	require.EqualError(t, ValidateCertificatePin(ProfileUltra, "zz"+"00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"[:62]), "ultra profile requires a hexadecimal SHA-256 certificate pin")
}

func TestValidateUltraQUIC(t *testing.T) {
	require.NoError(t, ValidateUltraQUIC(ProfileDefault, 1, 1, 1, 1, time.Nanosecond, 1, true, true))
	require.NoError(t, ValidateUltraQUIC(ProfileUltra, 1<<20, 2<<20, 4<<20, 8<<20, 30*time.Second, 1024, false, false))
	require.EqualError(t, ValidateUltraQUIC(ProfileUltra, 0, 0, 0, 0, 0, 0, true, false), "ultra profile requires path MTU discovery")
	require.EqualError(t, ValidateUltraQUIC(ProfileUltra, 0, 0, 0, 0, 0, 0, false, true), "ultra profile requires stateless reset support")
	require.EqualError(t, ValidateUltraQUIC(ProfileUltra, 1, 1, 0, 0, 0, 0, false, false), "invalid stream receive window: must be between 16384 and 268435456")
	require.EqualError(t, ValidateUltraQUIC(ProfileUltra, 2<<20, 1<<20, 0, 0, 0, 0, false, false), "initial stream receive window cannot exceed maximum")
}

func TestValidateCredential(t *testing.T) {
	require.NoError(t, ValidateCredential(ProfileDefault, "short"))
	require.NoError(t, ValidateCredential(ProfileUltra, "0123456789abcdef"))
	require.EqualError(t, ValidateCredential(ProfileUltra, "short"), "ultra profile requires credentials of at least 16 characters")
}

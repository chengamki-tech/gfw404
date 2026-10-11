package server

import (
	"crypto/tls"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertToStdTLSConfigRequiresTLS13(t *testing.T) {
	got := convertToStdTLSConfig(&Config{})
	require.Equal(t, uint16(tls.VersionTLS13), got.MinVersion)
}

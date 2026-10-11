package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthAttemptLimit(t *testing.T) {
	require.True(t, shouldAttemptAuth(0))
	require.True(t, shouldAttemptAuth(maxAuthAttempts-1))
	require.False(t, shouldAttemptAuth(maxAuthAttempts))
}

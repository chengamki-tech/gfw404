package server

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAuthFailureLimiterUsesSourceIP(t *testing.T) {
	limiter := NewAuthFailureLimiter(time.Minute, 2, 16)
	first := &net.UDPAddr{IP: net.ParseIP("192.0.2.10"), Port: 1000}
	secondPort := &net.UDPAddr{IP: net.ParseIP("192.0.2.10"), Port: 2000}

	require.True(t, limiter.Allow(first))
	limiter.RecordFailure(first)
	require.True(t, limiter.Allow(secondPort))
	limiter.RecordFailure(secondPort)
	require.False(t, limiter.Allow(first), "source-port changes must not bypass the IP limit")
}

func TestAuthFailureLimiterResetsOnSuccess(t *testing.T) {
	limiter := NewAuthFailureLimiter(time.Minute, 1, 16)
	addr := &net.UDPAddr{IP: net.ParseIP("192.0.2.11"), Port: 1000}

	limiter.RecordFailure(addr)
	require.False(t, limiter.Allow(addr))
	limiter.RecordSuccess(addr)
	require.True(t, limiter.Allow(addr))
}

func TestAuthFailureLimiterExpires(t *testing.T) {
	limiter := NewAuthFailureLimiter(5*time.Millisecond, 1, 16)
	addr := &net.UDPAddr{IP: net.ParseIP("192.0.2.12"), Port: 1000}

	limiter.RecordFailure(addr)
	require.Eventually(t, func() bool { return limiter.Allow(addr) }, time.Second, time.Millisecond)
}

func TestAuthFailureLimiterBoundsEntries(t *testing.T) {
	limiter := NewAuthFailureLimiter(time.Minute, 1, 2)
	for i := 0; i < 10; i++ {
		addr := &net.UDPAddr{IP: net.ParseIP(fmt.Sprintf("192.0.2.%d", i+1)), Port: 1000}
		limiter.RecordFailure(addr)
	}
	require.LessOrEqual(t, limiter.EntryCount(), 2)
}

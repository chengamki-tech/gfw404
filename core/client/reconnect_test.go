package client

import (
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

type reconnectTestClient struct {
	closed bool
}

func (c *reconnectTestClient) TCP(string) (net.Conn, error) { return nil, nil }
func (c *reconnectTestClient) UDP() (HyUDPConn, error)      { return nil, nil }
func (c *reconnectTestClient) Close() error {
	c.closed = true
	return nil
}

func TestReconnectClearsClosedClientAfterFailure(t *testing.T) {
	old := &reconnectTestClient{}
	failure := errors.New("temporary handshake failure")
	rc := &reconnectableClientImpl{
		configFunc: func() (*Config, error) { return &Config{}, nil },
		newClientFunc: func(*Config) (Client, *HandshakeInfo, error) {
			return nil, nil, failure
		},
		client: old,
	}

	require.ErrorIs(t, rc.reconnect(), failure)
	require.True(t, old.closed)
	require.Nil(t, rc.client)
}

package brain

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func testEndpoints() []Endpoint {
	return []Endpoint{
		{ID: "a", Address: "a.example:443"},
		{ID: "b", Address: "b.example:443"},
	}
}

func TestSelectorPicksLowestRTT(t *testing.T) {
	s, err := NewSelector(testEndpoints(), Options{})
	require.NoError(t, err)
	now := time.Unix(100, 0)
	require.NoError(t, s.RecordSuccess("a", 80*time.Millisecond, now))
	require.NoError(t, s.RecordSuccess("b", 20*time.Millisecond, now))

	got, err := s.Pick(now)
	require.NoError(t, err)
	require.Equal(t, "b", got.ID)
}

func TestSelectorBackoffAndRecovery(t *testing.T) {
	base := 10 * time.Second
	s, err := NewSelector(testEndpoints(), Options{
		FailureThreshold: 1,
		BackoffBase:      base,
		BackoffMax:       30 * time.Second,
		Jitter:           func(d time.Duration) time.Duration { return d + time.Second },
	})
	require.NoError(t, err)
	now := time.Unix(100, 0)
	require.NoError(t, s.RecordFailure("a", now))

	got, err := s.Pick(now)
	require.NoError(t, err)
	require.Equal(t, "b", got.ID)
	got, err = s.Pick(now.Add(10*time.Second + 999*time.Millisecond))
	require.NoError(t, err)
	require.Equal(t, "b", got.ID)
	require.NoError(t, s.RecordSuccess("a", 15*time.Millisecond, now.Add(11*time.Second)))
	status := s.Snapshot(now.Add(11 * time.Second))
	require.Equal(t, uint(0), status[0].Failures)
}

func TestSelectorProbeAllRecordsResults(t *testing.T) {
	s, err := NewSelector(testEndpoints(), Options{})
	require.NoError(t, err)
	probe := ProbeFunc(func(ctx context.Context, endpoint Endpoint) (time.Duration, error) {
		if endpoint.ID == "b" {
			return 0, errors.New("unreachable")
		}
		return 5 * time.Millisecond, nil
	})

	summary := s.ProbeAll(context.Background(), probe, time.Unix(200, 0))
	require.Equal(t, 2, summary.Checked)
	require.Equal(t, 1, summary.Healthy)
	got, err := s.Pick(time.Unix(200, 0))
	require.NoError(t, err)
	require.Equal(t, "a", got.ID)
}

func TestSelectorRejectsInvalidEndpoints(t *testing.T) {
	_, err := NewSelector([]Endpoint{{ID: "", Address: "a"}}, Options{})
	require.EqualError(t, err, "endpoint id is required")
	_, err = NewSelector([]Endpoint{{ID: "a", Address: ""}}, Options{})
	require.EqualError(t, err, "endpoint address is required")
	_, err = NewSelector([]Endpoint{{ID: "a", Address: "a"}, {ID: "a", Address: "b"}}, Options{})
	require.EqualError(t, err, "duplicate endpoint id \"a\"")
}

func TestSelectorUsesStableIDTieBreak(t *testing.T) {
	s, err := NewSelector([]Endpoint{
		{ID: "z", Address: "z.example:443"},
		{ID: "a", Address: "a.example:443"},
	}, Options{})
	require.NoError(t, err)
	got, err := s.Pick(time.Unix(100, 0))
	require.NoError(t, err)
	require.Equal(t, "a", got.ID)
}

func TestSelectorClampsJitter(t *testing.T) {
	s, err := NewSelector(testEndpoints(), Options{
		FailureThreshold: 1,
		BackoffBase:      time.Second,
		BackoffMax:       2 * time.Second,
		Jitter:           func(time.Duration) time.Duration { return 100 * time.Hour },
	})
	require.NoError(t, err)
	now := time.Unix(100, 0)
	require.NoError(t, s.RecordFailure("a", now))
	status := s.Snapshot(now.Add(2 * time.Second))
	var a EndpointStatus
	for _, item := range status {
		if item.ID == "a" {
			a = item
		}
	}
	require.Equal(t, now.Add(2*time.Second), a.RetryAt)
}

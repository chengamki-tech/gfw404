// Package brain provides bounded endpoint health selection for authorized
// deployments. It does not discover endpoints or alter transport packets.
package brain

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

var ErrNoHealthyEndpoints = errors.New("no healthy endpoints")

type Endpoint struct {
	ID      string
	Address string
}

type Options struct {
	FailureThreshold uint
	BackoffBase      time.Duration
	BackoffMax       time.Duration
	Jitter           func(time.Duration) time.Duration
}

type endpointState struct {
	Endpoint
	RTT      time.Duration
	Healthy  bool
	Failures uint
	RetryAt  time.Time
	LastSeen time.Time
}

type Selector struct {
	mu        sync.RWMutex
	options   Options
	endpoints map[string]*endpointState
}

type Probe interface {
	Probe(context.Context, Endpoint) (time.Duration, error)
}

type ProbeFunc func(context.Context, Endpoint) (time.Duration, error)

func (f ProbeFunc) Probe(ctx context.Context, endpoint Endpoint) (time.Duration, error) {
	return f(ctx, endpoint)
}

type ProbeSummary struct {
	Checked int
	Healthy int
}

func NewSelector(endpoints []Endpoint, options Options) (*Selector, error) {
	if options.FailureThreshold == 0 {
		options.FailureThreshold = 3
	}
	if options.BackoffBase <= 0 {
		options.BackoffBase = time.Second
	}
	if options.BackoffMax < options.BackoffBase {
		options.BackoffMax = 30 * time.Second
	}
	s := &Selector{options: options, endpoints: make(map[string]*endpointState, len(endpoints))}
	for _, endpoint := range endpoints {
		if endpoint.ID == "" {
			return nil, errors.New("endpoint id is required")
		}
		if endpoint.Address == "" {
			return nil, errors.New("endpoint address is required")
		}
		if _, ok := s.endpoints[endpoint.ID]; ok {
			return nil, fmt.Errorf("duplicate endpoint id %q", endpoint.ID)
		}
		s.endpoints[endpoint.ID] = &endpointState{Endpoint: endpoint, Healthy: true}
	}
	if len(s.endpoints) == 0 {
		return nil, errors.New("at least one endpoint is required")
	}
	return s, nil
}

func (s *Selector) Pick(now time.Time) (Endpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	states := make([]*endpointState, 0, len(s.endpoints))
	for _, state := range s.endpoints {
		if !state.Healthy && now.Before(state.RetryAt) {
			continue
		}
		states = append(states, state)
	}
	if len(states) == 0 {
		return Endpoint{}, ErrNoHealthyEndpoints
	}
	sort.SliceStable(states, func(i, j int) bool {
		left, right := selectionScore(states[i]), selectionScore(states[j])
		if left == right {
			return states[i].ID < states[j].ID
		}
		return left < right
	})
	return states[0].Endpoint, nil
}

func selectionScore(state *endpointState) time.Duration {
	if !state.Healthy {
		return time.Duration(math.MaxInt64)
	}
	if state.RTT <= 0 {
		return time.Duration(math.MaxInt64 - 1)
	}
	return state.RTT
}

func (s *Selector) RecordSuccess(id string, rtt time.Duration, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.endpoints[id]
	if !ok {
		return fmt.Errorf("unknown endpoint %q", id)
	}
	state.Healthy = true
	state.Failures = 0
	state.RetryAt = time.Time{}
	state.RTT = rtt
	state.LastSeen = now
	return nil
}

func (s *Selector) RecordFailure(id string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.endpoints[id]
	if !ok {
		return fmt.Errorf("unknown endpoint %q", id)
	}
	state.Failures++
	state.LastSeen = now
	if state.Failures >= s.options.FailureThreshold {
		state.Healthy = false
		backoff := s.options.BackoffBase
		for i := uint(1); i < state.Failures; i++ {
			if backoff >= s.options.BackoffMax/2 {
				backoff = s.options.BackoffMax
				break
			}
			backoff *= 2
		}
		if backoff > s.options.BackoffMax {
			backoff = s.options.BackoffMax
		}
		if s.options.Jitter != nil {
			backoff = s.options.Jitter(backoff)
		}
		if backoff < 0 {
			backoff = 0
		} else if backoff > s.options.BackoffMax {
			backoff = s.options.BackoffMax
		}
		state.RetryAt = now.Add(backoff)
	}
	return nil
}

func (s *Selector) ProbeAll(ctx context.Context, probe Probe, now time.Time) ProbeSummary {
	summary := ProbeSummary{}
	s.mu.RLock()
	endpoints := make([]Endpoint, 0, len(s.endpoints))
	for _, state := range s.endpoints {
		endpoints = append(endpoints, state.Endpoint)
	}
	s.mu.RUnlock()
	for _, endpoint := range endpoints {
		summary.Checked++
		rtt, err := probe.Probe(ctx, endpoint)
		if err != nil {
			_ = s.RecordFailure(endpoint.ID, now)
			continue
		}
		summary.Healthy++
		_ = s.RecordSuccess(endpoint.ID, rtt, now)
	}
	return summary
}

type EndpointStatus struct {
	Endpoint
	RTT      time.Duration
	Healthy  bool
	Failures uint
	RetryAt  time.Time
}

func (s *Selector) Snapshot(now time.Time) []EndpointStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]EndpointStatus, 0, len(s.endpoints))
	for _, state := range s.endpoints {
		result = append(result, EndpointStatus{
			Endpoint: state.Endpoint,
			RTT:      state.RTT,
			Healthy:  state.Healthy || now.After(state.RetryAt),
			Failures: state.Failures,
			RetryAt:  state.RetryAt,
		})
	}
	return result
}

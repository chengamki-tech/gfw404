package server

import (
	"net"
	"strings"
	"sync"
	"time"
)

// AuthLimiter gates authentication attempts before expensive credential
// verification. Implementations must be safe for concurrent use.
type AuthLimiter interface {
	Allow(net.Addr) bool
	RecordFailure(net.Addr)
	RecordSuccess(net.Addr)
}

type authFailureEntry struct {
	failures uint8
	resetAt  time.Time
	lastSeen time.Time
}

// AuthFailureLimiter limits failed authentication attempts per source IP.
// Entries are bounded so unauthenticated traffic cannot grow memory without
// limit.
type AuthFailureLimiter struct {
	mu          sync.Mutex
	window      time.Duration
	maxFailures uint8
	maxEntries  int
	entries     map[string]authFailureEntry
}

func NewAuthFailureLimiter(window time.Duration, maxFailures uint8, maxEntries int) *AuthFailureLimiter {
	if window <= 0 {
		window = time.Minute
	}
	if maxFailures == 0 {
		maxFailures = 3
	}
	if maxEntries <= 0 {
		maxEntries = 4096
	}
	return &AuthFailureLimiter{
		window:      window,
		maxFailures: maxFailures,
		maxEntries:  maxEntries,
		entries:     make(map[string]authFailureEntry),
	}
}

func (l *AuthFailureLimiter) Allow(addr net.Addr) bool {
	key := authSourceKey(addr)
	if key == "" {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	l.purgeExpiredLocked(now)
	entry, ok := l.entries[key]
	return !ok || now.After(entry.resetAt) || entry.failures < l.maxFailures
}

func (l *AuthFailureLimiter) RecordFailure(addr net.Addr) {
	key := authSourceKey(addr)
	if key == "" {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	l.purgeExpiredLocked(now)
	entry, ok := l.entries[key]
	if !ok {
		if len(l.entries) >= l.maxEntries {
			l.evictOldestLocked()
		}
		entry = authFailureEntry{resetAt: now.Add(l.window)}
	} else if now.After(entry.resetAt) {
		entry.failures = 0
		entry.resetAt = now.Add(l.window)
	}
	if entry.failures < l.maxFailures {
		entry.failures++
	}
	entry.lastSeen = now
	l.entries[key] = entry
}

func (l *AuthFailureLimiter) RecordSuccess(addr net.Addr) {
	key := authSourceKey(addr)
	if key == "" {
		return
	}
	l.mu.Lock()
	delete(l.entries, key)
	l.mu.Unlock()
}

// EntryCount is intended for diagnostics and tests, not authorization logic.
func (l *AuthFailureLimiter) EntryCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.purgeExpiredLocked(time.Now())
	return len(l.entries)
}

func (l *AuthFailureLimiter) purgeExpiredLocked(now time.Time) {
	for key, entry := range l.entries {
		if now.After(entry.resetAt) {
			delete(l.entries, key)
		}
	}
}

func (l *AuthFailureLimiter) evictOldestLocked() {
	var oldestKey string
	var oldest time.Time
	for key, entry := range l.entries {
		if oldestKey == "" || entry.lastSeen.Before(oldest) {
			oldestKey = key
			oldest = entry.lastSeen
		}
	}
	if oldestKey != "" {
		delete(l.entries, oldestKey)
	}
}

func authSourceKey(addr net.Addr) string {
	if addr == nil {
		return ""
	}
	switch value := addr.(type) {
	case *net.UDPAddr:
		return value.IP.String()
	case *net.TCPAddr:
		return value.IP.String()
	}
	text := addr.String()
	if host, _, err := net.SplitHostPort(text); err == nil {
		return strings.Trim(host, "[]")
	}
	if strings.HasPrefix(text, "[") && strings.HasSuffix(text, "]") {
		return strings.Trim(text, "[]")
	}
	return text
}

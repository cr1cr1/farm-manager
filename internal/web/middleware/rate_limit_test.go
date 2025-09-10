package middleware

import (
	"testing"
	"time"
)

func TestTokenBucketAllowAndRefill(t *testing.T) {
	b := &tokenBucket{
		tokens:     2, // burst
		lastRefill: time.Now(),
		rps:        2, // 2 tokens per second
		burst:      2,
	}

	now := time.Now()

	// Consume available burst.
	if !b.allow(now) {
		t.Fatalf("expected allow first")
	}
	if !b.allow(now) {
		t.Fatalf("expected allow second")
	}

	// Now should be empty.
	if b.allow(now) {
		t.Fatalf("expected deny when empty")
	}

	// Advance time enough to refill one token.
	now = now.Add(600 * time.Millisecond)
	if !b.allow(now) {
		t.Fatalf("expected allow after partial refill")
	}

	// And again consume until empty.
	if b.allow(now) {
		// At most one token after partial refill, so this should deny.
		t.Fatalf("expected deny after consuming partial refill")
	}

	// Advance time to refill to burst.
	now = now.Add(2 * time.Second)
	if !b.allow(now) {
		t.Fatalf("expected first allow after full refill")
	}
	if !b.allow(now) {
		t.Fatalf("expected second allow after full refill")
	}
	if b.allow(now) {
		t.Fatalf("expected deny after consuming burst")
	}
}

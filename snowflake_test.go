package snowflake

import (
	"sync"
	"testing"
)

func TestGenerateUniqueConcurrent(t *testing.T) {
	const goroutines = 64
	const perGoroutine = 5000

	var wg sync.WaitGroup
	results := make([][]int64, goroutines)

	for g := range goroutines {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ids := make([]int64, perGoroutine)
			for i := range perGoroutine {
				ids[i] = Generate()
			}
			results[idx] = ids
		}(g)
	}
	wg.Wait()

	seen := make(map[int64]struct{}, goroutines*perGoroutine)
	for _, ids := range results {
		for _, id := range ids {
			if _, dup := seen[id]; dup {
				t.Fatalf("duplicate id generated: %d", id)
			}
			seen[id] = struct{}{}
		}
	}

	if got, want := len(seen), goroutines*perGoroutine; got != want {
		t.Fatalf("got %d unique ids, want %d", got, want)
	}
}

func TestGenerateMonotonicSerial(t *testing.T) {
	prev := Generate()
	for i := range 10000 {
		id := Generate()
		if id <= prev {
			t.Fatalf("id %d not greater than previous %d at iteration %d", id, prev, i)
		}
		prev = id
	}
}

func TestParseRoundTrip(t *testing.T) {
	id := Generate()
	ts, machine, seq := Parse(id)

	if machine != machineID {
		t.Fatalf("parsed machine id = %d, want %d", machine, machineID)
	}
	if seq < 0 || seq > maxSequence {
		t.Fatalf("parsed sequence %d out of range [0, %d]", seq, maxSequence)
	}
	if ts < epoch {
		t.Fatalf("parsed timestamp %d is before epoch %d", ts, epoch)
	}
}

func TestSequenceWrapWaitsForNextMillisecond(t *testing.T) {
	mu.Lock()
	lastTimestamp = currentTimestamp()
	sequence = maxSequence
	startTs := lastTimestamp
	mu.Unlock()

	id := Generate()
	gotTs, _, gotSeq := Parse(id)

	if gotTs <= startTs {
		t.Fatalf("expected timestamp to advance past %d, got %d", startTs, gotTs)
	}
	if gotSeq != 0 {
		t.Fatalf("expected sequence to reset to 0 after wrap, got %d", gotSeq)
	}
}

func TestClockBackwardsDoesNotPanic(t *testing.T) {
	mu.Lock()
	lastTimestamp = currentTimestamp() + 60_000
	future := lastTimestamp
	sequence = 0
	mu.Unlock()

	id := Generate()
	gotTs, _, _ := Parse(id)

	if gotTs < future {
		t.Fatalf("expected timestamp to be clamped to last (%d), got %d", future, gotTs)
	}

	mu.Lock()
	lastTimestamp = 0
	sequence = 0
	mu.Unlock()
}

func BenchmarkGenerateSerial(b *testing.B) {
	for b.Loop() {
		_ = Generate()
	}
}

func BenchmarkGenerateParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = Generate()
		}
	})
}

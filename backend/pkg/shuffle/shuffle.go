// Package shuffle provides utilities for randomizing option order.
// This is a core part of the anti-cheat mechanism: every exam session
// sees options in a different order, so memorizing "A, B, C" is useless.
//
// We use Fisher-Yates (also known as Knuth shuffle), which produces
// a uniformly random permutation in O(n) time with O(1) extra space.
package shuffle

import (
	"math/rand"
	"time"
)

// rng is the package-level random source, seeded once at init.
// Using a local source rather than the global math/rand functions
// avoids lock contention and makes the shuffle reproducible in tests
// (by calling ResetSeed in test setup).
var rng *rand.Rand

func init() {
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// ResetSeed reinitializes the random source with a fixed seed.
// Use this only in tests to get deterministic output.
func ResetSeed(seed int64) {
	rng = rand.New(rand.NewSource(seed))
}

// Options shuffles a slice of uint (representing option IDs) in place
// using the Fisher-Yates algorithm and returns the slice for convenience.
//
// Algorithm: iterate from the end, swapping each element with a random
// element at or before its position. This guarantees every permutation
// is equally likely.
func Options(ids []uint) []uint {
	for i := len(ids) - 1; i > 0; i-- {
		j := rng.Intn(i + 1) // random index in [0, i]
		ids[i], ids[j] = ids[j], ids[i]
	}
	return ids
}

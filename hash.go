package synx

import "unsafe"

// MultiProbingHash returns a number of hashes from a given key.
// It must return at least one hash.
// Returning more than one allows for performance improvements
// via multi-probing.
//
// Ben Appleton, Michael O’Reilly, 2015, Multi-probe consistent hashing, Google
// http://arxiv.org/pdf/1505.00062.pdf
//
func MultiProbingHash(key string) [4]uint64 {
	p := *((*int64)(unsafe.Pointer(&key)))
	return [4]uint64{
		uint64(0xff & p),
		uint64(0xff & (p >> 8)),
		uint64(0xff & (p >> 16)),
		uint64(0xff & (p >> 24))}
}

// JumpConsistentHash returns index of map entry.
//
// Conceptually, it is a hash algorithm that can be expressed in about 5 lines of code.
// In comparison to the algorithm of Karger et al., jump consistent hash requires no storage,
// is faster, and does a better job of evenly dividing the key space among the buckets
// and of evenly dividing the workload when the number of buckets changes.
//
// Reference:
//
// John Lamping, Eric Veach, 2014, A Fast, Minimal Memory, Consistent Hash Algorithm, Google
// https://arxiv.org/pdf/1406.2294v1.pdf
//
func JumpConsistentHash(n int64, hashkey uint64) int64 {
	var b, j int64
	if n <= 0 {
		n = 1
	}

	for j < n {
		b = j
		hashkey = hashkey*2862933555777941757 + 1
		j = int64(float64(b+1) * (float64(int64(1)<<31) / float64((hashkey>>33)+1)))
	}
	return b
}

func hashChain(size int64, key string) []int64 {
	hashes := MultiProbingHash(key)
	idxs := make([]int64, len(hashes))
	for i, hash := range hashes {
		idxs[i] = JumpConsistentHash(size, hash)
	}
	return idxs
}

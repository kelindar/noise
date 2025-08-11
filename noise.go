package noise

import (
	"math"
	"math/bits"
)

// Number constraint for generic noise functions
type Number interface {
	~float32 | ~float64 | ~uint16 | ~uint32 | ~uint64 | ~int16 | ~int32 | ~int64
}

// ---------------------------------- White Noise ----------------------------------

// xxhash64 implements unrolled xxhash that produces same output as xxh3
// Source: https://github.com/zeebo/xxh3
func xxhash64(v, seed uint64) uint64 {
	x := v ^ (0x1cad21f72c81017c ^ 0xdb979083e96dd4de) + seed
	x ^= bits.RotateLeft64(x, 49) ^ bits.RotateLeft64(x, 24)
	x *= 0x9fb21c651e98df25
	x ^= (x >> 35) + 4
	x *= 0x9fb21c651e98df25
	x ^= (x >> 28)
	return x
}

// coordToUint64 converts a coordinate to uint64 for hashing (no allocations)
func coordToUint64[T Number](coord T) uint64 {
	switch any(coord).(type) {
	case float32:
		return uint64(math.Float32bits(float32(coord)))
	case float64:
		return math.Float64bits(float64(coord))
	case uint16:
		return uint64(coord)
	case uint32:
		return uint64(coord)
	case uint64:
		return uint64(coord)
	case int16:
		return uint64(uint16(coord))
	case int32:
		return uint64(uint32(coord))
	case int64:
		return uint64(coord)
	default:
		panic("unsupported coordinate type")
	}
}

// White generates deterministic white noise in [-1, 1] range based on coordinates
func White[T Number](seed uint32, coords ...T) float32 {
	const mix uint64 = 0x9e3779b97f4a7c15

	hash := uint64(seed)
	switch len(coords) {
	case 0:
		panic("noise: requires at least 1 coordinate")
	case 1:
		hash = xxhash64(coordToUint64(coords[0]), hash)
	case 2:
		hash = xxhash64(coordToUint64(coords[0]), hash)
		hash = xxhash64(coordToUint64(coords[1]), hash+mix)
	default:
		for i, coord := range coords {
			coordBits := coordToUint64(coord)
			hash = xxhash64(coordBits, hash+uint64(i)*mix)
		}
	}

	return float32(hash>>32)/float32(1<<31) - 1.0
}

// ---------------------------------- Random ----------------------------------

// Float32 returns a deterministic float32 in [0.0, 1.0) based on x
func Float32(seed uint32, x uint64) float32 {
	hash := xxhash64(x, uint64(seed))
	return float32(hash>>32) / float32(1<<32)
}

// Float64 returns a deterministic float64 in [0.0, 1.0) based on x
func Float64(seed uint32, x uint64) float64 {
	hash := xxhash64(x, uint64(seed))
	return float64(hash) / float64(1<<64)
}

// Uint32 returns a deterministic uint32 based on x
func Uint32(seed uint32, x uint64) uint32 {
	hash := xxhash64(x, uint64(seed))
	return uint32(hash >> 32)
}

// Uint64 returns a deterministic uint64 based on x
func Uint64(seed uint32, x uint64) uint64 {
	return xxhash64(x, uint64(seed))
}

// IntN returns a deterministic int in [0, n) based on x
func IntN(seed uint32, n int, x uint64) int {
	if n <= 0 {
		panic("invalid argument to IntN")
	}
	hash := xxhash64(x, uint64(seed))
	return int(hash % uint64(n))
}

// Norm64 returns a deterministic normally distributed float64 based on x
func Norm64(seed uint32, x uint64) float64 {
	hash1 := xxhash64(x, uint64(seed))
	hash2 := xxhash64(x, uint64(seed)+0x9e3779b97f4a7c15)

	// Convert to [0,1) range
	u1 := float64(hash1) / float64(1<<64)
	u2 := float64(hash2) / float64(1<<64)

	// Box-Muller transform
	return math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
}

// Norm32 returns a deterministic normally distributed float32 based on x
func Norm32(seed uint32, x uint64) float32 {
	return float32(Norm64(seed, x))
}

// UintN returns a deterministic uint in [0, n) based on x
func UintN(seed uint32, n uint, x uint64) uint {
	if n == 0 {
		panic("invalid argument to UintN")
	}

	hash := xxhash64(x, uint64(seed))
	return uint(hash % uint64(n))
}

// Int32 returns a deterministic int32 based on x
func Int32(seed uint32, x uint64) int32 {
	hash := xxhash64(x, uint64(seed))
	return int32(hash >> 32)
}

// Int64 returns a deterministic int64 based on x
func Int64(seed uint32, x uint64) int64 {
	hash := xxhash64(x, uint64(seed))
	return int64(hash)
}

// Int returns a deterministic int based on x
func Int(seed uint32, x uint64) int {
	hash := xxhash64(x, uint64(seed))
	return int(hash)
}

// Uint returns a deterministic uint based on x
func Uint(seed uint32, x uint64) uint {
	hash := xxhash64(x, uint64(seed))
	return uint(hash)
}

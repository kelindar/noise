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

// hashCoords combines multiple coordinates into a single hash (no allocations)
func hashCoords[T Number](seed uint32, coords ...T) uint64 {
	if len(coords) == 0 {
		panic("noise: requires at least 1 coordinate")
	}

	// Start with seed
	hash := uint64(seed)

	// Combine coordinates using xxhash64
	for i, coord := range coords {
		coordBits := coordToUint64(coord)
		// Mix in coordinate index to ensure different positions for same values
		hash = xxhash64(coordBits, hash+uint64(i)*0x9e3779b97f4a7c15)
	}

	return hash
}

// Float32 returns a deterministic float32 in [0.0, 1.0) based on coordinates
func Float32[T Number](seed uint32, coords ...T) float32 {
	hash := hashCoords(seed, coords...)
	// Use upper 32 bits and convert to [0, 1) range
	return float32(hash>>32) / float32(1<<32)
}

// Float64 returns a deterministic float64 in [0.0, 1.0) based on coordinates
func Float64[T Number](seed uint32, coords ...T) float64 {
	hash := hashCoords(seed, coords...)
	// Use full 64 bits and convert to [0, 1) range
	return float64(hash) / float64(1<<64)
}

// Uint32 returns a deterministic uint32 based on coordinates
func Uint32[T Number](seed uint32, coords ...T) uint32 {
	hash := hashCoords(seed, coords...)
	return uint32(hash >> 32)
}

// Uint64 returns a deterministic uint64 based on coordinates
func Uint64[T Number](seed uint32, coords ...T) uint64 {
	return hashCoords(seed, coords...)
}

// IntN returns a deterministic int in [0, n) based on coordinates
func IntN[T Number](seed uint32, n int, coords ...T) int {
	if n <= 0 {
		panic("invalid argument to IntN")
	}
	hash := hashCoords(seed, coords...)
	return int(hash % uint64(n))
}

// Norm64 returns a deterministic normally distributed float64 based on coordinates
func Norm64[T Number](seed uint32, coords ...T) float64 {
	// Use Box-Muller transform with two hash values
	hash1 := hashCoords(seed, coords...)
	// Generate second hash by adding offset
	hash2 := xxhash64(hash1, uint64(seed)+0x9e3779b97f4a7c15)

	// Convert to [0,1) range
	u1 := float64(hash1) / float64(1<<64)
	u2 := float64(hash2) / float64(1<<64)

	// Box-Muller transform
	return math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
}

// Norm32 returns a deterministic normally distributed float32 based on coordinates
func Norm32[T Number](seed uint32, coords ...T) float32 {
	return float32(Norm64(seed, coords...))
}

// UintN returns a deterministic uint in [0, n) based on coordinates
func UintN[T Number](seed uint32, n uint, coords ...T) uint {
	if n == 0 {
		panic("invalid argument to UintN")
	}
	hash := hashCoords(seed, coords...)
	return uint(hash % uint64(n))
}

// Int32 returns a deterministic int32 based on coordinates
func Int32[T Number](seed uint32, coords ...T) int32 {
	hash := hashCoords(seed, coords...)
	return int32(hash >> 32)
}

// Int64 returns a deterministic int64 based on coordinates
func Int64[T Number](seed uint32, coords ...T) int64 {
	hash := hashCoords(seed, coords...)
	return int64(hash)
}

// Int returns a deterministic int based on coordinates
func Int[T Number](seed uint32, coords ...T) int {
	hash := hashCoords(seed, coords...)
	return int(hash)
}

// Uint returns a deterministic uint based on coordinates
func Uint[T Number](seed uint32, coords ...T) uint {
	hash := hashCoords(seed, coords...)
	return uint(hash)
}

// White generates deterministic white noise in [-1, 1] range (for compatibility)
func White[T Number](seed uint32, coords ...T) float32 {
	return Float32(seed, coords...)*2 - 1
}

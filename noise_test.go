package noise

import (
	"image"
	"image/png"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhite(t *testing.T) {
	const seed = uint32(42)

	tests := []struct {
		name     string
		fixture  string
		generate func() interface{}
	}{
		{
			name:    "White2D",
			fixture: "fixtures/white2d.png",
			generate: func() interface{} {
				return generate2DNoiseImage(100, 100, 1.0, func(x, y float32) float32 {
					return White(seed, x, y)
				})
			},
		},
		{
			name:    "White1D",
			fixture: "fixtures/white1d.png",
			generate: func() interface{} {
				return generate1DNoiseImage(400, 100, 1.0, func(x float32) float32 {
					return White(seed, x)
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate the noise output
			actual := tt.generate()

			// Load reference from fixtures
			refFile, err := os.Open(tt.fixture)
			assert.NoError(t, err)
			defer refFile.Close()

			expected, err := png.Decode(refFile)
			assert.NoError(t, err)

			// Compare with reference
			compareImages(t, expected, actual.(image.Image), tt.name)
			t.Logf("%s matches reference: %s", tt.name, tt.fixture)
		})
	}
}

func TestRandomFunctions(t *testing.T) {
	const seed = uint32(42)
	const x = uint64(12345)

	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		// Range tests
		{"Float32 [0,1)", func(t *testing.T) {
			for i := 0; i < 100; i++ {
				v := Float32(seed, uint64(i))
				assert.True(t, v >= 0 && v < 1, "got %f", v)
			}
		}},
		{"Float64 [0,1)", func(t *testing.T) {
			for i := 0; i < 100; i++ {
				v := Float64(seed, uint64(i))
				assert.True(t, v >= 0 && v < 1, "got %f", v)
			}
		}},
		{"IntN [0,n)", func(t *testing.T) {
			for i := 0; i < 100; i++ {
				v := IntN(seed, 10, uint64(i))
				assert.True(t, v >= 0 && v < 10, "got %d", v)
			}
		}},
		{"Int32N [0,n)", func(t *testing.T) {
			for i := 0; i < 100; i++ {
				v := Int32N(seed, 50, uint64(i))
				assert.True(t, v >= 0 && v < 50, "got %d", v)
			}
		}},
		{"Int64N [0,n)", func(t *testing.T) {
			for i := 0; i < 100; i++ {
				v := Int64N(seed, 100, uint64(i))
				assert.True(t, v >= 0 && v < 100, "got %d", v)
			}
		}},
		{"Uint32N [0,n)", func(t *testing.T) {
			for i := 0; i < 100; i++ {
				v := Uint32N(seed, 25, uint64(i))
				assert.True(t, v < 25, "got %d", v)
			}
		}},
		{"Uint64N [0,n)", func(t *testing.T) {
			for i := 0; i < 100; i++ {
				v := Uint64N(seed, 1000, uint64(i))
				assert.True(t, v < 1000, "got %d", v)
			}
		}},
		{"UintN [0,n)", func(t *testing.T) {
			for i := 0; i < 100; i++ {
				v := UintN(seed, 75, uint64(i))
				assert.True(t, v < 75, "got %d", v)
			}
		}},

		// In range tests
		{"IntIn [a,b]", func(t *testing.T) {
			for i := 0; i < 100; i++ {
				v := IntIn(seed, 10, 20, uint64(i))
				assert.True(t, v >= 10 && v <= 20, "got %d", v)
			}
		}},
		{"Int32In [a,b]", func(t *testing.T) {
			for i := 0; i < 100; i++ {
				v := Int32In(seed, -5, 5, uint64(i))
				assert.True(t, v >= -5 && v <= 5, "got %d", v)
			}
		}},
		{"Int64In [a,b]", func(t *testing.T) {
			for i := 0; i < 100; i++ {
				v := Int64In(seed, 100, 200, uint64(i))
				assert.True(t, v >= 100 && v <= 200, "got %d", v)
			}
		}},
		{"UintIn [a,b]", func(t *testing.T) {
			for i := 0; i < 100; i++ {
				v := UintIn(seed, 5, 15, uint64(i))
				assert.True(t, v >= 5 && v <= 15, "got %d", v)
			}
		}},
		{"Uint32In [a,b]", func(t *testing.T) {
			for i := 0; i < 100; i++ {
				v := Uint32In(seed, 100, 200, uint64(i))
				assert.True(t, v >= 100 && v <= 200, "got %d", v)
			}
		}},
		{"Uint64In [a,b]", func(t *testing.T) {
			for i := 0; i < 100; i++ {
				v := Uint64In(seed, 1000, 2000, uint64(i))
				assert.True(t, v >= 1000 && v <= 2000, "got %d", v)
			}
		}},

		// Missing function coverage
		{"Int64", func(t *testing.T) {
			v := Int64(seed, x)
			assert.NotEqual(t, int64(0), v) // Just ensure it returns something
		}},
		{"Uint", func(t *testing.T) {
			v := Uint(seed, x)
			assert.NotEqual(t, uint(0), v) // Just ensure it returns something
		}},
		{"Uint32", func(t *testing.T) {
			v := Uint32(seed, x)
			assert.NotEqual(t, uint32(0), v) // Just ensure it returns something
		}},

		// Deterministic tests
		{"Deterministic", func(t *testing.T) {
			assert.Equal(t, Float32(seed, x), Float32(seed, x))
			assert.Equal(t, Int32(seed, x), Int32(seed, x))
			assert.Equal(t, Uint64(seed, x), Uint64(seed, x))
			assert.Equal(t, Int64(seed, x), Int64(seed, x))
			assert.Equal(t, Uint(seed, x), Uint(seed, x))
			assert.Equal(t, Uint32(seed, x), Uint32(seed, x))
		}},
		{"Different inputs", func(t *testing.T) {
			assert.NotEqual(t, Float32(seed, x), Float32(seed, x+1))
			assert.NotEqual(t, Int(seed, x), Int(seed, x+1))
			assert.NotEqual(t, Int64(seed, x), Int64(seed, x+1))
			assert.NotEqual(t, Uint(seed, x), Uint(seed, x+1))
		}},

		// Normal distribution
		{"Norm64 reasonable", func(t *testing.T) {
			v := Norm64(seed, x)
			assert.True(t, v >= -5 && v <= 5, "got %f", v)
		}},
		{"Norm32 reasonable", func(t *testing.T) {
			v := Norm32(seed, x)
			assert.True(t, v >= -5 && v <= 5, "got %f", v)
		}},

		// Probability tests
		{"Roll32 probability", func(t *testing.T) {
			count := 0
			for i := 0; i < 1000; i++ {
				if Roll32(seed, 0.3, uint64(i)) {
					count++
				}
			}
			assert.True(t, count > 250 && count < 350, "got %d/1000", count)
		}},
		{"Roll64 probability", func(t *testing.T) {
			count := 0
			for i := 0; i < 1000; i++ {
				if Roll64(seed, 0.7, uint64(i)) {
					count++
				}
			}
			assert.True(t, count > 650 && count < 750, "got %d/1000", count)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func TestWhiteCoordinates(t *testing.T) {
	const seed = uint32(42)

	// Test White with different coordinate counts
	v1 := White(seed, 1.0)
	v2 := White(seed, 1.0, 2.0)
	v3 := White(seed, 1.0, 2.0, 3.0)
	v4 := White(seed, 1.0, 2.0, 3.0, 4.0)

	// All should be in [-1, 1] range
	assert.True(t, v1 >= -1 && v1 <= 1)
	assert.True(t, v2 >= -1 && v2 <= 1)
	assert.True(t, v3 >= -1 && v3 <= 1)
	assert.True(t, v4 >= -1 && v4 <= 1)

	// Different coordinate counts should give different values
	assert.NotEqual(t, v1, v2)
	assert.NotEqual(t, v2, v3)
	assert.NotEqual(t, v3, v4)

	// Test with different coordinate types (covers coordToUint64)
	vFloat32 := White(seed, float32(1.5))
	vFloat64 := White(seed, float64(1.5))
	vInt16 := White(seed, int16(1))
	vInt32 := White(seed, int32(1))
	vInt64 := White(seed, int64(1))
	vUint16 := White(seed, uint16(1))
	vUint32 := White(seed, uint32(1))
	vUint64 := White(seed, uint64(1))

	// Test with 5+ coordinates to cover the default case in White
	v5Plus := White(seed, 1.0, 2.0, 3.0, 4.0, 5.0)

	// All should be valid
	assert.True(t, vFloat32 >= -1 && vFloat32 <= 1)
	assert.True(t, v5Plus >= -1 && v5Plus <= 1)
	assert.True(t, vFloat64 >= -1 && vFloat64 <= 1)
	assert.True(t, vInt16 >= -1 && vInt16 <= 1)
	assert.True(t, vInt32 >= -1 && vInt32 <= 1)
	assert.True(t, vInt64 >= -1 && vInt64 <= 1)
	assert.True(t, vUint16 >= -1 && vUint16 <= 1)
	assert.True(t, vUint32 >= -1 && vUint32 <= 1)
	assert.True(t, vUint64 >= -1 && vUint64 <= 1)
}

func TestPanicCases(t *testing.T) {
	const seed = uint32(42)
	const x = uint64(12345)

	// Test N function panics
	assert.Panics(t, func() { IntN(seed, 0, x) })
	assert.Panics(t, func() { Int32N(seed, 0, x) })
	assert.Panics(t, func() { Int32N(seed, -1, x) })
	assert.Panics(t, func() { Int64N(seed, 0, x) })
	assert.Panics(t, func() { Int64N(seed, -1, x) })
	assert.Panics(t, func() { Uint32N(seed, 0, x) })
	assert.Panics(t, func() { Uint64N(seed, 0, x) })
	assert.Panics(t, func() { UintN(seed, 0, x) })

	// Test In function panics
	assert.Panics(t, func() { IntIn(seed, 10, 5, x) })
	assert.Panics(t, func() { UintIn(seed, 10, 5, x) })
	assert.Panics(t, func() { Int32In(seed, 10, 5, x) })
	assert.Panics(t, func() { Int64In(seed, 10, 5, x) })
	assert.Panics(t, func() { Uint32In(seed, 10, 5, x) })
	assert.Panics(t, func() { Uint64In(seed, 10, 5, x) })
}

package noise

import (
	"image"
	"image/png"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhite_Eval(t *testing.T) {
	const seed = uint32(42)

	// Test that white noise returns values in [-1,1]
	for i := 0; i < 100; i++ {
		v := White(seed, float32(i))
		assert.True(t, v >= -1 && v <= 1, "White noise should be in [-1,1], got %f", v)
	}

	// Test deterministic behavior
	v1 := White(seed, 1.5, 2.5)
	v2 := White(seed, 1.5, 2.5)
	assert.Equal(t, v1, v2, "White noise should be deterministic")

	// Test that different coordinates give different values (with high probability)
	v3 := White(seed, 3.5, 4.5)
	assert.NotEqual(t, v1, v3, "Different coordinates should give different values")
}

func TestWhite_Functions(t *testing.T) {
	const seed = uint32(42)

	// Test Float32 returns values in [0,1)
	for i := 0; i < 100; i++ {
		v := Float32(seed, float32(i))
		assert.True(t, v >= 0 && v < 1, "WhiteFloat32 should be in [0,1), got %f", v)
	}

	// Test Float64 returns values in [0,1)
	for i := 0; i < 100; i++ {
		v := Float64(seed, float32(i))
		assert.True(t, v >= 0 && v < 1, "WhiteFloat64 should be in [0,1), got %f", v)
	}

	// Test IntN returns values in [0,n)
	for i := 0; i < 100; i++ {
		v := IntN(seed, 10, float32(i))
		assert.True(t, v >= 0 && v < 10, "WhiteIntN should be in [0,10), got %d", v)
	}

	// Test deterministic behavior
	v1 := Float32(seed, 1.5, 2.5)
	v2 := Float32(seed, 1.5, 2.5)
	assert.Equal(t, v1, v2, "White functions should be deterministic")

	// Test that different coordinates give different values
	v3 := Float32(seed, 3.5, 4.5)
	assert.NotEqual(t, v1, v3, "Different coordinates should give different values")

	// Test normal distribution (basic sanity check)
	norm64 := Norm64(seed, 1.0, 2.0)
	assert.True(t, norm64 >= -5 && norm64 <= 5, "Norm64 should be reasonable, got %f", norm64)

	norm32 := Norm32(seed, 1.0, 2.0)
	assert.True(t, norm32 >= -5 && norm32 <= 5, "Norm32 should be reasonable, got %f", norm32)

	// Test UintN returns values in [0,n)
	for i := 0; i < 100; i++ {
		v := UintN(seed, 50, float32(i))
		assert.True(t, v < 50, "UintN should be in [0,50), got %d", v)
	}

	// Test Int32, Int64, Int, Uint
	i32 := Int32(seed, 1.0)
	i64 := Int64(seed, 1.0)
	i := Int(seed, 1.0)
	u := Uint(seed, 1.0)

	// Basic sanity checks (just ensure they don't panic and return different values)
	assert.NotEqual(t, Int32(seed, 2.0), i32, "Different coordinates should give different Int32")
	assert.NotEqual(t, Int64(seed, 2.0), i64, "Different coordinates should give different Int64")
	assert.NotEqual(t, Int(seed, 2.0), i, "Different coordinates should give different Int")
	assert.NotEqual(t, Uint(seed, 2.0), u, "Different coordinates should give different Uint")
}

func BenchmarkWhite(b *testing.B) {
	const seed = uint32(42)

	b.Run("Float32-1D", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Float32(seed, float32(i))
		}
	})

	b.Run("Float32-2D", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Float32(seed, float32(i), float32(i+1))
		}
	})

	b.Run("Float32-3D", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Float32(seed, float32(i), float32(i+1), float32(i+2))
		}
	})

	b.Run("IntN-1D", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			IntN(seed, 100, float32(i))
		}
	})

	b.Run("Uint64-2D", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Uint64(seed, float32(i), float32(i+1))
		}
	})

	b.Run("Norm32-1D", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Norm32(seed, float32(i))
		}
	})

	b.Run("Int32-2D", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Int32(seed, float32(i), float32(i+1))
		}
	})
}

func TestWhiteVisualRegression(t *testing.T) {
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

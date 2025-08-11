package noise

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimplex_500x500(t *testing.T) {
	n := 500
	freq := float32(25)
	s := NewSimplex(0)
	img := image.NewGray(image.Rect(0, 0, n, n))
	for x := 0; x < n; x++ {
		for y := 0; y < n; y++ {
			v := (1 + s.Eval(float32(x)/freq, float32(y)/freq)) / 2
			img.Set(x, y, color.Gray{
				Y: uint8(v * 255),
			})
		}
	}

	// Compare with the reference
	/*expect, err := os.Open("fixtures/500.png")
	assert.NoError(t, err)
	out, err := png.Decode(expect)
	assert.NoError(t, err)
	assert.Equal(t, out, img)*/

	f, err := os.Create("out.png")
	assert.NoError(t, err)
	assert.NoError(t, png.Encode(f, img))
}

func TestFloor(t *testing.T) {
	assert.Equal(t, int(math.Floor(1.5)), floor(1.5))
	assert.Equal(t, int(math.Floor(0.5)), floor(0.5))
	assert.Equal(t, int(math.Floor(-1.5)), floor(-1.5))
}

func TestSimplex_Eval(t *testing.T) {
	s := NewSimplex(42)

	// Test 1D noise
	v1 := s.Eval(1.5)
	assert.True(t, v1 >= -1 && v1 <= 1, "1D noise should be in [-1,1]")

	// Test 2D noise
	v2 := s.Eval(1.5, 2.5)
	assert.True(t, v2 >= -1 && v2 <= 1, "2D noise should be in [-1,1]")

	// Test 3D noise
	v3 := s.Eval(1.5, 2.5, 3.5)
	assert.True(t, v3 >= -1 && v3 <= 1, "3D noise should be in [-1,1]")

	// Test determinism
	assert.Equal(t, v2, s.Eval(1.5, 2.5), "Noise should be deterministic")
}

func TestFBM_Eval(t *testing.T) {
	f := NewFBM(42)

	// Test 2D FBM
	v := f.Eval(2.0, 0.5, 4, 1.5, 2.5) // 4 octaves, lacunarity=2.0, gain=0.5, x=1.5, y=2.5
	assert.True(t, v >= -1 && v <= 1, "FBM should be roughly in [-1,1]")

	// Test determinism
	assert.Equal(t, v, f.Eval(2.0, 0.5, 4, 1.5, 2.5), "FBM should be deterministic")

	// Test 3D FBM
	v3 := f.Eval(2.0, 0.5, 3, 1.0, 2.0, 3.0)
	assert.True(t, v3 >= -2 && v3 <= 2, "3D FBM should be reasonable")
}

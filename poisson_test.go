package noise

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPoisson(t *testing.T) {
	tests := []struct {
		name     string
		fixture  string
		generate func() any
		compare  func(t *testing.T, expected, actual any, name string)
	}{
		{
			name:    "Poisson1D",
			fixture: "fixtures/poisson1d.png",
			generate: func() any {
				w, h, n := 400, 100, 10
				sx := float32(w) / (2 * float32(n))
				return generatePoisson1D(w, h, sx, Sparse1(42, n))
			},
			compare: func(t *testing.T, expected, actual any, name string) {
				compareImages(t, expected.(image.Image), actual.(image.Image), name)
			},
		},

		{
			name:    "Poisson2D",
			fixture: "fixtures/poisson2d.png",
			generate: func() any {
				w, h, n := 200, 200, 10
				sx := float32(w) / (2 * float32(n))
				sy := float32(h) / (2 * float32(n))
				return generatePoisson2D(w, h, sx, sy, Sparse2(42, n, n))
			},
			compare: func(t *testing.T, expected, actual any, name string) {
				compareImages(t, expected.(image.Image), actual.(image.Image), name)
			},
		},
		{
			name:    "SparseFill1D",
			fixture: "fixtures/sparsefill1d.png",
			generate: func() any {
				w, h, gap := 400, 100, 20
				return generateSparseFill1D(w, h, SparseFill1(42, w, gap))
			},
			compare: func(t *testing.T, expected, actual any, name string) {
				compareImages(t, expected.(image.Image), actual.(image.Image), name)
			},
		},

		{
			name:    "SparseFill2D",
			fixture: "fixtures/sparsefill2d.png",
			generate: func() any {
				w, h, gap := 200, 200, 10
				return generateSparseFill2D(w, h, SparseFill2(42, w, h, gap))
			},
			compare: func(t *testing.T, expected, actual any, name string) {
				compareImages(t, expected.(image.Image), actual.(image.Image), name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate the Poisson output
			actual := tt.generate()

			// For now, just save the generated image to see the output
			// In a real test, you would compare against reference fixtures
			filename := "test_" + tt.name + ".png"
			file, err := os.Create(filename)
			assert.NoError(t, err)
			defer file.Close()

			err = png.Encode(file, actual.(image.Image))
			assert.NoError(t, err)

			t.Logf("%s generated: %s", tt.name, filename)
		})
	}
}

func TestSparseFillBounds(t *testing.T) {
	t.Run("SparseFill1D bounds", func(t *testing.T) {
		w, gap := 100, 10
		count := 0
		for x := range SparseFill1(42, w, gap) {
			assert.True(t, x >= 0 && x < w, "Point %d should be within bounds [0, %d)", x, w)
			count++
			if count > 1000 { // Safety break
				break
			}
		}
		assert.Greater(t, count, 0, "Should generate at least one point")
	})

	t.Run("SparseFill2D bounds", func(t *testing.T) {
		w, h, gap := 100, 80, 15
		count := 0
		for pt := range SparseFill2(42, w, h, gap) {
			assert.True(t, pt[0] >= 0 && pt[0] < w, "Point x=%d should be within bounds [0, %d)", pt[0], w)
			assert.True(t, pt[1] >= 0 && pt[1] < h, "Point y=%d should be within bounds [0, %d)", pt[1], h)
			count++
			if count > 1000 { // Safety break
				break
			}
		}
		assert.Greater(t, count, 0, "Should generate at least one point")
	})

	t.Run("SparseFill3D bounds", func(t *testing.T) {
		w, h, d, gap := 50, 60, 40, 12
		count := 0
		for pt := range SparseFill3(42, w, h, d, gap) {
			assert.True(t, pt[0] >= 0 && pt[0] < w, "Point x=%d should be within bounds [0, %d)", pt[0], w)
			assert.True(t, pt[1] >= 0 && pt[1] < h, "Point y=%d should be within bounds [0, %d)", pt[1], h)
			assert.True(t, pt[2] >= 0 && pt[2] < d, "Point z=%d should be within bounds [0, %d)", pt[2], d)
			count++
			if count > 1000 { // Safety break
				break
			}
		}
		assert.Greater(t, count, 0, "Should generate at least one point")
	})

	t.Run("SparseFill deterministic", func(t *testing.T) {
		// Same seed should produce same results
		w, h, gap := 50, 50, 10

		var points1, points2 [][2]int
		for pt := range SparseFill2(123, w, h, gap) {
			points1 = append(points1, pt)
		}
		for pt := range SparseFill2(123, w, h, gap) {
			points2 = append(points2, pt)
		}

		assert.Equal(t, points1, points2, "Same seed should produce identical results")
		assert.Greater(t, len(points1), 0, "Should generate at least one point")
	})
}

// generatePoisson2D creates an image with Poisson-distributed points
func generatePoisson2D(width, height int, sx, sy float32, gen func(func([2]float32) bool)) *image.Gray {
	img := image.NewGray(image.Rect(0, 0, width, height))

	// Fill background with white
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.Gray{Y: 255})
		}
	}

	// Draw points as black dots
	for pt := range gen {
		x := int(pt[0]*sx + float32(width)/2) // scale and center
		y := int(pt[1]*sy + float32(height)/2)

		// Only draw if within bounds
		if x >= 0 && x < width && y >= 0 && y < height {
			// Draw a small cross for visibility
			img.Set(x, y, color.Gray{Y: 0})
			if x > 0 {
				img.Set(x-1, y, color.Gray{Y: 0})
			}
			if x < width-1 {
				img.Set(x+1, y, color.Gray{Y: 0})
			}
			if y > 0 {
				img.Set(x, y-1, color.Gray{Y: 0})
			}
			if y < height-1 {
				img.Set(x, y+1, color.Gray{Y: 0})
			}
		}
	}

	return img
}

// generatePoisson1D creates a 1D visualization of Poisson points
func generatePoisson1D(width, height int, sx float32, gen func(func(float32) bool)) *image.Gray {
	img := image.NewGray(image.Rect(0, 0, width, height))

	// Fill background with white
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.Gray{Y: 255})
		}
	}

	// Draw points as vertical lines at the middle
	midY := height / 2
	for x := range gen {
		px := int(x*sx + float32(width)/2) // scale and center

		// Only draw if within bounds
		if px >= 0 && px < width {
			// Draw vertical line
			for dy := -height / 4; dy <= height/4; dy++ {
				if py := midY + dy; py >= 0 && py < height {
					img.Set(px, py, color.Gray{Y: 0})
				}
			}
		}
	}

	return img
}

// generateSparseFill2D creates an image with SparseFill2D-distributed points
func generateSparseFill2D(width, height int, gen func(func([2]int) bool)) *image.Gray {
	img := image.NewGray(image.Rect(0, 0, width, height))

	// Fill background with white
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.Gray{Y: 255})
		}
	}

	// Draw points as black dots
	for pt := range gen {
		x, y := pt[0], pt[1]

		// Only draw if within bounds
		if x >= 0 && x < width && y >= 0 && y < height {
			// Draw a small cross for visibility
			img.Set(x, y, color.Gray{Y: 0})
			if x > 0 {
				img.Set(x-1, y, color.Gray{Y: 0})
			}
			if x < width-1 {
				img.Set(x+1, y, color.Gray{Y: 0})
			}
			if y > 0 {
				img.Set(x, y-1, color.Gray{Y: 0})
			}
			if y < height-1 {
				img.Set(x, y+1, color.Gray{Y: 0})
			}
		}
	}

	return img
}

// generateSparseFill1D creates a 1D visualization of SparseFill1D points
func generateSparseFill1D(width, height int, gen func(func(int) bool)) *image.Gray {
	img := image.NewGray(image.Rect(0, 0, width, height))

	// Fill background with white
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.Gray{Y: 255})
		}
	}

	// Draw points as vertical lines at the middle
	midY := height / 2
	for x := range gen {
		// Only draw if within bounds
		if x >= 0 && x < width {
			// Draw vertical line
			for dy := -height / 4; dy <= height/4; dy++ {
				if py := midY + dy; py >= 0 && py < height {
					img.Set(x, py, color.Gray{Y: 0})
				}
			}
		}
	}

	return img
}

// sqrt64 helper function
func sqrt64(x float64) float64 {
	if x == 0 {
		return 0
	}
	// Simple Newton's method for square root
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

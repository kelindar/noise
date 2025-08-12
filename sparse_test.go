package noise

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSparse(t *testing.T) {
	tests := []struct {
		name     string
		fixture  string
		generate func() any
		compare  func(t *testing.T, expected, actual any, name string)
	}{
		{
			name:    "Sparse1D",
			fixture: "fixtures/sparse1d.png",
			generate: func() any {
				w, h, gap := 400, 100, 20
				return generateSparse1D(w, h, Sparse1(42, w, gap))
			},
			compare: func(t *testing.T, expected, actual any, name string) {
				compareImages(t, expected.(image.Image), actual.(image.Image), name)
			},
		},

		{
			name:    "Sparse2D",
			fixture: "fixtures/sparse2d.png",
			generate: func() any {
				w, h, gap := 200, 200, 5
				return generateSparse2D(w, h, Sparse2(42, w, h, gap))
			},
			compare: func(t *testing.T, expected, actual any, name string) {
				compareImages(t, expected.(image.Image), actual.(image.Image), name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.generate()

			// Save the generated image for reference
			/*outFile, err := os.Create("out_" + tt.name + ".png")
			assert.NoError(t, err)
			defer outFile.Close()
			png.Encode(outFile, actual.(image.Image))*/

			// Load reference from fixtures
			refFile, err := os.Open(tt.fixture)
			assert.NoError(t, err)
			defer refFile.Close()

			// Decode the reference image
			expected, err := png.Decode(refFile)
			assert.NoError(t, err)

			// Compare with reference
			tt.compare(t, expected, actual, tt.name)
			t.Logf("%s matches reference: %s", tt.name, tt.fixture)
		})
	}
}

// generateSparse2D creates an image with SparseFill2D-distributed points
func generateSparse2D(width, height int, gen func(func([2]int) bool)) *image.Gray {
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

// generateSparse1D creates a 1D visualization of SparseFill1D points
func generateSparse1D(width, height int, gen func(func(int) bool)) *image.Gray {
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

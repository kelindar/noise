package noise

import (
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimplex(t *testing.T) {
	s := NewSimplex(42)
	f := NewFBM(42)

	tests := []struct {
		name     string
		fixture  string
		generate func() any
		compare  func(t *testing.T, expected, actual any, name string)
	}{
		{
			name:    "FBM3D",
			fixture: "fixtures/fbm3d.gif",
			generate: func() any {
				return generate3DNoiseGIF(50, 50, 10, 0.1, func(x, y, z float32) float32 {
					return f.Eval(2.0, 0.5, 4, x, y, z)
				})
			},
			compare: func(t *testing.T, expected, actual any, name string) {
				compareGIFs(t, expected.(*gif.GIF), actual.(*gif.GIF), name)
			},
		},
		{
			name:    "Simplex3D",
			fixture: "fixtures/simplex3d.gif",
			generate: func() any {
				return generate3DNoiseGIF(50, 50, 10, 0.1, func(x, y, z float32) float32 {
					return s.Eval(x, y, z)
				})
			},
			compare: func(t *testing.T, expected, actual any, name string) {
				compareGIFs(t, expected.(*gif.GIF), actual.(*gif.GIF), name)
			},
		},
		{
			name:    "Simplex2D",
			fixture: "fixtures/simplex2d.png",
			generate: func() any {
				return generate2DNoiseImage(100, 100, 0.05, func(x, y float32) float32 {
					return s.Eval(x, y)
				})
			},
			compare: func(t *testing.T, expected, actual any, name string) {
				compareImages(t, expected.(image.Image), actual.(image.Image), name)
			},
		},
		{
			name:    "FBM2D",
			fixture: "fixtures/fbm2d.png",
			generate: func() any {
				return generate2DNoiseImage(100, 100, 0.05, func(x, y float32) float32 {
					return f.Eval(2.0, 0.5, 4, x, y)
				})
			},
			compare: func(t *testing.T, expected, actual any, name string) {
				compareImages(t, expected.(image.Image), actual.(image.Image), name)
			},
		},
		{
			name:    "Simplex1D",
			fixture: "fixtures/simplex1d.png",
			generate: func() any {
				return generate1DNoiseImage(400, 100, 0.02, func(x float32) float32 {
					return s.Eval(x)
				})
			},
			compare: func(t *testing.T, expected, actual any, name string) {
				compareImages(t, expected.(image.Image), actual.(image.Image), name)
			},
		},
		{
			name:    "FBM1D",
			fixture: "fixtures/fbm1d.png",
			generate: func() any {
				return generate1DNoiseImage(400, 100, 0.02, func(x float32) float32 {
					return f.Eval(2.0, 0.5, 4, x)
				})
			},
			compare: func(t *testing.T, expected, actual any, name string) {
				compareImages(t, expected.(image.Image), actual.(image.Image), name)
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

			var expected any
			if tt.name == "Simplex3D" || tt.name == "FBM3D" {
				expected, err = gif.DecodeAll(refFile)
			} else {
				expected, err = png.Decode(refFile)
			}
			assert.NoError(t, err)

			// Compare with reference
			tt.compare(t, expected, actual, tt.name)
			t.Logf("%s matches reference: %s", tt.name, tt.fixture)
		})
	}
}

// createGreyscalePalette creates a 256-color greyscale palette
func createGreyscalePalette() color.Palette {
	palette := make(color.Palette, 256)
	for i := 0; i < 256; i++ {
		grey := uint8(i)
		palette[i] = color.RGBA{grey, grey, grey, 255}
	}
	return palette
}

// normalizeNoise converts noise from [-1,1] to [0,1] range
func normalizeNoise(noise float32) float32 {
	normalized := (noise + 1.0) / 2.0
	if normalized < 0 {
		normalized = 0
	}
	if normalized > 1 {
		normalized = 1
	}
	return normalized
}

// compareImages compares two images pixel by pixel
func compareImages(t *testing.T, expected, actual image.Image, testName string) {
	assert.Equal(t, expected.Bounds(), actual.Bounds(), "%s: image bounds should match", testName)

	bounds := actual.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			expectedColor := expected.At(x, y)
			actualColor := actual.At(x, y)
			assert.Equal(t, expectedColor, actualColor, "%s: pixel (%d,%d) should match", testName, x, y)
		}
	}
}

// compareGIFs compares two GIF animations frame by frame
func compareGIFs(t *testing.T, expected, actual *gif.GIF, testName string) {
	assert.Equal(t, len(expected.Image), len(actual.Image), "%s: frame count should match", testName)
	assert.Equal(t, len(expected.Delay), len(actual.Delay), "%s: delay count should match", testName)

	for i := 0; i < len(actual.Image); i++ {
		assert.Equal(t, expected.Image[i].Bounds(), actual.Image[i].Bounds(), "%s: frame %d bounds should match", testName, i)
		assert.Equal(t, expected.Delay[i], actual.Delay[i], "%s: frame %d delay should match", testName, i)

		// Compare pixel data
		bounds := actual.Image[i].Bounds()
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				expectedIdx := expected.Image[i].ColorIndexAt(x, y)
				actualIdx := actual.Image[i].ColorIndexAt(x, y)
				assert.Equal(t, expectedIdx, actualIdx, "%s: frame %d pixel (%d,%d) should match", testName, i, x, y)
			}
		}
	}
}

// generate2DNoiseImage creates a 2D noise image using the provided noise function
func generate2DNoiseImage(width, height int, scale float32, noiseFunc func(x, y float32) float32) *image.Gray {
	img := image.NewGray(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			noise := noiseFunc(float32(x)*scale, float32(y)*scale)
			normalized := normalizeNoise(noise)
			img.Set(x, y, color.Gray{Y: uint8(normalized * 255)})
		}
	}

	return img
}

// generate1DNoiseImage creates a 1D noise visualization as a line graph
func generate1DNoiseImage(width, height int, scale float32, noiseFunc func(x float32) float32) *image.Gray {
	img := image.NewGray(image.Rect(0, 0, width, height))

	// Fill background with white
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.Gray{Y: 255})
		}
	}

	// Generate 1D noise line
	for x := 0; x < width; x++ {
		noise := noiseFunc(float32(x) * scale)
		normalized := normalizeNoise(noise)
		y := int(normalized * float32(height-1))

		// Draw the line (make it thicker for visibility)
		for dy := -1; dy <= 1; dy++ {
			if py := y + dy; py >= 0 && py < height {
				img.Set(x, py, color.Gray{Y: 0})
			}
		}
	}

	return img
}

// generate3DNoiseGIF creates a 3D noise animation as a GIF
func generate3DNoiseGIF(width, height, frames int, scale float32, noiseFunc func(x, y, z float32) float32) *gif.GIF {
	palette := createGreyscalePalette()
	anim := &gif.GIF{}

	for frame := 0; frame < frames; frame++ {
		img := image.NewPaletted(image.Rect(0, 0, width, height), palette)
		z := float32(frame) * 0.1 // Time parameter for animation

		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				noise := noiseFunc(float32(x)*scale, float32(y)*scale, z)
				normalized := normalizeNoise(noise)
				img.SetColorIndex(x, y, uint8(normalized*255))
			}
		}

		anim.Image = append(anim.Image, img)
		anim.Delay = append(anim.Delay, 10) // 100ms per frame
	}

	return anim
}

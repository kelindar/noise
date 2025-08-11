package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"time"

	"github.com/kelindar/bench"
	"github.com/kelindar/noise"
)

var sizes = []int{1e3, 1e6}

func main() {
	bench.Run(func(b *bench.B) {
		runNoise(b)
		runFBM(b)
	}, bench.WithDuration(10*time.Millisecond), bench.WithSamples(100))
}

func runNoise(b *bench.B) {
	shapes := []struct {
		name string
		gen  func(int) [][2]float32
	}{
		{"seq", dataSeq},
		{"rnd", dataRand},
		{"circ", dataCircle},
	}

	const size = 1000
	s := noise.NewSimplex(0)
	for _, shape := range shapes {
		points := shape.gen(size)
		name := fmt.Sprintf("simplex %s (%s)", formatSize(size), shape.name)
		b.Run(name, func(i int) {
			p := points[i%len(points)]
			_ = s.Eval(p[0], p[1])
		})
	}
}

func runFBM(b *bench.B) {
	shapes := []struct {
		name string
		gen  func(int) [][2]float32
	}{
		{"seq", dataSeq},
		{"rnd", dataRand},
		{"circ", dataCircle},
	}

	const size = 1000
	fbm := noise.NewFBM(0)

	// Test different FBM configurations
	configs := []struct {
		name       string
		octaves    int
		lacunarity float32
		gain       float32
	}{
		{"basic", 4, 2.0, 0.5},
		{"detailed", 6, 2.0, 0.5},
		{"rough", 4, 2.0, 0.7},
		{"smooth", 4, 2.0, 0.3},
	}

	for _, config := range configs {
		for _, shape := range shapes {
			points := shape.gen(size)
			name := fmt.Sprintf("fbm-%s %s (%s)", config.name, formatSize(size), shape.name)
			b.Run(name, func(i int) {
				p := points[i%len(points)]
				_ = fbm.Eval(config.lacunarity, config.gain, config.octaves, p[0], p[1])
			})
		}
	}
}

func formatSize(size int) string {
	if size >= 1e6 {
		return fmt.Sprintf("%.0fM", float64(size)/1e6)
	}
	return fmt.Sprintf("%.0fK", float64(size)/1e3)
}

func dataSeq(n int) [][2]float32 {
	pts := make([][2]float32, n)
	for i := 0; i < n; i++ {
		f := float32(i)
		pts[i] = [2]float32{f, f}
	}
	return pts
}

func dataRand(n int) [][2]float32 {
	pts := make([][2]float32, n)
	for i := 0; i < n; i++ {
		pts[i] = [2]float32{rand.Float32(), rand.Float32()}
	}
	return pts
}

func dataCircle(n int) [][2]float32 {
	pts := make([][2]float32, n)
	for i := 0; i < n; i++ {
		angle := 2 * math.Pi * float64(i) / float64(n)
		pts[i] = [2]float32{float32(math.Cos(angle)), float32(math.Sin(angle))}
	}
	return pts
}

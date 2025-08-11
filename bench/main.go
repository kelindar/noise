package main

import (
	"math/rand/v2"
	"time"

	"github.com/kelindar/bench"
	"github.com/kelindar/noise"
)

func main() {
	bench.Run(func(b *bench.B) {
		runBenchmarks(b)
	}, bench.WithDuration(10*time.Millisecond), bench.WithSamples(100))
}

func runBenchmarks(b *bench.B) {
	const size = 1000

	// Initialize noise generators
	s := noise.NewSimplex(0)
	fbm := noise.NewFBM(0)
	const seed = uint32(42)

	// Generate test data
	seq1D := dataSeq1D(size)
	rnd1D := dataRand1D(size)
	seq2D := dataSeq2D(size)
	rnd2D := dataRand2D(size)
	seq3D := dataSeq3D(size)
	rnd3D := dataRand3D(size)

	// Benchmark table
	benchmarks := []struct {
		name string
		fn   func(int)
	}{
		// Simplex benchmarks
		{"simplex 1D (seq)", func(i int) {
			p := seq1D[i%len(seq1D)]
			_ = s.Eval(p)
		}},
		{"simplex 1D (rnd)", func(i int) {
			p := rnd1D[i%len(rnd1D)]
			_ = s.Eval(p)
		}},
		{"simplex 2D (seq)", func(i int) {
			p := seq2D[i%len(seq2D)]
			_ = s.Eval(p[0], p[1])
		}},
		{"simplex 2D (rnd)", func(i int) {
			p := rnd2D[i%len(rnd2D)]
			_ = s.Eval(p[0], p[1])
		}},
		{"simplex 3D (seq)", func(i int) {
			p := seq3D[i%len(seq3D)]
			_ = s.Eval(p[0], p[1], p[2])
		}},
		{"simplex 3D (rnd)", func(i int) {
			p := rnd3D[i%len(rnd3D)]
			_ = s.Eval(p[0], p[1], p[2])
		}},

		// FBM benchmarks
		{"fbm 1D (seq)", func(i int) {
			p := seq1D[i%len(seq1D)]
			_ = fbm.Eval(2.0, 0.5, 4, p)
		}},
		{"fbm 1D (rnd)", func(i int) {
			p := rnd1D[i%len(rnd1D)]
			_ = fbm.Eval(2.0, 0.5, 4, p)
		}},
		{"fbm 2D (seq)", func(i int) {
			p := seq2D[i%len(seq2D)]
			_ = fbm.Eval(2.0, 0.5, 4, p[0], p[1])
		}},
		{"fbm 2D (rnd)", func(i int) {
			p := rnd2D[i%len(rnd2D)]
			_ = fbm.Eval(2.0, 0.5, 4, p[0], p[1])
		}},
		{"fbm 3D (seq)", func(i int) {
			p := seq3D[i%len(seq3D)]
			_ = fbm.Eval(2.0, 0.5, 4, p[0], p[1], p[2])
		}},
		{"fbm 3D (rnd)", func(i int) {
			p := rnd3D[i%len(rnd3D)]
			_ = fbm.Eval(2.0, 0.5, 4, p[0], p[1], p[2])
		}},

		// White noise benchmarks (using White function with coordinates)
		{"white 1D (seq)", func(i int) {
			p := seq1D[i%len(seq1D)]
			_ = noise.White(seed, p)
		}},
		{"white 1D (rnd)", func(i int) {
			p := rnd1D[i%len(rnd1D)]
			_ = noise.White(seed, p)
		}},
		{"white 2D (seq)", func(i int) {
			p := seq2D[i%len(seq2D)]
			_ = noise.White(seed, p[0], p[1])
		}},
		{"white 2D (rnd)", func(i int) {
			p := rnd2D[i%len(rnd2D)]
			_ = noise.White(seed, p[0], p[1])
		}},
		{"white 3D (seq)", func(i int) {
			p := seq3D[i%len(seq3D)]
			_ = noise.White(seed, p[0], p[1], p[2])
		}},
		{"white 3D (rnd)", func(i int) {
			p := rnd3D[i%len(rnd3D)]
			_ = noise.White(seed, p[0], p[1], p[2])
		}},
	}

	// Run all benchmarks
	for _, bm := range benchmarks {
		b.Run(bm.name, bm.fn)
	}
}

// 1D data generators
func dataSeq1D(n int) []float32 {
	pts := make([]float32, n)
	for i := 0; i < n; i++ {
		pts[i] = float32(i)
	}
	return pts
}

func dataRand1D(n int) []float32 {
	pts := make([]float32, n)
	for i := 0; i < n; i++ {
		pts[i] = rand.Float32() * 1000
	}
	return pts
}

// 2D data generators
func dataSeq2D(n int) [][2]float32 {
	pts := make([][2]float32, n)
	for i := 0; i < n; i++ {
		f := float32(i)
		pts[i] = [2]float32{f, f}
	}
	return pts
}

func dataRand2D(n int) [][2]float32 {
	pts := make([][2]float32, n)
	for i := 0; i < n; i++ {
		pts[i] = [2]float32{rand.Float32() * 1000, rand.Float32() * 1000}
	}
	return pts
}

// 3D data generators
func dataSeq3D(n int) [][3]float32 {
	pts := make([][3]float32, n)
	for i := 0; i < n; i++ {
		f := float32(i)
		pts[i] = [3]float32{f, f, f}
	}
	return pts
}

func dataRand3D(n int) [][3]float32 {
	pts := make([][3]float32, n)
	for i := 0; i < n; i++ {
		pts[i] = [3]float32{rand.Float32() * 1000, rand.Float32() * 1000, rand.Float32() * 1000}
	}
	return pts
}

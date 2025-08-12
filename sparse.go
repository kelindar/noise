package noise

import (
	"iter"
	"math"
)

// SSI1 generates a 1D hard-core pattern as a streaming iterator.
// Method: Simple Sequential Inhibition on a unit lattice with one jittered
// candidate per integer cell in [−r1, +r1]. A candidate is accepted only if
// it is at least 1.0 units from all previously accepted samples.
// Traversal order: center-out, visiting 0, +1, −1, +2, −2, ...
// Deterministic for a given seed.
// Complexity: O(n²) with the global scan used for distance checks.
//
// Notes:
//   - Up to 3 jitter attempts per cell, at most one accepted sample per cell.
//   - Produces stratified, well-spaced samples. Not a Poisson-disk generator,
//     and no blue-noise guarantee is implied.
//
// Example:
//
//	for x := range SSI1(12345, 128) {
//	    // use x
//	}
func SSI1(seed uint32, r1 int) iter.Seq[float32] {
	return func(yield func(float32) bool) {
		if r1 <= 0 {
			return
		}
		points := make([]float32, 0, 2*r1+1)
		const minDist = 1.0

		tryCell := func(ix int) bool {
			for t := 0; t < 3; t++ {
				h := xxhash64(uint64(int64(ix)), uint64(seed)^uint64(t))
				x := float32(ix) + (Float32(seed, h) - 0.5)
				ok := true
				for _, e := range points {
					if abs(x-e) < minDist {
						ok = false
						break
					}
				}
				if ok {
					points = append(points, x)
					return !yield(x)
				}
			}
			return false
		}

		if tryCell(0) {
			return
		}
		for r := 1; r <= r1; r++ {
			if tryCell(r) {
				return
			}
			if tryCell(-r) {
				return
			}
		}
	}
}

// SSI2 generates a 2D hard-core pattern as a streaming iterator.
// Method: Simple Sequential Inhibition on a unit lattice with one jittered
// candidate per integer cell in the rectangle [−r1, +r1] × [−r2, +r2].
// A candidate is accepted only if its squared distance to all accepted samples
// is ≥ 1.0. Cells are visited in expanding square rings, center-out.
// Deterministic for a given seed.
// Complexity: O(n²) with the global scan used for distance checks.
//
// Notes:
//   - Up to 2 jitter attempts per cell, at most one accepted sample per cell.
//   - Pattern is stratified and well-spaced, not a Poisson-disk generator and
//     no blue-noise guarantee is implied.
//   - For large radii, replace the global scan with a 3×3 cell-neighborhood
//     check to obtain O(1) acceptance time per sample.
//
// Example:
//
//	for p := range SSI2(12345, 128, 128) {
//	    x, y := p[0], p[1]
//	    // use x,y
//	}
func SSI2(seed uint32, r1, r2 int) iter.Seq[[2]float32] {
	return func(yield func([2]float32) bool) {
		if r1 <= 0 || r2 <= 0 {
			return
		}
		pts := make([][2]float32, 0, (2*r1+1)*(2*r2+1))
		const min2 = 1.0

		tryCell := func(ix, iy int) bool {
			for t := 0; t < 2; t++ {
				h := xxhash64(uint64(int64(ix))*0x9e3779b97f4a7c15^uint64(int64(iy))*0xc2b2ae3d27d4eb4f, uint64(seed)^uint64(t))
				x := float32(ix) + (Float32(seed, h) - 0.5)
				y := float32(iy) + (Float32(seed^1, h) - 0.5)
				ok := true
				for _, q := range pts {
					dx := x - q[0]
					dy := y - q[1]
					if dx*dx+dy*dy < min2 {
						ok = false
						break
					}
				}
				if ok {
					pt := [2]float32{x, y}
					pts = append(pts, pt)
					return !yield(pt)
				}
			}
			return false
		}

		if tryCell(0, 0) {
			return
		}
		maxR := r1
		if r2 > maxR {
			maxR = r2
		}
		for r := 1; r <= maxR; r++ {
			// Only iterate within the bounds defined by r1 and r2
			for ix := -r; ix <= r; ix++ {
				if ix >= -r1 && ix <= r1 {
					if -r >= -r2 && -r <= r2 {
						if tryCell(ix, -r) {
							return
						}
					}
					if r >= -r2 && r <= r2 {
						if tryCell(ix, r) {
							return
						}
					}
				}
			}
			for iy := -r + 1; iy <= r-1; iy++ {
				if iy >= -r2 && iy <= r2 {
					if -r >= -r1 && -r <= r1 {
						if tryCell(-r, iy) {
							return
						}
					}
					if r >= -r1 && r <= r1 {
						if tryCell(r, iy) {
							return
						}
					}
				}
			}
		}
	}
}

// Sparse1 emits integer x positions with at least gap units between any two,
// streamed in a center-out order across [0, w).
// Method: maps the SSI1 jittered lattice samples to pixel space by scaling
// by gap and centering to w, then drops out-of-bounds indices.
// Properties:
//   - Deterministic for a given seed.
//   - Minimum integer spacing equals gap.
//   - Empty sequence if w <= 0 or gap <= 0.
//
// Complexity: inherits SSI1 cost; O(n²) with global checks, O(n) with neighbor map.
//
// Example:
//
//	for ix := range Sparse1(12345, 512, 8) {
//	    // use ix
//	}
func Sparse1(seed uint32, w, gap int) iter.Seq[int] {
	return func(yield func(int) bool) {
		if w <= 0 || gap <= 0 {
			return
		}

		// Radius in cell units so that after scaling and centering we cover [0,w)
		r1 := int(math.Ceil(float64(w) / float64(2*gap)))
		center := float32(w) / 2

		for x := range SSI1(seed, r1) {
			ix := int(x*float32(gap) + center) // scale and center, cast like in tests
			if ix < 0 || ix >= w {
				continue
			}
			if !yield(ix) {
				return
			}
		}
	}
}

// Sparse2 emits integer (x, y) positions with at least gap units between any two,
// streamed in a center-out order across the rectangle [0, w) × [0, h).
// Method: maps the SSI2 jittered lattice samples to pixel space by scaling
// by gap in both axes and centering to w and h, then drops out-of-bounds indices.
// Properties:
//   - Deterministic for a given seed.
//   - Minimum integer spacing equals gap in Euclidean metric.
//   - Empty sequence if w <= 0, h <= 0, or gap <= 0.
//
// Complexity: inherits SSI2 cost; O(n²) with global checks, O(n) with neighbor map.
//
// Example:
//
//	for p := range Sparse2(12345, 512, 256, 8) {
//	    x, y := p[0], p[1]
//	    // use x, y
//	}
func Sparse2(seed uint32, w, h, gap int) iter.Seq[[2]int] {
	return func(yield func([2]int) bool) {
		if w <= 0 || h <= 0 || gap <= 0 {
			return
		}

		// Radii in cell units so that after scaling and centering we cover [0,w) x [0,h)
		r1 := int(math.Ceil(float64(w) / float64(2*gap)))
		r2 := int(math.Ceil(float64(h) / float64(2*gap)))
		centerX := float32(w) / 2
		centerY := float32(h) / 2

		for pt := range SSI2(seed, r1, r2) {
			ix := int(pt[0]*float32(gap) + centerX) // scale and center, cast like in tests
			iy := int(pt[1]*float32(gap) + centerY)
			if ix < 0 || ix >= w || iy < 0 || iy >= h {
				continue
			}
			if !yield([2]int{ix, iy}) {
				return
			}
		}
	}
}

// local helper
func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

package noise

import (
	"iter"
	"math"

	"github.com/kelindar/bitmap"
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

		// Spatial grid using bitmap: each bit tracks if a cell contains a point
		// Grid resolution: 0.5 units per cell (since minDist = 1.0)
		gridSize := r1*4 + 10 // Extra padding for jitter
		var grid bitmap.Bitmap
		grid.Grow(uint32(gridSize - 1)) // Preallocate bitmap to avoid reallocations
		gridOffset := gridSize / 2      // Center offset
		const minDist = 1.0
		const cellSize = 0.5 // Grid resolution

		// Convert world coordinate to grid index
		worldToGrid := func(x float32) uint32 {
			return uint32(int(x/cellSize) + gridOffset)
		}

		// Check if position conflicts with existing points
		isValid := func(x float32) bool {
			gx := worldToGrid(x)
			// Check 5-cell neighborhood (covers minDist = 1.0 with cellSize = 0.5)
			for dx := -2; dx <= 2; dx++ {
				idx := int(gx) + dx
				if idx >= 0 && idx < gridSize && grid.Contains(uint32(idx)) {
					return false // Conflict found
				}
			}
			return true
		}

		// Mark position as occupied
		markOccupied := func(x float32) {
			gx := worldToGrid(x)
			if int(gx) >= 0 && int(gx) < gridSize {
				grid.Set(gx)
			}
		}

		tryCell := func(ix int) bool {
			for t := 0; t < 3; t++ {
				h := xxhash64(uint64(int64(ix)), uint64(seed)^uint64(t))
				x := float32(ix) + (Float32(seed, h) - 0.5)

				if isValid(x) {
					markOccupied(x)
					return !yield(x)
				}
			}
			return false
		}

		// BFS-style expansion from center
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

		// 2D Spatial grid using bitmap: each bit tracks if a cell contains a point
		// Grid resolution: 0.5 units per cell (since minDist = 1.0)
		gridW := r1*4 + 10 // Extra padding for jitter
		gridH := r2*4 + 10
		var grid bitmap.Bitmap
		totalCells := uint32(gridW * gridH)
		grid.Grow(totalCells - 1) // Preallocate bitmap to avoid reallocations
		gridOffsetX := gridW / 2  // Center offset
		gridOffsetY := gridH / 2
		const minDist2 = 1.0 // Squared distance
		const cellSize = 0.5 // Grid resolution

		// Convert grid coordinates to 1D index
		coordToIndex := func(gx, gy int) uint32 {
			return uint32(gy*gridW + gx)
		}

		// Check if position conflicts with existing points
		isValid := func(x, y float32) bool {
			gx := int(x/cellSize) + gridOffsetX
			gy := int(y/cellSize) + gridOffsetY
			// Check 5x5 neighborhood (covers minDist = 1.0 with cellSize = 0.5)
			for dy := -2; dy <= 2; dy++ {
				for dx := -2; dx <= 2; dx++ {
					nx, ny := gx+dx, gy+dy
					if nx >= 0 && nx < gridW && ny >= 0 && ny < gridH {
						idx := coordToIndex(nx, ny)
						if grid.Contains(idx) {
							return false // Conflict found
						}
					}
				}
			}
			return true
		}

		// Mark position as occupied
		markOccupied := func(x, y float32) {
			gx := int(x/cellSize) + gridOffsetX
			gy := int(y/cellSize) + gridOffsetY
			if gx >= 0 && gx < gridW && gy >= 0 && gy < gridH {
				idx := coordToIndex(gx, gy)
				grid.Set(idx)
			}
		}

		tryCell := func(ix, iy int) bool {
			for t := 0; t < 2; t++ {
				h := xxhash64(uint64(int64(ix))*0x9e3779b97f4a7c15^uint64(int64(iy))*0xc2b2ae3d27d4eb4f, uint64(seed)^uint64(t))
				x := float32(ix) + (Float32(seed, h) - 0.5)
				y := float32(iy) + (Float32(seed^1, h) - 0.5)

				if isValid(x, y) {
					markOccupied(x, y)
					pt := [2]float32{x, y}
					return !yield(pt)
				}
			}
			return false
		}

		// BFS-style expansion from center
		if tryCell(0, 0) {
			return
		}
		maxR := r1
		if r2 > maxR {
			maxR = r2
		}
		for r := 1; r <= maxR; r++ {
			// Pre-compute valid ranges for this ring to eliminate redundant boundary checks
			ixMin := -r
			if ixMin < -r1 {
				ixMin = -r1
			}
			ixMax := r
			if ixMax > r1 {
				ixMax = r1
			}

			iyMin := -r + 1
			if iyMin < -r2 {
				iyMin = -r2
			}
			iyMax := r - 1
			if iyMax > r2 {
				iyMax = r2
			}

			// Top edge (y = -r)
			if -r >= -r2 && -r <= r2 {
				for ix := ixMin; ix <= ixMax; ix++ {
					if tryCell(ix, -r) {
						return
					}
				}
			}

			// Bottom edge (y = r)
			if r >= -r2 && r <= r2 {
				for ix := ixMin; ix <= ixMax; ix++ {
					if tryCell(ix, r) {
						return
					}
				}
			}

			// Left edge (x = -r)
			if -r >= -r1 && -r <= r1 {
				for iy := iyMin; iy <= iyMax; iy++ {
					if tryCell(-r, iy) {
						return
					}
				}
			}

			// Right edge (x = r)
			if r >= -r1 && r <= r1 {
				for iy := iyMin; iy <= iyMax; iy++ {
					if tryCell(r, iy) {
						return
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

		// Pre-compute gap as float32 to reduce conversions
		gapF := float32(gap)

		for x := range SSI1(seed, r1) {
			ix := int(x*gapF + center) // scale and center, cast like in tests
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

		// Pre-compute gap as float32 to reduce conversions
		gapF := float32(gap)

		for pt := range SSI2(seed, r1, r2) {
			ix := int(pt[0]*gapF + centerX) // scale and center, cast like in tests
			iy := int(pt[1]*gapF + centerY)
			if ix < 0 || ix >= w || iy < 0 || iy >= h {
				continue
			}
			if !yield([2]int{ix, iy}) {
				return
			}
		}
	}
}

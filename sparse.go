package noise

import (
	"iter"
	"math"

	"github.com/kelindar/bitmap"
)

// Shared SSI constants
const (
	ssiCellSize       = float32(0.5) // half the min distance (1.0)
	ssiNeighborRadius = 2            // covers 1.0 at 0.5 cell size
)

// coordToIndex packs 2D grid coords into a row-major 1D index
func coordToIndex(gx, gy, w int) uint32 { return uint32(gy*w + gx) }

// grid1 encapsulates a 1D bitmap grid used by SSI1
type grid1 struct {
	grid   bitmap.Bitmap
	size   int
	offset int
}

func newGrid1(r1 int) grid1 {
	gridSize := r1*4 + 10
	var b bitmap.Bitmap
	b.Grow(uint32(gridSize - 1))
	return grid1{grid: b, size: gridSize, offset: gridSize / 2}
}

func (g *grid1) IsValid(x float32) bool {
	gx := int(x/ssiCellSize) + g.offset
	if gx < 0 || gx >= g.size {
		return false
	}
	for dx := -ssiNeighborRadius; dx <= ssiNeighborRadius; dx++ {
		idx := gx + dx
		if idx >= 0 && idx < g.size && g.grid.Contains(uint32(idx)) {
			return false
		}
	}
	return true
}

func (g *grid1) Set(x float32) {
	gx := int(x/ssiCellSize) + g.offset
	if gx >= 0 && gx < g.size {
		g.grid.Set(uint32(gx))
	}
}

// grid2 encapsulates a 2D bitmap grid used by SSI2
type grid2 struct {
	grid       bitmap.Bitmap
	w, h       int
	offX, offY int
}

func newGrid2(r1, r2 int) grid2 {
	w := r1*4 + 10
	h := r2*4 + 10
	var b bitmap.Bitmap
	b.Grow(uint32(w*h - 1))
	return grid2{grid: b, w: w, h: h, offX: w / 2, offY: h / 2}
}

func (g *grid2) IsValid(x, y float32) bool {
	gx := int(x/ssiCellSize) + g.offX
	gy := int(y/ssiCellSize) + g.offY
	if gx < 0 || gx >= g.w || gy < 0 || gy >= g.h {
		return false
	}
	for dy := -ssiNeighborRadius; dy <= ssiNeighborRadius; dy++ {
		for dx := -ssiNeighborRadius; dx <= ssiNeighborRadius; dx++ {
			nx, ny := gx+dx, gy+dy
			if nx >= 0 && nx < g.w && ny >= 0 && ny < g.h {
				if g.grid.Contains(coordToIndex(nx, ny, g.w)) {
					return false
				}
			}
		}
	}
	return true
}

func (g *grid2) Set(x, y float32) {
	gx := int(x/ssiCellSize) + g.offX
	gy := int(y/ssiCellSize) + g.offY
	if gx >= 0 && gx < g.w && gy >= 0 && gy < g.h {
		g.grid.Set(coordToIndex(gx, gy, g.w))
	}
}

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

		g := newGrid1(r1)
		tryCell := func(ix int) bool {
			for t := 0; t < 3; t++ {
				h := xxhash64(uint64(int64(ix)), uint64(seed)^uint64(t))
				x := float32(ix) + (Float32(seed, h) - 0.5)

				if g.IsValid(x) {
					g.Set(x)
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

		g := newGrid2(r1, r2)
		tryCell := func(ix, iy int) bool {
			for t := 0; t < 2; t++ {
				h := xxhash64(uint64(int64(ix))*0x9e3779b97f4a7c15^uint64(int64(iy))*0xc2b2ae3d27d4eb4f, uint64(seed)^uint64(t))
				x := float32(ix) + (Float32(seed, h) - 0.5)
				y := float32(iy) + (Float32(seed^1, h) - 0.5)

				if g.IsValid(x, y) {
					g.Set(x, y)
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

		for r := 1; r <= max(r1, r2); r++ {
			ixMin := max(-r, -r1)
			ixMax := min(r, r1)
			iyMin := max(-r+1, -r2)
			iyMax := min(r-1, r2)

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
		c := float32(w) / 2
		gapF := float32(gap)
		for x := range SSI1(seed, r1) {
			ix := int(x*gapF + c) // scale and center, cast like in tests
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
		cx := float32(w) / 2
		cy := float32(h) / 2
		gapF := float32(gap)

		for pt := range SSI2(seed, r1, r2) {
			ix := int(pt[0]*gapF + cx) // scale and center, cast like in tests
			iy := int(pt[1]*gapF + cy)
			if ix < 0 || ix >= w || iy < 0 || iy >= h {
				continue
			}
			if !yield([2]int{ix, iy}) {
				return
			}
		}
	}
}

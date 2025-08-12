package noise

import (
	"iter"
	"math"
)

// Sparse1: float 1D sparse sequence, radiating out to ±r1
func Sparse1(seed uint32, r1 int) iter.Seq[float32] {
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

// Sparse2: float 2D sparse sequence, radiating until x reaches r1 and y reaches r2
func Sparse2(seed uint32, r1, r2 int) iter.Seq[[2]float32] {
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

// Sparse3: float 3D sparse sequence, radiating until x reaches r1, y reaches r2, z reaches r3
func Sparse3(seed uint32, r1, r2, r3 int) iter.Seq[[3]float32] {
	return func(yield func([3]float32) bool) {
		if r1 <= 0 || r2 <= 0 || r3 <= 0 {
			return
		}
		pts := make([][3]float32, 0, (2*r1+1)*(2*r2+1)*(2*r3+1))
		const min2 = 1.0

		tryCell := func(ix, iy, iz int) bool {
			for t := 0; t < 2; t++ {
				h := uint64(int64(ix))*0x9e3779b97f4a7c15 ^ uint64(int64(iy))*0xc2b2ae3d27d4eb4f ^ uint64(int64(iz))*0x165667b19e3779f9
				h = xxhash64(h, uint64(seed)^uint64(t))
				x := float32(ix) + (Float32(seed, h) - 0.5)
				y := float32(iy) + (Float32(seed^1, h) - 0.5)
				z := float32(iz) + (Float32(seed^2, h) - 0.5)
				ok := true
				for _, q := range pts {
					dx := x - q[0]
					dy := y - q[1]
					dz := z - q[2]
					if dx*dx+dy*dy+dz*dz < min2 {
						ok = false
						break
					}
				}
				if ok {
					pt := [3]float32{x, y, z}
					pts = append(pts, pt)
					return !yield(pt)
				}
			}
			return false
		}

		if tryCell(0, 0, 0) {
			return
		}
		maxR := r1
		if r2 > maxR {
			maxR = r2
		}
		if r3 > maxR {
			maxR = r3
		}
		for r := 1; r <= maxR; r++ {
			// Only iterate within the bounds defined by r1, r2, and r3
			for ix := -r; ix <= r; ix++ {
				if ix >= -r1 && ix <= r1 {
					for iy := -r; iy <= r; iy++ {
						if iy >= -r2 && iy <= r2 {
							if -r >= -r3 && -r <= r3 {
								if tryCell(ix, iy, -r) {
									return
								}
							}
							if r >= -r3 && r <= r3 {
								if tryCell(ix, iy, r) {
									return
								}
							}
						}
					}
				}
			}
			for iy := -r + 1; iy <= r-1; iy++ {
				if iy >= -r2 && iy <= r2 {
					for iz := -r + 1; iz <= r-1; iz++ {
						if iz >= -r3 && iz <= r3 {
							if -r >= -r1 && -r <= r1 {
								if tryCell(-r, iy, iz) {
									return
								}
							}
							if r >= -r1 && r <= r1 {
								if tryCell(r, iy, iz) {
									return
								}
							}
						}
					}
				}
			}
			for ix := -r + 1; ix <= r-1; ix++ {
				if ix >= -r1 && ix <= r1 {
					for iz := -r + 1; iz <= r-1; iz++ {
						if iz >= -r3 && iz <= r3 {
							if -r >= -r2 && -r <= r2 {
								if tryCell(ix, -r, iz) {
									return
								}
							}
							if r >= -r2 && r <= r2 {
								if tryCell(ix, r, iz) {
									return
								}
							}
						}
					}
				}
			}
		}
	}
}

// SparseFill1: integer 1D sparse sequence, scaled to fit within width w with minimum gap
func SparseFill1(seed uint32, w, gap int) iter.Seq[int] {
	return func(yield func(int) bool) {
		if w <= 0 || gap <= 0 {
			return
		}

		// Radius in cell units so that after scaling and centering we cover [0,w)
		r1 := int(math.Ceil(float64(w) / float64(2*gap)))
		center := float32(w) / 2
		seen := make(map[int]struct{})

		for x := range Sparse1(seed, r1) {
			ix := int(x*float32(gap) + center) // scale and center, cast like in tests
			if ix < 0 || ix >= w {
				continue
			}
			if _, ok := seen[ix]; ok {
				continue
			}
			seen[ix] = struct{}{}
			if !yield(ix) {
				return
			}
		}
	}
}

// SparseFill2: integer 2D sparse sequence, scaled to fit within width w and height h with minimum gap
func SparseFill2(seed uint32, w, h, gap int) iter.Seq[[2]int] {
	return func(yield func([2]int) bool) {
		if w <= 0 || h <= 0 || gap <= 0 {
			return
		}

		// Radii in cell units so that after scaling and centering we cover [0,w) x [0,h)
		r1 := int(math.Ceil(float64(w) / float64(2*gap)))
		r2 := int(math.Ceil(float64(h) / float64(2*gap)))
		centerX := float32(w) / 2
		centerY := float32(h) / 2
		seen := make(map[int]struct{}) // key := iy*w + ix

		for pt := range Sparse2(seed, r1, r2) {
			ix := int(pt[0]*float32(gap) + centerX) // scale and center, cast like in tests
			iy := int(pt[1]*float32(gap) + centerY)
			if ix < 0 || ix >= w || iy < 0 || iy >= h {
				continue
			}
			key := iy*w + ix
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			if !yield([2]int{ix, iy}) {
				return
			}
		}
	}
}

// SparseFill3: integer 3D sparse sequence, scaled to fit within width w, height h, and depth d with minimum gap
func SparseFill3(seed uint32, w, h, d, gap int) iter.Seq[[3]int] {
	return func(yield func([3]int) bool) {
		if w <= 0 || h <= 0 || d <= 0 || gap <= 0 {
			return
		}

		// Calculate the radii needed to cover the dimensions
		r1 := (w + gap - 1) / gap
		r2 := (h + gap - 1) / gap
		r3 := (d + gap - 1) / gap

		for pt := range Sparse3(seed, r1, r2, r3) {
			// Scale and cast to integer coordinates
			ix := int(pt[0] * float32(gap))
			iy := int(pt[1] * float32(gap))
			iz := int(pt[2] * float32(gap))

			// Check bounds and handle negative coordinates
			validX := (ix >= 0 && ix < w) || (ix < 0 && -ix < w)
			validY := (iy >= 0 && iy < h) || (iy < 0 && -iy < h)
			validZ := (iz >= 0 && iz < d) || (iz < 0 && -iz < d)

			if validX && validY && validZ {
				// Convert negative coordinates to positive
				if ix < 0 {
					ix = -ix
				}
				if iy < 0 {
					iy = -iy
				}
				if iz < 0 {
					iz = -iz
				}

				if !yield([3]int{ix, iy, iz}) {
					return
				}
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

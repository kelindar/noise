package noise

import "math/rand"

const (
	f2 = 0.36602542 // float32(0.5 * (math.Sqrt(3) - 1))
	g2 = 0.21132487 // float32((3 - math.Sqrt(3)) / 6)
	f3 = 1.0 / 3.0  // for 3D skewing
	g3 = 1.0 / 6.0  // for 3D unskewing
)

var (
	perm  [512]uint8
	grad2 [512][2]float32
	grad3 [512][3]float32
)

// Simplex represents a simplex noise generator with its own permutation table
type Simplex struct {
	perm  [512]uint8
	grad2 [512][2]float32
	grad3 [512][3]float32
}

// FBM represents a fractal Brownian motion generator
type FBM struct {
	simplex *Simplex
}

var table = [...]uint8{151, 160, 137, 91, 90, 15,
	131, 13, 201, 95, 96, 53, 194, 233, 7, 225, 140, 36, 103, 30, 69, 142, 8, 99, 37, 240, 21, 10, 23,
	190, 6, 148, 247, 120, 234, 75, 0, 26, 197, 62, 94, 252, 219, 203, 117, 35, 11, 32, 57, 177, 33,
	88, 237, 149, 56, 87, 174, 20, 125, 136, 171, 168, 68, 175, 74, 165, 71, 134, 139, 48, 27, 166,
	77, 146, 158, 231, 83, 111, 229, 122, 60, 211, 133, 230, 220, 105, 92, 41, 55, 46, 245, 40, 244,
	102, 143, 54, 65, 25, 63, 161, 1, 216, 80, 73, 209, 76, 132, 187, 208, 89, 18, 169, 200, 196,
	135, 130, 116, 188, 159, 86, 164, 100, 109, 198, 173, 186, 3, 64, 52, 217, 226, 250, 124, 123,
	5, 202, 38, 147, 118, 126, 255, 82, 85, 212, 207, 206, 59, 227, 47, 16, 58, 17, 182, 189, 28, 42,
	223, 183, 170, 213, 119, 248, 152, 2, 44, 154, 163, 70, 221, 153, 101, 155, 167, 43, 172, 9,
	129, 22, 39, 253, 19, 98, 108, 110, 79, 113, 224, 232, 178, 185, 112, 104, 218, 246, 97, 228,
	251, 34, 242, 193, 238, 210, 144, 12, 191, 179, 162, 241, 81, 51, 145, 235, 249, 14, 239, 107,
	49, 192, 214, 31, 181, 199, 106, 157, 184, 84, 204, 176, 115, 121, 50, 45, 127, 4, 150, 254,
	138, 236, 205, 93, 222, 114, 67, 29, 24, 72, 243, 141, 128, 195, 78, 66, 215, 61, 156, 180}

func init() {
	var g2d = [12]uint16{
		0x0101, // [+1, +1]
		0xff01, // [-1, +1]
		0x01ff, // [+1, -1]
		0xffff, // [-1, -1]
		0x0100, // [+1, +0]
		0xff00, // [-1, +0]
		0x0100, // [+1, +0]
		0xff00, // [-1, +0]
		0x0001, // [+0, +1]
		0x00ff, // [+0, -1]
		0x0001, // [+0, +1]
		0x00ff, // [+0, -1]
	}

	for i := 0; i < 512; i++ {
		perm[i] = table[i&255]
		idx := g2d[perm[i]%12]
		gx := int8(idx >> 8)
		gy := int8(idx)
		grad2[i] = [2]float32{float32(gx), float32(gy)}
		// Initialize 3D gradients (12 unit vectors)
		grad3[i] = [3]float32{float32(gx), float32(gy), 0}
	}
}

// Noise2 computes a two dimensional simplex noise
// Public Domain: https://weber.itn.liu.se/~stegu/simplexnoise/simplexnoise.pdf
// Reference: https://mrl.cs.nyu.edu/~perlin/noise/
func Noise2(x, y float32, seed uint32) float32 {

	// Skew the input space to determine which simplex cell we're in
	s := (x + y) * f2
	i := floor(x + s)
	j := floor(y + s)

	// Unskew the cell origin back to (x,y) space
	t := float32(i+j) * g2
	x0 := x - (float32(i) - t)
	y0 := y - (float32(j) - t)

	// For the 2D case, the simplex shape is an equilateral triangle.
	// Determine which simplex we are in
	i1, j1 := float32(0), float32(1) // upper triangle
	if x0 > y0 {                     // lower triangle
		i1 = 1
		j1 = 0
	}

	// Offsets for middle corner in (x,y) unskewed coords
	x1 := x0 - i1 + g2
	y1 := y0 - j1 + g2

	// Offsets for middle corner in (x,y) unskewed coords
	const g = 2*g2 - 1
	x2 := x0 + g
	y2 := y0 + g

	// Work out the hashed gradient indices of the three simplex corners
	si := int(seed) & 255
	pp := perm[(j+si)&255:]
	gg := grad2[(i+si)&255:]
	p0 := int(pp[0])
	p1 := int(pp[int(j1)])
	p2 := int(pp[1])
	g0 := gg[p0]
	g1 := gg[int(i1)+p1]
	g2 := gg[1+p2]

	// Calculate the contribution from the three corners
	n := float32(0.0)
	if t := 0.5 - x0*x0 - y0*y0; t > 0 {
		n += pow4(t) * (g0[0]*x0 + g0[1]*y0)
	}
	if t := 0.5 - x1*x1 - y1*y1; t > 0 {
		n += pow4(t) * (g1[0]*x1 + g1[1]*y1)
	}
	if t := 0.5 - x2*x2 - y2*y2; t > 0 {
		n += pow4(t) * (g2[0]*x2 + g2[1]*y2)
	}

	// Add contributions from each corner to get the final noise value.
	// The result is scaled to return values in the interval [-1,1].
	return 70.0 * n

}

// FBM2 computes fractal Brownian motion using multiple octaves of Noise2
// The output is approximately in [-1,1].
func FBM2(x, y float32, octaves int, lacunarity, gain float32, seed uint32) float32 {
	if octaves <= 0 {
		return 0
	}
	var sum float32
	var amp float32 = 1
	var freq float32 = 1
	var totalAmp float32
	for o := 0; o < octaves; o++ {
		// decorrelate octaves by offsetting the seed
		so := seed + uint32(o)*0x9E3779B1
		sum += amp * Noise2(x*freq, y*freq, so)
		totalAmp += amp
		freq *= lacunarity
		amp *= gain
	}
	if totalAmp > 0 {
		return sum / totalAmp
	}
	return 0
}

// pow4 lifts the value to the power of 4
func pow4(v float32) float32 {
	v *= v
	return v * v
}

// floor floors the floating-point value to an integer
func floor(x float32) int {
	v := int(x)
	if x < float32(v) {
		return v - 1
	}
	return v
}

// NewSimplex creates a new Simplex noise generator with the given seed
func NewSimplex(seed uint32) *Simplex {
	s := &Simplex{}
	s.initWithSeed(seed)
	return s
}

// NewFBM creates a new FBM generator with the given seed
func NewFBM(seed uint32) *FBM {
	return &FBM{
		simplex: NewSimplex(seed),
	}
}

// initWithSeed initializes the Simplex generator with a seeded permutation table
func (s *Simplex) initWithSeed(seed uint32) {
	// Create a seeded random number generator
	rng := rand.New(rand.NewSource(int64(seed)))

	// Initialize permutation table with Fisher-Yates shuffle
	for i := 0; i < 256; i++ {
		s.perm[i] = uint8(i)
	}
	for i := 255; i > 0; i-- {
		j := rng.Intn(i + 1)
		s.perm[i], s.perm[j] = s.perm[j], s.perm[i]
	}
	// Duplicate for wrapping
	for i := 0; i < 256; i++ {
		s.perm[i+256] = s.perm[i]
	}

	// Initialize gradient tables
	var g2d = [12]uint16{
		0x0101, 0xff01, 0x01ff, 0xffff, // diagonal gradients
		0x0100, 0xff00, 0x0100, 0xff00, // horizontal gradients
		0x0001, 0x00ff, 0x0001, 0x00ff, // vertical gradients
	}

	var g3d = [12][3]float32{
		{1, 1, 0}, {-1, 1, 0}, {1, -1, 0}, {-1, -1, 0},
		{1, 0, 1}, {-1, 0, 1}, {1, 0, -1}, {-1, 0, -1},
		{0, 1, 1}, {0, -1, 1}, {0, 1, -1}, {0, -1, -1},
	}

	for i := 0; i < 512; i++ {
		idx2 := g2d[s.perm[i&255]%12]
		gx := int8(idx2 >> 8)
		gy := int8(idx2)
		s.grad2[i] = [2]float32{float32(gx), float32(gy)}

		idx3 := s.perm[i&255] % 12
		s.grad3[i] = g3d[idx3]
	}
}

// Eval evaluates simplex noise at the given coordinates
// Supports 1D, 2D, and 3D noise based on number of arguments
func (s *Simplex) Eval(coords ...float32) float32 {
	switch len(coords) {
	case 1:
		return s.noise1D(coords[0])
	case 2:
		return s.noise2D(coords[0], coords[1])
	case 3:
		return s.noise3D(coords[0], coords[1], coords[2])
	default:
		panic("Eval requires 1, 2, or 3 coordinates")
	}
}

// Eval evaluates fractal Brownian motion at the given coordinates
// First 3 parameters are octaves, lacunarity, gain, followed by 1-3 coordinates
func (f *FBM) Eval(params ...float32) float32 {
	if len(params) < 4 {
		panic("FBM.Eval requires at least 4 parameters: octaves, lacunarity, gain, and 1-3 coordinates")
	}

	octaves := int(params[0])
	lacunarity := params[1]
	gain := params[2]
	coords := params[3:]

	if octaves <= 0 {
		return 0
	}

	var sum float32
	var amp float32 = 1
	var freq float32 = 1
	var totalAmp float32

	for o := 0; o < octaves; o++ {
		// Scale coordinates by frequency
		scaledCoords := make([]float32, len(coords))
		for i, coord := range coords {
			scaledCoords[i] = coord * freq
		}

		sum += amp * f.simplex.Eval(scaledCoords...)
		totalAmp += amp
		freq *= lacunarity
		amp *= gain
	}

	if totalAmp > 0 {
		return sum / totalAmp
	}
	return 0
}

// noise1D computes 1D simplex noise (using 2D with y=0)
func (s *Simplex) noise1D(x float32) float32 {
	return s.noise2D(x, 0)
}

// noise2D computes 2D simplex noise using the generator's permutation table
func (s *Simplex) noise2D(x, y float32) float32 {
	// Skew the input space to determine which simplex cell we're in
	sk := (x + y) * f2
	i := floor(x + sk)
	j := floor(y + sk)

	// Unskew the cell origin back to (x,y) space
	t := float32(i+j) * g2
	x0 := x - (float32(i) - t)
	y0 := y - (float32(j) - t)

	// For the 2D case, the simplex shape is an equilateral triangle.
	// Determine which simplex we are in
	i1, j1 := float32(0), float32(1) // upper triangle
	if x0 > y0 {                     // lower triangle
		i1 = 1
		j1 = 0
	}

	// Offsets for middle corner in (x,y) unskewed coords
	x1 := x0 - i1 + g2
	y1 := y0 - j1 + g2

	// Offsets for last corner in (x,y) unskewed coords
	const g = 2*g2 - 1
	x2 := x0 + g
	y2 := y0 + g

	// Work out the hashed gradient indices of the three simplex corners
	pp := s.perm[j&255:]
	gg := s.grad2[i&255:]
	p0 := int(pp[0])
	p1 := int(pp[int(j1)])
	p2 := int(pp[1])
	g0 := gg[p0]
	g1 := gg[int(i1)+p1]
	g2 := gg[1+p2]

	// Calculate the contribution from the three corners
	n := float32(0.0)
	if t := 0.5 - x0*x0 - y0*y0; t > 0 {
		n += pow4(t) * (g0[0]*x0 + g0[1]*y0)
	}
	if t := 0.5 - x1*x1 - y1*y1; t > 0 {
		n += pow4(t) * (g1[0]*x1 + g1[1]*y1)
	}
	if t := 0.5 - x2*x2 - y2*y2; t > 0 {
		n += pow4(t) * (g2[0]*x2 + g2[1]*y2)
	}

	// Add contributions from each corner to get the final noise value.
	// The result is scaled to return values in the interval [-1,1].
	return 70.0 * n
}

// noise3D computes 3D simplex noise using the generator's permutation table
func (s *Simplex) noise3D(x, y, z float32) float32 {
	// Skew the input space to determine which simplex cell we're in
	sk := (x + y + z) * f3
	i := floor(x + sk)
	j := floor(y + sk)
	k := floor(z + sk)

	// Unskew the cell origin back to (x,y,z) space
	t := float32(i+j+k) * g3
	x0 := x - (float32(i) - t)
	y0 := y - (float32(j) - t)
	z0 := z - (float32(k) - t)

	// For the 3D case, the simplex shape is a slightly irregular tetrahedron.
	// Determine which simplex we are in.
	var i1, j1, k1 float32 // Offsets for second corner of simplex in (i,j,k) coords
	var i2, j2, k2 float32 // Offsets for third corner of simplex in (i,j,k) coords

	if x0 >= y0 {
		if y0 >= z0 {
			i1, j1, k1 = 1, 0, 0
			i2, j2, k2 = 1, 1, 0
		} else if x0 >= z0 {
			i1, j1, k1 = 1, 0, 0
			i2, j2, k2 = 1, 0, 1
		} else {
			i1, j1, k1 = 0, 0, 1
			i2, j2, k2 = 1, 0, 1
		}
	} else {
		if y0 < z0 {
			i1, j1, k1 = 0, 0, 1
			i2, j2, k2 = 0, 1, 1
		} else if x0 < z0 {
			i1, j1, k1 = 0, 1, 0
			i2, j2, k2 = 0, 1, 1
		} else {
			i1, j1, k1 = 0, 1, 0
			i2, j2, k2 = 1, 1, 0
		}
	}

	// A step of (1,0,0) in (i,j,k) means a step of (1-c,-c,-c) in (x,y,z),
	// a step of (0,1,0) in (i,j,k) means a step of (-c,1-c,-c) in (x,y,z), and
	// a step of (0,0,1) in (i,j,k) means a step of (-c,-c,1-c) in (x,y,z), where c = 1/6.
	x1 := x0 - i1 + g3
	y1 := y0 - j1 + g3
	z1 := z0 - k1 + g3
	x2 := x0 - i2 + 2.0*g3
	y2 := y0 - j2 + 2.0*g3
	z2 := z0 - k2 + 2.0*g3
	x3 := x0 - 1.0 + 3.0*g3
	y3 := y0 - 1.0 + 3.0*g3
	z3 := z0 - 1.0 + 3.0*g3

	// Work out the hashed gradient indices of the four simplex corners
	ii := i & 255
	jj := j & 255
	kk := k & 255
	gi0 := s.perm[ii+int(s.perm[jj+int(s.perm[kk])])] % 12
	gi1 := s.perm[ii+int(i1)+int(s.perm[jj+int(j1)+int(s.perm[kk+int(k1)])])] % 12
	gi2 := s.perm[ii+int(i2)+int(s.perm[jj+int(j2)+int(s.perm[kk+int(k2)])])] % 12
	gi3 := s.perm[ii+1+int(s.perm[jj+1+int(s.perm[kk+1])])] % 12

	// Calculate the contribution from the four corners
	var n0, n1, n2, n3 float32

	t0 := 0.6 - x0*x0 - y0*y0 - z0*z0
	if t0 >= 0 {
		g := s.grad3[gi0]
		n0 = t0 * t0 * t0 * t0 * (g[0]*x0 + g[1]*y0 + g[2]*z0)
	}

	t1 := 0.6 - x1*x1 - y1*y1 - z1*z1
	if t1 >= 0 {
		g := s.grad3[gi1]
		n1 = t1 * t1 * t1 * t1 * (g[0]*x1 + g[1]*y1 + g[2]*z1)
	}

	t2 := 0.6 - x2*x2 - y2*y2 - z2*z2
	if t2 >= 0 {
		g := s.grad3[gi2]
		n2 = t2 * t2 * t2 * t2 * (g[0]*x2 + g[1]*y2 + g[2]*z2)
	}

	t3 := 0.6 - x3*x3 - y3*y3 - z3*z3
	if t3 >= 0 {
		g := s.grad3[gi3]
		n3 = t3 * t3 * t3 * t3 * (g[0]*x3 + g[1]*y3 + g[2]*z3)
	}

	// Add contributions from each corner to get the final noise value.
	// The result is scaled to stay just inside [-1,1]
	return 32.0 * (n0 + n1 + n2 + n3)
}

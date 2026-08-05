package geom

import (
	"fmt"
	"math"
)

// Define a 2x3 affine transform matrix
type AffineMatrix [2][3]float64

// Transform returns the result.
func (m AffineMatrix) Transform(p Point) Point {
	return Point{
		X: m[0][0]*p.X + m[0][1]*p.Y + m[0][2],
		Y: m[1][0]*p.X + m[1][1]*p.Y + m[1][2],
	}
}

// RotationAngle returns the result.
func (m AffineMatrix) RotationAngle() float64 {
	return math.Atan2(m[1][0], m[0][0])
}

// NewRotationMat creates and returns a new instance.
func NewRotationMat(center Point, angleDeg float64) AffineMatrix {
	angleRad := angleDeg * math.Pi / 180.0
	cosA := math.Cos(angleRad)
	sinA := math.Sin(angleRad)
	return AffineMatrix{
		{cosA, -sinA, center.X - cosA*center.X + sinA*center.Y},
		{sinA, cosA, center.Y - sinA*center.X - cosA*center.Y},
	}
}

// NewTranslateMat creates and returns a new instance.
func NewTranslateMat(src, dst Point) AffineMatrix {
	return AffineMatrix{
		{1, 0, dst.X - src.X},
		{0, 1, dst.Y - src.Y},
	}
}

// Two coordinate frames O1,O2 with different origins; O2 is rotated by c degrees relative to O1. Points are (x1,y1)/(x2,y2). Given one point in both frames and another point in O2,
//y2), find that point's coordinates in O1 (x1,y1).
// Image transform: treat the image as quadrant IV; input -y, return -y

// NewTranslateRotationMat creates and returns a new instance.
func NewTranslateRotationMat(src, dst Point, angleDeg float64) AffineMatrix {
	// Convert angle from degrees to radians
	angleRad := angleDeg * math.Pi / 180.0
	// Calculate cosine and sine of the angle
	cosA := math.Cos(angleRad)
	sinA := math.Sin(angleRad)
	return AffineMatrix{
		{cosA, -sinA, dst.X - cosA*src.X + sinA*src.Y},
		{sinA, cosA, dst.Y - sinA*src.X - cosA*src.Y},
	}
}

// newAffineMatrix creates and returns a new instance.
func newAffineMatrix(src, dst [3]Point) (AffineMatrix, error) {
	// Build coefficient matrix A and constant vector b
	A := [][]float64{
		{src[0].X, src[0].Y, 1, 0, 0, 0},
		{0, 0, 0, src[0].X, src[0].Y, 1},
		{src[1].X, src[1].Y, 1, 0, 0, 0},
		{0, 0, 0, src[1].X, src[1].Y, 1},
		{src[2].X, src[2].Y, 1, 0, 0, 0},
		{0, 0, 0, src[2].X, src[2].Y, 1},
	}
	b := []float64{dst[0].X, dst[0].Y, dst[1].X, dst[1].Y, dst[2].X, dst[2].Y}

	// Solve Ax = b with Gauss-Jordan elimination
	solution, err := GaussJordanElimination(A, b)
	if err != nil {
		return AffineMatrix{}, err
	}

	// Build the affine transform matrix
	transformMatrix := AffineMatrix{
		{solution[0], solution[1], solution[2]},
		{solution[3], solution[4], solution[5]},
	}

	return transformMatrix, nil
}

// NewAffineMatrix creates and returns a new instance.
func NewAffineMatrix(src, dst [3]Point) (AffineMatrix, error) {
	// Build source and destination point matrices
	srcMatrix := [3][3]float64{
		{src[0].X, src[0].Y, 1},
		{src[1].X, src[1].Y, 1},
		{src[2].X, src[2].Y, 1},
	}
	dstMatrix := [3][2]float64{
		{dst[0].X, dst[0].Y},
		{dst[1].X, dst[1].Y},
		{dst[2].X, dst[2].Y},
	}

	// Invert the source point matrix
	invSrcMatrix, err := InverseMatrix(srcMatrix)
	if err != nil {
		return AffineMatrix{}, err
	}

	// Affine matrix: inv(srcMatrix) * dstMatrix
	affineMatrix := AffineMatrix{}
	for i := 0; i < 3; i++ {
		for j := 0; j < 2; j++ {
			affineMatrix[j][i] = invSrcMatrix[i][0]*dstMatrix[0][j] + invSrcMatrix[i][1]*dstMatrix[1][j] + invSrcMatrix[i][2]*dstMatrix[2][j]
		}
	}
	// Convert to 2x3 form
	return affineMatrix, nil
}

// InverseMatrix performs the operation.
func InverseMatrix(m [3][3]float64) ([3][3]float64, error) {
	det := m[0][0]*(m[1][1]*m[2][2]-m[1][2]*m[2][1]) -
		m[0][1]*(m[1][0]*m[2][2]-m[1][2]*m[2][0]) +
		m[0][2]*(m[1][0]*m[2][1]-m[1][1]*m[2][0])

	if det == 0 {
		return [3][3]float64{}, fmt.Errorf("矩阵不可逆")
	}

	inv := [3][3]float64{}
	inv[0][0] = (m[1][1]*m[2][2] - m[1][2]*m[2][1]) / det
	inv[0][1] = (m[0][2]*m[2][1] - m[0][1]*m[2][2]) / det
	inv[0][2] = (m[0][1]*m[1][2] - m[0][2]*m[1][1]) / det
	inv[1][0] = (m[1][2]*m[2][0] - m[1][0]*m[2][2]) / det
	inv[1][1] = (m[0][0]*m[2][2] - m[0][2]*m[2][0]) / det
	inv[1][2] = (m[0][2]*m[1][0] - m[0][0]*m[1][2]) / det
	inv[2][0] = (m[1][0]*m[2][1] - m[1][1]*m[2][0]) / det
	inv[2][1] = (m[0][1]*m[2][0] - m[0][0]*m[2][1]) / det
	inv[2][2] = (m[0][0]*m[1][1] - m[0][1]*m[1][0]) / det

	return inv, nil
}

// GaussJordanElimination performs the operation.
func GaussJordanElimination(A [][]float64, b []float64) ([]float64, error) {
	n := len(b)
	m := len(A)
	if m != n || len(A[0]) != n {
		return nil, fmt.Errorf("invalid matrix dimensions")
	}

	// Augmented matrix [A | b]
	extendedMatrix := make([][]float64, n)
	for i := range extendedMatrix {
		extendedMatrix[i] = make([]float64, n+1)
		copy(extendedMatrix[i][:n], A[i])
		extendedMatrix[i][n] = b[i]
	}

	// Gauss-Jordan elimination
	for i := 0; i < n; i++ {
		// Find the pivot
		maxRow := i
		for k := i + 1; k < n; k++ {
			if math.Abs(extendedMatrix[k][i]) > math.Abs(extendedMatrix[maxRow][i]) {
				maxRow = k
			}
		}

		// Swap rows
		extendedMatrix[i], extendedMatrix[maxRow] = extendedMatrix[maxRow], extendedMatrix[i]

		// Cannot continue if the pivot is 0
		if extendedMatrix[i][i] == 0 {
			return nil, fmt.Errorf("matrix is singular")
		}

		// Eliminate
		pivot := extendedMatrix[i][i]
		for j := 0; j < n+1; j++ {
			extendedMatrix[i][j] /= pivot
		}
		for k := 0; k < n; k++ {
			if k != i {
				factor := extendedMatrix[k][i]
				for j := 0; j < n+1; j++ {
					extendedMatrix[k][j] -= factor * extendedMatrix[i][j]
				}
			}
		}
	}

	// Extract the solution
	solution := make([]float64, n)
	for i := 0; i < n; i++ {
		solution[i] = extendedMatrix[i][n]
	}

	return solution, nil
}

package geom

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTransformPoint(t *testing.T) {
	angle := -30.0 // degrees
	point1A := Point{4.0, -2.0}
	point2A := Point{5.0, -3.0}
	point1B := Point{7.46, -6.73}
	point2B := Point{X: 7.826025403784438, Y: -8.096025403784438}
	// {7.826025403784439 8.09602540378444}
	mat := NewTranslateRotationMat(point1A, point1B, angle)
	assert.Equal(t, point2B, mat.Transform(point2A))
	t.Log(NewVector(point1A, point2A).AngleWith(NewVector(point1B, point2B)))

	mat = NewTranslateRotationMat(point1A, point1B, angle)
	t.Log(mat.Transform(point2A))
	t.Log(mat)
	mat1 := NewRotationMat(point1A, angle)
	tmpA1 := mat1.Transform(point1A)
	mat2 := NewTranslateMat(tmpA1, point1B)
	tmpA2 := mat1.Transform(point2A)
	t.Log(mat2.Transform(tmpA2))

	mat3 := NewTranslateMat(point1A, point1B)
	tmpB1 := mat3.Transform(point1A)
	mat4 := NewRotationMat(tmpB1, angle)
	tmpB2 := mat3.Transform(point2A)
	t.Log(mat4.Transform(tmpB2))
}

func TestAffineMatrix(t *testing.T) {
	p1, p2, p3, q1, q2, q3 := Point{2000, 7000}, Point{48000, 80000}, Point{2000, 85000}, Point{3558, 17895}, Point{11016, 5997}, Point{3538, 5182}
	transformMatrix, err := NewAffineMatrix([3]Point{p1, p2, p3}, [3]Point{q1, q2, q3})
	if err != nil {
		t.Error(err)
	}

	// apply transform to a point
	p := Point{X: 48000, Y: 13000}
	q := transformMatrix.Transform(p)
	t.Log(q)

	transformMatrix = NewRotationMat(Pt(1, 1), 90)
	t.Log(transformMatrix.Transform(Pt(0, 0)))
	t.Log(transformMatrix.Transform(Pt(2, 2)))
	t.Log(transformMatrix.Transform(Pt(0, 2)))
	t.Log(transformMatrix.Transform(Pt(2, 0)))
}

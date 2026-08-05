package geom

import "math"

type Coord struct{}

// GetOuterPoints ...
func (c Coord) GetOuterPoints(points []Point) [4]Point {
	if len(points) < 4 {
		panic("points at least 4 points")
	}
	var outerPoints [4]Point
	// 初始化四个方向的点
	minX, maxX, minY, maxY := math.MaxFloat64, -math.MaxFloat64, math.MaxFloat64, -math.MaxFloat64
	distance1, distance2, distance3, distance4 := math.MaxFloat64, math.MaxFloat64, math.MaxFloat64, math.MaxFloat64
	for _, point := range points {
		// 计算点的四个方向的外扩位置
		if distance := point.Length(Point{X: minX, Y: minY}); distance < distance1 {
			distance1 = distance
			outerPoints[0] = point
		}
		if distance := point.Length(Point{X: maxX, Y: minY}); distance < distance2 {
			distance2 = distance
			outerPoints[1] = point
		}
		if distance := point.Length(Point{X: minX, Y: maxY}); distance < distance3 {
			distance3 = distance
			outerPoints[2] = point
		}
		if distance := point.Length(Point{X: maxX, Y: maxY}); distance < distance4 {
			distance4 = distance
			outerPoints[3] = point
		}
	}
	return outerPoints
}

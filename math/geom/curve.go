package geom

import "math"

type PolynomialCurve struct {
	Coefficients []float64
	Angle        float64
}

// ZeroAngleY returns the result.
func (p *PolynomialCurve) ZeroAngleY(x float64) float64 {
	var y float64
	n := len(p.Coefficients)
	for i, coff := range p.Coefficients {
		y += coff * math.Pow(x, float64(n-1-i))
	}
	return y
}

type SineWave struct {
	Amplitude     float64
	Frequency     float64
	PhaseShift    float64
	VerticalShift float64
	Angle         float64
}

// ZeroAngleY returns the result.
func (p *SineWave) ZeroAngleY(x float64) float64 {
	return p.Amplitude*math.Sin(p.Frequency*x+p.PhaseShift) + p.VerticalShift
}

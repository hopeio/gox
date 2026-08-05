package image

import (
	"errors"
	"image"
)

type BitMask struct {
	Data []uint8
	Rect image.Rectangle
}

// NewBitMask creates and returns a new instance.
func NewBitMask(rect image.Rectangle) *BitMask {
	total := rect.Dx() * rect.Dy()
	n := total / 8
	if n*8 < total {
		n++
	}
	return &BitMask{
		Rect: rect,
		Data: make([]uint8, n),
	}
}

// Set updates or inserts a value.
func (m *BitMask) Set(x, y int, v bool) error {
	if !(image.Point{X: x, Y: y}).In(m.Rect) {
		return errors.New("out of range")
	}
	total := m.Rect.Dx()*(y-m.Rect.Min.Y) + x - m.Rect.Min.X
	n := total / 8
	bit := 8 - (total - n*8)
	if bit == 0 {
		n--
	}
	bit--
	if v {
		m.Data[n] |= 1 << bit
	} else {
		m.Data[n] &= ^(1 << bit)
	}
	return nil
}

// Get returns the value.
func (m *BitMask) Get(x, y int) (bool, bool) {
	if !(image.Point{X: x, Y: y}).In(m.Rect) {
		return false, false
	}
	total := m.Rect.Dx()*(y-m.Rect.Min.Y) + x - m.Rect.Min.X
	n := total / 8
	bit := 8 - (total - n*8)
	if bit == 0 {
		n--
	}
	bit--
	return m.Data[n]&(1<<bit) != 0, true
}

type Mask struct {
	Data []uint8
	Rect image.Rectangle
}

// NewMask creates and returns a new instance.
func NewMask(rect image.Rectangle) *Mask {
	return &Mask{
		Rect: rect,
		Data: make([]uint8, rect.Dx()*rect.Dy()),
	}
}

// Set updates or inserts a value.
func (m *Mask) Set(x, y int, v uint8) error {
	if !(image.Point{X: x, Y: y}).In(m.Rect) {
		return errors.New("out of range")
	}
	n := m.Rect.Dx()*(y-m.Rect.Min.Y) + x - m.Rect.Min.X
	m.Data[n] = v
	return nil
}

// Get returns the value.
func (m *Mask) Get(x, y int) (uint8, bool) {
	if !(image.Point{X: x, Y: y}).In(m.Rect) {
		return 0, false
	}
	n := m.Rect.Dx()*(y-m.Rect.Min.Y) + x - m.Rect.Min.X
	return m.Data[n], true
}

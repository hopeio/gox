/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package gerber

import "log"

type LogProcessor struct {
}

// Circle performs the operation.
func (l LogProcessor) Circle(circle *Circle) {
	log.Println("circle", circle)
}

// Rectangle performs the operation.
func (l LogProcessor) Rectangle(rectangle *Rectangle) {
	log.Println("rectangle", rectangle)
}

// Obround performs the operation.
func (l LogProcessor) Obround(obround *Obround) {
	log.Println("obround", obround)
}

// Contour performs the operation.
func (l LogProcessor) Contour(contour *Contour) {
	log.Println("contour", contour)
}

// Line performs the operation.
func (l LogProcessor) Line(line *Line) {
	log.Println("line", line)
}

// Arc performs the operation.
func (l LogProcessor) Arc(arc *Arc) {
	log.Println("arc", arc)
}

// SetViewBox updates or inserts a value.
func (l LogProcessor) SetViewBox(box *ViewBox) {
	log.Println("SetViewBox", box)
}

var _ Processor = (*LogProcessor)(nil)

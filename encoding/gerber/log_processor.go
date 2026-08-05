/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package gerber

import "log"

type LogProcessor struct {
}

// Circle ...
func (l LogProcessor) Circle(circle *Circle) {
	log.Println("circle", circle)
}

// Rectangle ...
func (l LogProcessor) Rectangle(rectangle *Rectangle) {
	log.Println("rectangle", rectangle)
}

// Obround ...
func (l LogProcessor) Obround(obround *Obround) {
	log.Println("obround", obround)
}

// Contour ...
func (l LogProcessor) Contour(contour *Contour) {
	log.Println("contour", contour)
}

// Line ...
func (l LogProcessor) Line(line *Line) {
	log.Println("line", line)
}

// Arc ...
func (l LogProcessor) Arc(arc *Arc) {
	log.Println("arc", arc)
}

// SetViewBox ...
func (l LogProcessor) SetViewBox(box *ViewBox) {
	log.Println("SetViewBox", box)
}

var _ Processor = (*LogProcessor)(nil)

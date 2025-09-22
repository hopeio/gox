// Package svg parses Gerber to SVG.
package svg

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/hopeio/gox/encoding/gerber"
	jsonx "github.com/hopeio/gox/encoding/json"

	"github.com/mitchellh/mapstructure"
)

func (p *Processor) UnmarshalJSON(b []byte) error {
	// Skip '{'.
	if len(b) < 1 {
		return fmt.Errorf("%d", len(b))
	}
	residue := b[1:]

	for {
		if len(residue) == 0 || bytes.Equal(residue, []byte{'\n'}) {
			break
		}
		closingK := bytes.Index(residue[1:], []byte(`":`))
		if closingK == -1 {
			return fmt.Errorf("\"%s\" %#v", residue, residue)
		}
		key := string(residue[1 : 1+closingK])
		residue = residue[1+closingK+2:]
		var err error
		switch key {
		case "Data":
			residue, err = p.decodeData(residue)
		case "MinX":
			p.Min.X, residue, err = jsonx.DecodeFloat(residue)
		case "MaxX":
			p.Max.X, residue, err = jsonx.DecodeFloat(residue)
		case "MinY":
			p.Min.Y, residue, err = jsonx.DecodeFloat(residue)
		case "MaxY":
			p.Max.Y, residue, err = jsonx.DecodeFloat(residue)
		case "PolarityDark":
			p.PolarityDark, residue, err = jsonx.DecodeString(residue)
		case "PolarityClear":
			p.PolarityClear, residue, err = jsonx.DecodeString(residue)
		case "Scale":
			p.Scale, residue, err = jsonx.DecodeFloat(residue)
		case "StrokeWidth":
			p.Width, residue, err = jsonx.DecodeString(residue)
		case "Height":
			p.Height, residue, err = jsonx.DecodeString(residue)
		case "PanZoom":
			p.PanZoom, residue, err = jsonx.DecodeBool(residue)
		default:
			err = fmt.Errorf("\"%s\"", key)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Processor) UnmarshalJSON_1(b []byte) error {
	pmap := make(map[string]interface{})
	if err := json.Unmarshal(b, &pmap); err != nil {
		return err
	}

	if err := mapstructure.Decode(pmap, p); err != nil {
		return err
	}

	data := make([]interface{}, 0, len(p.Data))
	for i, d := range p.Data {
		m, ok := d.(map[string]interface{})
		if !ok {
			return fmt.Errorf("%d %+v", i, d)
		}
		eType, ok := m["Type"].(string)
		if !ok {
			return fmt.Errorf("%d %+v", i, d)
		}
		switch ElementType(eType) {
		case ElementTypeCircle:
			e := Circle{}
			if err := mapstructure.Decode(m, &e); err != nil {
				return fmt.Errorf("%d %+v %w", i, d, err)
			}
			data = append(data, e)
		case ElementTypeRectangle:
			e := Rectangle{}
			if err := mapstructure.Decode(m, &e); err != nil {
				return fmt.Errorf("%d %+v %w", i, d, err)
			}
			data = append(data, e)
		case ElementTypePath:
			e := Path{}
			if err := mapstructure.Decode(m, &e); err != nil {
				return fmt.Errorf("%d %+v %w", i, d, err)
			}
			cmds := make([]interface{}, 0, len(e.Commands))
			for j, cj := range e.Commands {
				c, ok := cj.(map[string]interface{})
				if !ok {
					return fmt.Errorf("%d %d %+v %+v", i, j, cj, d)
				}
				cType, ok := c["Type"].(string)
				if !ok {
					return fmt.Errorf("%d %d %+v %+v", i, j, cj, d)
				}
				switch ElementType(cType) {
				case ElementTypeLine:
					cmd := PathLine{}
					if err := mapstructure.Decode(c, &cmd); err != nil {
						return fmt.Errorf("%d %d %+v %+v", i, j, c, d)
					}
					cmds = append(cmds, cmd)
				case ElementTypeArc:
					cmd := PathArc{}
					if err := mapstructure.Decode(c, &cmd); err != nil {
						return fmt.Errorf("%d %d %+v %+v", i, j, c, d)
					}
					cmds = append(cmds, cmd)
				default:
					return fmt.Errorf("%d %d %+v %+v", i, j, c, d)
				}
			}
			e.Commands = cmds
			data = append(data, e)
		case ElementTypeLine:
			e := Line{}
			if err := mapstructure.Decode(m, &e); err != nil {
				return fmt.Errorf("%d %+v %w", i, d, err)
			}
			data = append(data, e)
		case ElementTypeArc:
			e := Arc{}
			if err := mapstructure.Decode(m, &e); err != nil {
				return fmt.Errorf("%d %+v %w", i, d, err)
			}
			data = append(data, e)
		default:
			return fmt.Errorf("%d %+v", i, m)
		}
	}
	p.Data = data

	return nil
}

func (p *Processor) decodeData(b []byte) ([]byte, error) {
	// Empty array.
	if bytes.HasPrefix(b, []byte("null")) {
		b = b[4:]
		if b[0] == ',' {
			b = b[1:]
		}
		return b, nil
	}

	// Opening '['.
	if len(b) < 1 {
		return b, fmt.Errorf("%d", len(b))
	}
	b = b[1:]

	for {
		segment, err := findSegment(b)
		if err != nil {
			return b, err
		}
		if err := p.decodeSegment(segment); err != nil {
			return b, err
		}
		if b[len(segment)] == ']' {
			b = b[len(segment)+1:]
			if b[0] == ',' {
				b = b[1:]
			}
			break
		}
		if len(segment)+1 >= len(b) {
			return b, fmt.Errorf("%d %d \"%s\"", len(segment)+1, len(b), segment)
		}
		b = b[len(segment)+1:]
	}

	return b, nil
}

func findSegment(bs []byte) ([]byte, error) {
	level := 0
	end := -1
Loop:
	for i, b := range bs {
		switch b {
		case '}':
			level--
			if level == 0 {
				end = i
				break Loop
			}
		case '{':
			level++
		}
	}
	if end == -1 {
		return nil, fmt.Errorf("not closed")
	}
	return bs[:end+1], nil
}

func (p *Processor) decodeSegment(b []byte) error {
	elmType, err := findElementType(b)
	if err != nil {
		return err
	}
	switch elmType {
	case ElementTypeCircle:
		var c Circle
		if err := decodeCircle(b, &c); err != nil {
			return err
		}
		p.Data = append(p.Data, c)
	case ElementTypeRectangle:
		var r Rectangle
		if err := decodeRectangle(b, &r); err != nil {
			return err
		}
		p.Data = append(p.Data, r)
	case ElementTypePath:
		var ph Path
		if err := decodePath(b, &ph); err != nil {
			return err
		}
		p.Data = append(p.Data, ph)
	case ElementTypeLine:
		var l Line
		if err := decodeLine(b, &l); err != nil {
			return err
		}
		p.Data = append(p.Data, l)
	case ElementTypeArc:
		var a Arc
		if err := decodeArc(b, &a); err != nil {
			return err
		}
		p.Data = append(p.Data, a)
	default:
		return fmt.Errorf("\"%s\"", elmType)
	}
	return nil
}

func findElementType(bs []byte) (ElementType, error) {
	level := -1
	for i, b := range bs {
		switch b {
		case '}':
			level--
		case '{':
			level++
		case ':':
			if level != 0 {
				break
			}

			if i-5 < 0 {
				break
			}
			if !bytes.Equal(bs[i-5:i-1], []byte("Type")) {
				break
			}
			endIdx := bytes.IndexByte(bs[i+2:], '"')
			if endIdx == -1 {
				return "", fmt.Errorf("not closed")
			}
			return ElementType(bs[i+2 : i+2+endIdx]), nil
		}
	}
	return "", fmt.Errorf("not found")
}

func decodeCircle(b []byte, elm *Circle) error {
	// Skip '{'.
	if len(b) < 1 {
		return fmt.Errorf("%d", len(b))
	}
	residue := b[1:]

	for {
		if len(residue) == 0 {
			break
		}
		closingK := bytes.Index(residue[1:], []byte(`":`))
		if closingK == -1 {
			return fmt.Errorf("\"%s\"", residue)
		}
		key := string(residue[1 : 1+closingK])
		residue = residue[1+closingK+2:]
		var err error
		switch key {
		case "Type":
			var elmType string
			elmType, residue, err = jsonx.DecodeString(residue)
			elm.Type = ElementType(elmType)
		case "Line":
			elm.Line, residue, err = jsonx.DecodeInt(residue)
		case "X":
			elm.Centre.X, residue, err = jsonx.DecodeFloat(residue)
		case "Y":
			elm.Centre.Y, residue, err = jsonx.DecodeFloat(residue)
		case "Diameter":
			elm.Diameter, residue, err = jsonx.DecodeFloat(residue)
		case "Fill":
			elm.Fill, residue, err = jsonx.DecodeString(residue)
		case "Attr":
			// Attr is expected to be always null.
			residue = residue[5:]
		default:
			err = fmt.Errorf("\"%s\"", key)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func decodeRectangle(b []byte, elm *Rectangle) error {
	// Skip '{'.
	if len(b) < 1 {
		return fmt.Errorf("%d", len(b))
	}
	residue := b[1:]

	for {
		if len(residue) == 0 {
			break
		}
		closingK := bytes.Index(residue[1:], []byte(`":`))
		if closingK == -1 {
			return fmt.Errorf("\"%s\"", residue)
		}
		key := string(residue[1 : 1+closingK])
		residue = residue[1+closingK+2:]
		var err error
		switch key {
		case "Type":
			var elmType string
			elmType, residue, err = jsonx.DecodeString(residue)
			elm.Type = ElementType(elmType)
		case "Line":
			elm.Line, residue, err = jsonx.DecodeInt(residue)
		case "Aperture":
			elm.Aperture, residue, err = jsonx.DecodeString(residue)
		case "X":
			elm.Center.X, residue, err = jsonx.DecodeFloat(residue)
		case "Y":
			elm.Center.Y, residue, err = jsonx.DecodeFloat(residue)
		case "StrokeWidth":
			elm.Width, residue, err = jsonx.DecodeFloat(residue)
		case "Height":
			elm.Height, residue, err = jsonx.DecodeFloat(residue)
		case "CenterX":
			elm.RadiusX, residue, err = jsonx.DecodeFloat(residue)
		case "CenterY":
			elm.RadiusY, residue, err = jsonx.DecodeFloat(residue)
		case "Fill":
			elm.Fill, residue, err = jsonx.DecodeString(residue)
		case "Attr":
			// Attr is expected to be always null.
			residue = residue[5:]
		default:
			err = fmt.Errorf("\"%s\"", key)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func decodePath(b []byte, elm *Path) error {
	// Skip '{'.
	if len(b) < 1 {
		return fmt.Errorf("%d", len(b))
	}
	residue := b[1:]

	for {
		if len(residue) == 0 {
			break
		}
		closingK := bytes.Index(residue[1:], []byte(`":`))
		if closingK == -1 {
			return fmt.Errorf("\"%s\"", residue)
		}
		key := string(residue[1 : 1+closingK])
		residue = residue[1+closingK+2:]
		var err error
		switch key {
		case "Type":
			var elmType string
			elmType, residue, err = jsonx.DecodeString(residue)
			elm.Type = ElementType(elmType)
		case "Line":
			elm.Line, residue, err = jsonx.DecodeInt(residue)
		case "X":
			elm.X, residue, err = jsonx.DecodeFloat(residue)
		case "Y":
			elm.Y, residue, err = jsonx.DecodeFloat(residue)
		case "Commands":
			residue, err = decodePathCommands(elm, residue)
		case "Fill":
			elm.Fill, residue, err = jsonx.DecodeString(residue)
		case "Attr":
			// Attr is expected to be always null.
			residue = residue[5:]
		default:
			err = fmt.Errorf("\"%s\"", key)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func decodeLine(b []byte, elm *Line) error {
	// Skip '{'.
	if len(b) < 1 {
		return fmt.Errorf("%d", len(b))
	}
	residue := b[1:]

	for {
		if len(residue) == 0 {
			break
		}
		closingK := bytes.Index(residue[1:], []byte(`":`))
		if closingK == -1 {
			return fmt.Errorf("\"%s\"", residue)
		}
		key := string(residue[1 : 1+closingK])
		residue = residue[1+closingK+2:]
		var err error
		switch key {
		case "Type":
			var elmType string
			elmType, residue, err = jsonx.DecodeString(residue)
			elm.Type = ElementType(elmType)
		case "Line":
			elm.Line.LineNo, residue, err = jsonx.DecodeInt(residue)
		case "StartX":
			elm.Start.X, residue, err = jsonx.DecodeFloat(residue)
		case "StartY":
			elm.Start.Y, residue, err = jsonx.DecodeFloat(residue)
		case "EndX":
			elm.End.X, residue, err = jsonx.DecodeFloat(residue)
		case "EndY":
			elm.End.Y, residue, err = jsonx.DecodeFloat(residue)
		case "StrokeWidth":
			elm.StrokeWidth, residue, err = jsonx.DecodeFloat(residue)
		case "Cap":
			var lineCap string
			lineCap, residue, err = jsonx.DecodeString(residue)
			elm.Cap = gerber.LineCap(lineCap)
		case "Stroke":
			elm.Stroke, residue, err = jsonx.DecodeString(residue)
		case "Attr":
			// Attr is expected to be always null.
			residue = residue[5:]
		default:
			err = fmt.Errorf("\"%s\"", key)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func decodeArc(b []byte, elm *Arc) error {
	// Skip '{'.
	if len(b) < 1 {
		return fmt.Errorf("%d", len(b))
	}
	residue := b[1:]

	for {
		if len(residue) == 0 {
			break
		}
		closingK := bytes.Index(residue[1:], []byte(`":`))
		if closingK == -1 {
			return fmt.Errorf("\"%s\"", residue)
		}
		key := string(residue[1 : 1+closingK])
		residue = residue[1+closingK+2:]
		var err error
		switch key {
		case "Type":
			var elmType string
			elmType, residue, err = jsonx.DecodeString(residue)
			elm.Type = ElementType(elmType)
		case "Line":
			elm.Line, residue, err = jsonx.DecodeInt(residue)
		case "StartX":
			elm.Start.X, residue, err = jsonx.DecodeFloat(residue)
		case "StartY":
			elm.Start.Y, residue, err = jsonx.DecodeFloat(residue)
		case "XRadius":
			elm.RadiusX, residue, err = jsonx.DecodeFloat(residue)
		case "YRadius":
			elm.RadiusY, residue, err = jsonx.DecodeFloat(residue)
		case "LargeArc":
			elm.LargeArc, residue, err = jsonx.DecodeInt(residue)
		case "Sweep":
			elm.Sweep, residue, err = jsonx.DecodeInt(residue)
		case "EndX":
			elm.End.X, residue, err = jsonx.DecodeFloat(residue)
		case "EndY":
			elm.End.Y, residue, err = jsonx.DecodeFloat(residue)
		case "StrokeWidth":
			elm.StrokeWidth, residue, err = jsonx.DecodeFloat(residue)
		case "CenterX":
			elm.Center.X, residue, err = jsonx.DecodeFloat(residue)
		case "CenterY":
			elm.Center.Y, residue, err = jsonx.DecodeFloat(residue)
		case "Stroke":
			elm.Stroke, residue, err = jsonx.DecodeString(residue)
		case "Attr":
			// Attr is expected to be always null.
			residue = residue[5:]
		default:
			err = fmt.Errorf("\"%s\"", key)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func decodePathLine(b []byte, elm *PathLine) error {
	// Skip '{'.
	if len(b) < 1 {
		return fmt.Errorf("%d", len(b))
	}
	residue := b[1:]

	for {
		if len(residue) == 0 {
			break
		}
		closingK := bytes.Index(residue[1:], []byte(`":`))
		if closingK == -1 {
			return fmt.Errorf("\"%s\"", residue)
		}
		key := string(residue[1 : 1+closingK])
		residue = residue[1+closingK+2:]
		var err error
		switch key {
		case "Type":
			var elmType string
			elmType, residue, err = jsonx.DecodeString(residue)
			elm.Type = ElementType(elmType)
		case "X":
			elm.X, residue, err = jsonx.DecodeFloat(residue)
		case "Y":
			elm.Y, residue, err = jsonx.DecodeFloat(residue)
		default:
			err = fmt.Errorf("\"%s\"", key)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func decodePathArc(b []byte, elm *PathArc) error {
	// Skip '{'.
	if len(b) < 1 {
		return fmt.Errorf("%d", len(b))
	}
	residue := b[1:]

	for {
		if len(residue) == 0 {
			break
		}
		closingK := bytes.Index(residue[1:], []byte(`":`))
		if closingK == -1 {
			return fmt.Errorf("\"%s\"", residue)
		}
		key := string(residue[1 : 1+closingK])
		residue = residue[1+closingK+2:]
		var err error
		switch key {
		case "Type":
			var elmType string
			elmType, residue, err = jsonx.DecodeString(residue)
			elm.Type = ElementType(elmType)
		case "LargeArc":
			elm.LargeArc, residue, err = jsonx.DecodeInt(residue)
		case "Sweep":
			elm.Sweep, residue, err = jsonx.DecodeInt(residue)
		case "X":
			elm.X, residue, err = jsonx.DecodeFloat(residue)
		case "Y":
			elm.Y, residue, err = jsonx.DecodeFloat(residue)
		default:
			err = fmt.Errorf("\"%s\"", key)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func decodePathCommands(ph *Path, b []byte) ([]byte, error) {
	// Opening '['.
	if len(b) < 1 {
		return b, fmt.Errorf("%d", len(b))
	}
	b = b[1:]

	for {
		segment, err := findSegment(b)
		if err != nil {
			return b, err
		}
		if err := decodePathSegment(ph, segment); err != nil {
			return b, err
		}
		if b[len(segment)] == ']' {
			b = b[len(segment)+1:]
			if b[0] == ',' {
				b = b[1:]
			}
			break
		}
		if len(segment)+1 >= len(b) {
			return b, fmt.Errorf("%d %d \"%s\"", len(segment)+1, len(b), segment)
		}
		b = b[len(segment)+1:]
	}

	return b, nil
}

func decodePathSegment(ph *Path, b []byte) error {
	elmType, err := findElementType(b)
	if err != nil {
		return err
	}
	switch elmType {
	case ElementTypeLine:
		var l PathLine
		if err := decodePathLine(b, &l); err != nil {
			return err
		}
		ph.Commands = append(ph.Commands, l)
	case ElementTypeArc:
		var a PathArc
		if err := decodePathArc(b, &a); err != nil {
			return err
		}
		ph.Commands = append(ph.Commands, a)
	default:
		return fmt.Errorf("\"%s\"", elmType)
	}
	return nil
}

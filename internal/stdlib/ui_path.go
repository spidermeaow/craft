package stdlib

import (
	rt "craft/internal/runtime"
	"fmt"
	"math"
)

// Path commands use Canvas-local device-independent pixels. Arcs are described
// by a center, radius, start angle and signed sweep in degrees. Backends share
// this conversion so their geometry and subpath boundaries remain identical.
type uiPathOp struct {
	kind string
	v    [6]float64
}

type uiPathPoint struct{ x, y float64 }
type uiPathEdge struct {
	cubic       bool
	p1, p2, end uiPathPoint
}
type uiPathFigure struct {
	start  uiPathPoint
	edges  []uiPathEdge
	closed bool
}

func validPathCoordinate(v int64) bool { return v >= -8192 && v <= 8192 }

func pathSegment(name string, args []any) (*UiDraw, error) {
	d := &UiDraw{kind: name}
	if name == "std.ui.pathClose" {
		return d, nil
	}
	for i, value := range args {
		n := value.(int64)
		if i >= len(d.pathValues) || !validPathCoordinate(n) {
			return nil, fmt.Errorf("UiPathError: coordinate outside -8192..8192")
		}
		d.pathValues[i] = float64(n)
	}
	if name == "std.ui.pathArcTo" {
		radius, start, sweep := args[2].(int64), args[3].(int64), args[4].(int64)
		if radius < 1 || radius > 8192 || start < -3600 || start > 3600 || sweep < -360 || sweep > 360 || sweep == 0 {
			return nil, fmt.Errorf("UiPathError: invalid arc radius or angles")
		}
	}
	return d, nil
}

func pathDraw(name string, args []any) (*UiDraw, error) {
	items := args[0].(*rt.Array).Items
	if len(items) == 0 || len(items) > 1024 {
		return nil, fmt.Errorf("UiPathError: path needs 1..1024 commands")
	}
	color := args[1].(int64)
	if !validCanvasColor(color) || color>>24 != 255 {
		return nil, fmt.Errorf("UiPathError: path color must be opaque for Direct2D/GDI parity")
	}
	d := &UiDraw{kind: "fillPath", color: uint32(color)}
	if name == "std.ui.strokePath" {
		d.kind = "strokePath"
		d.thickness = args[2].(int64)
		if d.thickness < 1 || d.thickness > 256 {
			return nil, fmt.Errorf("UiPathError: stroke width must be 1..256")
		}
	}
	for _, item := range items {
		segment, ok := item.(*UiDraw)
		if !ok || segment == nil || !isPathSegment(segment.kind) {
			return nil, fmt.Errorf("UiPathError: path contains a non-path command")
		}
		d.path = append(d.path, uiPathOp{kind: segment.kind, v: segment.pathValues})
	}
	figures, err := buildPathFigures(d.path)
	if err != nil {
		return nil, err
	}
	d.figures = figures
	return d, nil
}

func isPathSegment(kind string) bool {
	switch kind {
	case "std.ui.pathMoveTo", "std.ui.pathLineTo", "std.ui.pathArcTo", "std.ui.pathQuadraticTo", "std.ui.pathCubicTo", "std.ui.pathClose":
		return true
	}
	return false
}

func buildPathFigures(ops []uiPathOp) ([]uiPathFigure, error) {
	var figures []uiPathFigure
	var current uiPathFigure
	var point uiPathPoint
	active := false
	flush := func(closed bool) {
		if active && len(current.edges) > 0 {
			current.closed = closed
			figures = append(figures, current)
		}
		current = uiPathFigure{}
		active = false
	}
	begin := func(p uiPathPoint) {
		current = uiPathFigure{start: p}
		point = p
		active = true
	}
	line := func(p uiPathPoint) {
		current.edges = append(current.edges, uiPathEdge{end: p})
		point = p
	}
	cubic := func(p1, p2, end uiPathPoint) {
		current.edges = append(current.edges, uiPathEdge{cubic: true, p1: p1, p2: p2, end: end})
		point = end
	}
	for _, op := range ops {
		v := op.v
		switch op.kind {
		case "std.ui.pathMoveTo":
			flush(false)
			begin(uiPathPoint{v[0], v[1]})
		case "std.ui.pathLineTo":
			if !active {
				return nil, fmt.Errorf("UiPathError: lineTo requires moveTo")
			}
			line(uiPathPoint{v[0], v[1]})
		case "std.ui.pathQuadraticTo":
			if !active {
				return nil, fmt.Errorf("UiPathError: quadraticTo requires moveTo")
			}
			control, end := uiPathPoint{v[0], v[1]}, uiPathPoint{v[2], v[3]}
			cubic(uiPathPoint{point.x + 2*(control.x-point.x)/3, point.y + 2*(control.y-point.y)/3}, uiPathPoint{end.x + 2*(control.x-end.x)/3, end.y + 2*(control.y-end.y)/3}, end)
		case "std.ui.pathCubicTo":
			if !active {
				return nil, fmt.Errorf("UiPathError: cubicTo requires moveTo")
			}
			cubic(uiPathPoint{v[0], v[1]}, uiPathPoint{v[2], v[3]}, uiPathPoint{v[4], v[5]})
		case "std.ui.pathArcTo":
			start := v[3] * math.Pi / 180
			sweep := v[4] * math.Pi / 180
			first := uiPathPoint{v[0] + v[2]*math.Cos(start), v[1] + v[2]*math.Sin(start)}
			if !active {
				begin(first)
			} else if math.Hypot(point.x-first.x, point.y-first.y) > 0.001 {
				line(first)
			}
			steps := int(math.Ceil(math.Abs(sweep) / (math.Pi / 2)))
			for i := 0; i < steps; i++ {
				a0 := start + sweep*float64(i)/float64(steps)
				a1 := start + sweep*float64(i+1)/float64(steps)
				k := 4 * math.Tan((a1-a0)/4) / 3
				p1 := uiPathPoint{v[0] + v[2]*(math.Cos(a0)-k*math.Sin(a0)), v[1] + v[2]*(math.Sin(a0)+k*math.Cos(a0))}
				p2 := uiPathPoint{v[0] + v[2]*(math.Cos(a1)+k*math.Sin(a1)), v[1] + v[2]*(math.Sin(a1)-k*math.Cos(a1))}
				end := uiPathPoint{v[0] + v[2]*math.Cos(a1), v[1] + v[2]*math.Sin(a1)}
				cubic(p1, p2, end)
			}
		case "std.ui.pathClose":
			if !active {
				return nil, fmt.Errorf("UiPathError: closePath requires an open subpath")
			}
			flush(true)
		}
		if len(figures) > 128 {
			return nil, fmt.Errorf("UiPathError: too many subpaths")
		}
	}
	flush(false)
	if len(figures) == 0 {
		return nil, fmt.Errorf("UiPathError: path has no drawable segments")
	}
	return figures, nil
}

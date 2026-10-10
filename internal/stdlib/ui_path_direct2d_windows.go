//go:build windows

package stdlib

import (
	"math"
	"runtime"
	"unsafe"
)

type d2dBezier struct {
	p1, p2, end struct{ x, y float32 }
}

// Path geometry is rebuilt from immutable Canvas commands. The factory owns
// geometry resources, while each temporary geometry and sink is released in
// this paint call. Canvas and outer render clips are already active.
func (d *direct2D) paintPath(box *uiBox, cmd UiDraw, scale float64) bool {
	if d == nil || d.factory == 0 || d.target == 0 || len(cmd.figures) == 0 {
		return false
	}
	var geometry uintptr
	if hr := comCall(d.factory, 10, uintptr(unsafe.Pointer(&geometry))); failedHR(hr) || geometry == 0 {
		return false
	}
	defer releaseCOM(&geometry)
	var sink uintptr
	// ID2D1PathGeometry::Open follows ID2D1Geometry::Widen (slot 16).
	// Calling slot 16 here fails path creation and forces the whole frame to GDI.
	if hr := comCall(geometry, 17, uintptr(unsafe.Pointer(&sink))); failedHR(hr) || sink == 0 {
		return false
	}
	comCall(sink, 3, 0) // D2D1_FILL_MODE_ALTERNATE: holes across subpaths.
	for _, figure := range cmd.figures {
		start := d2dPoint(float32(float64(box.x)+figure.start.x), float32(float64(box.y)+figure.start.y))
		comCall(sink, 5, start, 0) // BeginFigure, filled
		for _, edge := range figure.edges {
			if edge.cubic {
				segment := d2dBezier{}
				segment.p1.x, segment.p1.y = float32(float64(box.x)+edge.p1.x), float32(float64(box.y)+edge.p1.y)
				segment.p2.x, segment.p2.y = float32(float64(box.x)+edge.p2.x), float32(float64(box.y)+edge.p2.y)
				segment.end.x, segment.end.y = float32(float64(box.x)+edge.end.x), float32(float64(box.y)+edge.end.y)
				comCall(sink, 11, uintptr(unsafe.Pointer(&segment)))
				runtime.KeepAlive(segment)
			} else {
				comCall(sink, 10, d2dPoint(float32(float64(box.x)+edge.end.x), float32(float64(box.y)+edge.end.y)))
			}
		}
		if figure.closed {
			comCall(sink, 8, 1)
		} else {
			comCall(sink, 8, 0)
		}
	}
	hr := comCall(sink, 9)
	releaseCOM(&sink)
	if failedHR(hr) {
		return false
	}
	brush := d.brushARGB(cmd.color)
	if brush == 0 {
		return false
	}
	if cmd.kind == "fillPath" {
		comCall(d.target, 23, geometry, brush, 0)
	} else {
		comCall(d.target, 22, geometry, brush, uintptr(math.Float32bits(float32(d2dStrokePixels(cmd.thickness, scale)/scale))), 0)
	}
	return true
}

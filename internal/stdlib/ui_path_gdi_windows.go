//go:build windows

package stdlib

import (
	"math"
	"runtime"
	"unsafe"
)

var (
	uiBeginPathGDI  = uiGdi32.NewProc("BeginPath")
	uiEndPathGDI    = uiGdi32.NewProc("EndPath")
	uiCloseFigure   = uiGdi32.NewProc("CloseFigure")
	uiPolyBezierTo  = uiGdi32.NewProc("PolyBezierTo")
	uiFillPathGDI   = uiGdi32.NewProc("FillPath")
	uiStrokePathGDI = uiGdi32.NewProc("StrokePath")
	uiSetPolyFill   = uiGdi32.NewProc("SetPolyFillMode")
)

func paintPathGDI(hdc uintptr, box *uiBox, cmd UiDraw, scale float64) {
	if hdc == 0 || box == nil || len(cmd.figures) == 0 {
		return
	}
	if scale <= 0 {
		scale = 1
	}
	toPoint := func(p uiPathPoint) windowsPoint {
		return windowsPoint{int32(math.Round((float64(box.x) + p.x) * scale)), int32(math.Round((float64(box.y) + p.y) * scale))}
	}
	uiSetPolyFill.Call(hdc, 1) // ALTERNATE for holes and multiple subpaths.
	if ok, _, _ := uiBeginPathGDI.Call(hdc); ok == 0 {
		return
	}
	for _, figure := range cmd.figures {
		start := toPoint(figure.start)
		uiMoveToEx.Call(hdc, uintptr(start.x), uintptr(start.y), 0)
		for _, edge := range figure.edges {
			if edge.cubic {
				points := [3]windowsPoint{toPoint(edge.p1), toPoint(edge.p2), toPoint(edge.end)}
				uiPolyBezierTo.Call(hdc, uintptr(unsafe.Pointer(&points[0])), 3)
				runtime.KeepAlive(points)
			} else {
				point := toPoint(edge.end)
				uiLineTo.Call(hdc, uintptr(point.x), uintptr(point.y))
			}
		}
		if figure.closed {
			uiCloseFigure.Call(hdc)
		}
	}
	if ok, _, _ := uiEndPathGDI.Call(hdc); ok == 0 {
		return
	}
	color := gdiColor(cmd.color)
	if cmd.kind == "fillPath" {
		brush, _, _ := uiCreateBrush.Call(color)
		if brush == 0 {
			return
		}
		old, _, _ := uiSelectObject.Call(hdc, brush)
		uiFillPathGDI.Call(hdc)
		uiSelectObject.Call(hdc, old)
		uiDeleteObject.Call(brush)
	} else {
		width := max64(1, int64(math.Round(float64(cmd.thickness)*scale)))
		pen, _, _ := uiCreatePen.Call(0, uintptr(width), color)
		if pen == 0 {
			return
		}
		old, _, _ := uiSelectObject.Call(hdc, pen)
		uiStrokePathGDI.Call(hdc)
		uiSelectObject.Call(hdc, old)
		uiDeleteObject.Call(pen)
	}
}

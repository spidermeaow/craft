//go:build windows

package stdlib

import (
	"math"
	"unsafe"
)

// uiDirtyRegions owns paint invalidation for retained geometry. All rectangles
// are window-client physical pixels; the paint routine clips every draw to the
// resulting Win32 update region. Uncertain geometry falls back to full redraw.
type uiDirtyRegions struct {
	w       *windowsState
	regions []windowsRect
	full    bool
}

func newUIDirtyRegions(w *windowsState) *uiDirtyRegions { return &uiDirtyRegions{w: w} }

func (d *uiDirtyRegions) addLogical(x, y, width, height int64) {
	if d.full || d.w == nil || width <= 0 || height <= 0 {
		return
	}
	scale := d.w.scale
	if scale <= 0 {
		scale = 1
	}
	r := windowsRect{
		left:   int32(math.Floor(float64(x)*scale)) - 2,
		top:    int32(math.Floor(float64(y)*scale)) - 2,
		right:  int32(math.Ceil(float64(x+width)*scale)) + 2,
		bottom: int32(math.Ceil(float64(y+height)*scale)) + 2,
	}
	maxX, maxY := int32(math.Ceil(float64(d.w.width)*scale)), int32(math.Ceil(float64(d.w.height)*scale))
	if r.left < 0 {
		r.left = 0
	}
	if r.top < 0 {
		r.top = 0
	}
	if r.right > maxX {
		r.right = maxX
	}
	if r.bottom > maxY {
		r.bottom = maxY
	}
	if r.right <= r.left || r.bottom <= r.top {
		return
	}
	for i := 0; i < len(d.regions); {
		old := d.regions[i]
		if r.right+2 < old.left || old.right+2 < r.left || r.bottom+2 < old.top || old.bottom+2 < r.top {
			i++
			continue
		}
		if old.left < r.left {
			r.left = old.left
		}
		if old.top < r.top {
			r.top = old.top
		}
		if old.right > r.right {
			r.right = old.right
		}
		if old.bottom > r.bottom {
			r.bottom = old.bottom
		}
		d.regions = append(d.regions[:i], d.regions[i+1:]...)
		i = 0 // The expanded union may now touch an earlier region.
	}
	d.regions = append(d.regions, r)
	if len(d.regions) > 32 {
		d.full = true
	}
}

func (d *uiDirtyRegions) addBox(box *uiBox) {
	if box != nil {
		d.addLogical(box.x, box.y, box.w, box.h)
	}
}

func (d *uiDirtyRegions) addChange(change uiChange) {
	box, old := change.box, change.old
	if box == nil || old == nil {
		d.full = true
		return
	}
	if box.node.kind == "Canvas" {
		if !d.addCanvasChanges(box, old.draws, box.node.draws) {
			d.addBox(box)
		}
		return
	}
	if box.node.kind == "VirtualGrid" {
		before := *box
		before.node = old
		for _, selected := range []*uiBox{&before, box} {
			if x, y, w, h, ok := virtualGridSelectionRect(selected); ok {
				d.addLogical(x, y, w, h)
			}
			if x, y, w, h, ok := virtualGridRangeRect(selected); ok {
				d.addLogical(x, y, w, h)
			}
		}
		return
	}
	d.addBox(box)
}

func (d *uiDirtyRegions) addCanvasChanges(box *uiBox, old, next []UiDraw) bool {
	if len(old) != len(next) {
		return false
	}
	changed := 0
	for i := range old {
		if sameDraws(old[i:i+1], next[i:i+1]) {
			continue
		}
		changed++
		if changed > 64 {
			return false
		}
		for _, draw := range []UiDraw{old[i], next[i]} {
			x, y, w, h, ok := canvasDrawBounds(draw)
			if !ok && draw.kind != "hitRegion" {
				return false
			}
			if ok {
				left, top := max64(0, x), max64(0, y)
				right, bottom := min64(box.w, x+w), min64(box.h, y+h)
				d.addLogical(box.x+left, box.y+top, right-left, bottom-top)
			}
		}
	}
	return true
}

func canvasDrawBounds(draw UiDraw) (int64, int64, int64, int64, bool) {
	switch draw.kind {
	case "fillRect", "image", "text":
		return draw.x, draw.y, draw.w, draw.h, true
	case "strokeRect":
		pad := draw.thickness/2 + 2
		return draw.x - pad, draw.y - pad, draw.w + 2*pad, draw.h + 2*pad, true
	case "line":
		pad := draw.thickness/2 + 2
		x, y := min64(draw.x, draw.w), min64(draw.y, draw.h)
		return x - pad, y - pad, max64(draw.x, draw.w) - x + 2*pad, max64(draw.y, draw.h) - y + 2*pad, true
	case "fillPath", "strokePath":
		minX, minY, maxX, maxY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
		visit := func(p uiPathPoint) {
			minX, minY = math.Min(minX, p.x), math.Min(minY, p.y)
			maxX, maxY = math.Max(maxX, p.x), math.Max(maxY, p.y)
		}
		for _, figure := range draw.figures {
			visit(figure.start)
			for _, edge := range figure.edges {
				if edge.cubic {
					visit(edge.p1)
					visit(edge.p2)
				}
				visit(edge.end)
			}
		}
		if math.IsInf(minX, 0) {
			return 0, 0, 0, 0, false
		}
		pad := float64(draw.thickness)/2 + 2
		if draw.kind == "fillPath" {
			pad = 2
		}
		x, y := int64(math.Floor(minX-pad)), int64(math.Floor(minY-pad))
		return x, y, int64(math.Ceil(maxX+pad)) - x, int64(math.Ceil(maxY+pad)) - y, true
	case "hitRegion":
		return 0, 0, 0, 0, false
	}
	return 0, 0, 0, 0, false
}

func (d *uiDirtyRegions) invalidate() {
	if d.w == nil || d.w.hwnd == 0 {
		return
	}
	if !d.full {
		var area int64
		for _, r := range d.regions {
			area += int64(r.right-r.left) * int64(r.bottom-r.top)
		}
		scale := d.w.scale
		if scale <= 0 {
			scale = 1
		}
		if area > int64(float64(d.w.width*d.w.height)*scale*scale/2) {
			d.full = true
		}
	}
	if d.full {
		uiInvalidateRect.Call(d.w.hwnd, 0, 0)
		if d.w.trace.enabled {
			d.w.trace.fullRedraws++
		}
		return
	}
	for _, r := range d.regions {
		uiInvalidateRect.Call(d.w.hwnd, uintptr(unsafe.Pointer(&r)), 0)
		if d.w.trace.enabled {
			d.w.trace.invalidationRegions++
			d.w.trace.invalidatedPixels += int64(r.right-r.left) * int64(r.bottom-r.top)
		}
	}
}

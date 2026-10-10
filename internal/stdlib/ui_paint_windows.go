//go:build windows

package stdlib

import (
	"math"
	"runtime"
	"syscall"
	"unsafe"
)

var (
	uiRoundRect          = uiGdi32.NewProc("RoundRect")
	uiRectangle          = uiGdi32.NewProc("Rectangle")
	uiCreatePen          = uiGdi32.NewProc("CreatePen")
	uiMoveToEx           = uiGdi32.NewProc("MoveToEx")
	uiLineTo             = uiGdi32.NewProc("LineTo")
	uiGetStockObject     = uiGdi32.NewProc("GetStockObject")
	uiCreateCompatibleDC = uiGdi32.NewProc("CreateCompatibleDC")
	uiDeleteDC           = uiGdi32.NewProc("DeleteDC")
	uiCreateDIBSection   = uiGdi32.NewProc("CreateDIBSection")
	uiAlphaBlend         = syscall.NewLazyDLL("msimg32.dll").NewProc("AlphaBlend")
)

type uiBitmapHeader struct {
	size               uint32
	width, height      int32
	planes, bits       uint16
	compression, bytes uint32
	xPixels, yPixels   int32
	colors, important  uint32
}
type uiBitmapInfo struct {
	header uiBitmapHeader
	colors [1]uint32
}

func gdiColor(argb uint32) uintptr {
	return uintptr((argb&0xff)<<16 | (argb & 0xff00) | (argb >> 16 & 0xff))
}

func alphaImage(hdc uintptr, asset *uiImage, x, y, w, h int32) {
	if asset == nil || len(asset.pixels) == 0 || w <= 0 || h <= 0 {
		return
	}
	info := uiBitmapInfo{header: uiBitmapHeader{size: uint32(unsafe.Sizeof(uiBitmapHeader{})), width: int32(asset.width), height: -int32(asset.height), planes: 1, bits: 32, compression: 0}}
	var bits uintptr
	bitmap, _, _ := uiCreateDIBSection.Call(hdc, uintptr(unsafe.Pointer(&info)), 0, uintptr(unsafe.Pointer(&bits)), 0, 0)
	if bitmap == 0 || bits == 0 {
		return
	}
	copy(unsafe.Slice((*byte)(unsafe.Pointer(bits)), len(asset.pixels)), asset.pixels)
	mem, _, _ := uiCreateCompatibleDC.Call(hdc)
	if mem == 0 {
		uiDeleteObject.Call(bitmap)
		return
	}
	old, _, _ := uiSelectObject.Call(mem, bitmap)
	uiAlphaBlend.Call(hdc, uintptr(x), uintptr(y), uintptr(w), uintptr(h), mem, 0, 0, uintptr(asset.width), uintptr(asset.height), 0x01ff0000)
	uiSelectObject.Call(mem, old)
	uiDeleteDC.Call(mem)
	uiDeleteObject.Call(bitmap)
	runtime.KeepAlive(info)
}

func insideRounded(px, py, x, y, width, height, radius float64) bool {
	if px < x || py < y || px >= x+width || py >= y+height {
		return false
	}
	radius = math.Min(radius, math.Min(width, height)/2)
	if radius <= 0 {
		return true
	}
	cx := math.Max(x+radius, math.Min(px, x+width-radius))
	cy := math.Max(y+radius, math.Min(py, y+height-radius))
	dx, dy := px-cx, py-cy
	return dx*dx+dy*dy <= radius*radius
}

func canvasShapeContains(cmd UiDraw, px, py float64) bool {
	x, y := float64(cmd.x), float64(cmd.y)
	switch cmd.kind {
	case "fillRect":
		return insideRounded(px, py, x, y, float64(cmd.w), float64(cmd.h), float64(cmd.radius))
	case "strokeRect":
		width := float64(cmd.thickness)
		return insideRounded(px, py, x-width/2, y-width/2, float64(cmd.w)+width, float64(cmd.h)+width, float64(cmd.radius)+width/2) && !insideRounded(px, py, x+width/2, y+width/2, math.Max(0, float64(cmd.w)-width), math.Max(0, float64(cmd.h)-width), math.Max(0, float64(cmd.radius)-width/2))
	case "line":
		x2, y2 := float64(cmd.w), float64(cmd.h)
		dx, dy := x2-x, y2-y
		length := dx*dx + dy*dy
		fraction := 0.0
		if length > 0 {
			fraction = math.Max(0, math.Min(1, ((px-x)*dx+(py-y)*dy)/length))
		}
		qx, qy := x+fraction*dx, y+fraction*dy
		distanceX, distanceY := px-qx, py-qy
		radius := float64(cmd.thickness) / 2
		return distanceX*distanceX+distanceY*distanceY <= radius*radius
	}
	return false
}

// AlphaBlend requires premultiplied pixels. Tiles keep fallback allocations bounded.
func alphaCanvasShape(hdc uintptr, box *uiBox, cmd UiDraw, scale float64) {
	if scale <= 0 {
		scale = 1
	}
	minX, minY, maxX, maxY := float64(cmd.x), float64(cmd.y), float64(cmd.x+cmd.w), float64(cmd.y+cmd.h)
	if cmd.kind == "line" {
		minX = math.Min(float64(cmd.x), float64(cmd.w))
		minY = math.Min(float64(cmd.y), float64(cmd.h))
		maxX = math.Max(float64(cmd.x), float64(cmd.w))
		maxY = math.Max(float64(cmd.y), float64(cmd.h))
	}
	pad := float64(cmd.thickness)/2 + 1
	if cmd.kind != "fillRect" {
		minX -= pad
		minY -= pad
		maxX += pad
		maxY += pad
	}
	left := int(math.Max(float64(box.x)*scale, math.Floor((float64(box.x)+minX)*scale)))
	top := int(math.Max(float64(box.y)*scale, math.Floor((float64(box.y)+minY)*scale)))
	right := int(math.Min(float64(box.x+box.w)*scale, math.Ceil((float64(box.x)+maxX)*scale)))
	bottom := int(math.Min(float64(box.y+box.h)*scale, math.Ceil((float64(box.y)+maxY)*scale)))
	if left >= right || top >= bottom {
		return
	}
	a := int(cmd.color >> 24)
	r, g, b := int(cmd.color>>16&255)*a/255, int(cmd.color>>8&255)*a/255, int(cmd.color&255)*a/255
	for ty := top; ty < bottom; ty += 256 {
		for tx := left; tx < right; tx += 256 {
			width, height := min(256, right-tx), min(256, bottom-ty)
			tile := &uiImage{width: width, height: height, pixels: make([]byte, width*height*4)}
			for py := 0; py < height; py++ {
				for px := 0; px < width; px++ {
					localX := (float64(tx+px)+0.5)/scale - float64(box.x)
					localY := (float64(ty+py)+0.5)/scale - float64(box.y)
					if canvasShapeContains(cmd, localX, localY) {
						i := (py*width + px) * 4
						tile.pixels[i], tile.pixels[i+1], tile.pixels[i+2], tile.pixels[i+3] = byte(b), byte(g), byte(r), byte(a)
					}
				}
			}
			alphaImage(hdc, tile, int32(tx), int32(ty), int32(width), int32(height))
		}
	}
}

func (w *windowsState) paintCanvasGDI(hdc uintptr, box *uiBox) {
	if box == nil || box.node.kind != "Canvas" {
		return
	}
	scale := w.scale
	if scale <= 0 {
		scale = 1
	}
	saved, _, _ := uiSaveDC.Call(hdc)
	uiIntersectClip.Call(hdc, uintptr(int32(float64(box.x)*scale)), uintptr(int32(float64(box.y)*scale)), uintptr(int32(float64(box.x+box.w)*scale)), uintptr(int32(float64(box.y+box.h)*scale)))
	for _, cmd := range box.node.draws {
		if cmd.kind == "hitRegion" {
			continue
		}
		if w.trace.enabled {
			w.trace.canvasPaintOps++
		}
		if cmd.kind == "fillPath" || cmd.kind == "strokePath" {
			paintPathGDI(hdc, box, cmd, scale)
			continue
		}
		x1 := int32(float64(box.x+cmd.x) * scale)
		y1 := int32(float64(box.y+cmd.y) * scale)
		x2 := int32(float64(box.x+cmd.x+cmd.w) * scale)
		y2 := int32(float64(box.y+cmd.y+cmd.h) * scale)
		if cmd.kind == "image" {
			alphaImage(hdc, cmd.image, x1, y1, x2-x1, y2-y1)
			continue
		}
		if cmd.color>>24 == 0 {
			continue
		}
		if cmd.color>>24 < 255 && cmd.kind != "text" {
			alphaCanvasShape(hdc, box, cmd, scale)
			continue
		}
		color := gdiColor(cmd.color)
		switch cmd.kind {
		case "fillRect":
			brush, _, _ := uiCreateBrush.Call(color)
			if brush == 0 {
				continue
			}
			if cmd.radius == 0 {
				r := windowsRect{x1, y1, x2, y2}
				uiFillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), brush)
			} else {
				oldBrush, _, _ := uiSelectObject.Call(hdc, brush)
				pen, _, _ := uiGetStockObject.Call(8)
				oldPen, _, _ := uiSelectObject.Call(hdc, pen)
				radius := int32(math.Min(float64(cmd.radius*2)*scale, float64(min64(cmd.w, cmd.h))*scale))
				uiRoundRect.Call(hdc, uintptr(x1), uintptr(y1), uintptr(x2), uintptr(y2), uintptr(radius), uintptr(radius))
				uiSelectObject.Call(hdc, oldPen)
				uiSelectObject.Call(hdc, oldBrush)
			}
			uiDeleteObject.Call(brush)
		case "strokeRect":
			pen, _, _ := uiCreatePen.Call(0, uintptr(max64(1, int64(math.Round(float64(cmd.thickness)*scale)))), color)
			if pen == 0 {
				continue
			}
			oldPen, _, _ := uiSelectObject.Call(hdc, pen)
			hollow, _, _ := uiGetStockObject.Call(5)
			oldBrush, _, _ := uiSelectObject.Call(hdc, hollow)
			if cmd.radius > 0 {
				radius := int32(float64(cmd.radius*2) * scale)
				uiRoundRect.Call(hdc, uintptr(x1), uintptr(y1), uintptr(x2), uintptr(y2), uintptr(radius), uintptr(radius))
			} else {
				uiRectangle.Call(hdc, uintptr(x1), uintptr(y1), uintptr(x2), uintptr(y2))
			}
			uiSelectObject.Call(hdc, oldBrush)
			uiSelectObject.Call(hdc, oldPen)
			uiDeleteObject.Call(pen)
		case "line":
			pen, _, _ := uiCreatePen.Call(0, uintptr(max64(1, int64(math.Round(float64(cmd.thickness)*scale)))), color)
			if pen == 0 {
				continue
			}
			old, _, _ := uiSelectObject.Call(hdc, pen)
			uiMoveToEx.Call(hdc, uintptr(x1), uintptr(y1), 0)
			uiLineTo.Call(hdc, uintptr(int32(float64(box.x+cmd.w)*scale)), uintptr(int32(float64(box.y+cmd.h)*scale)))
			uiSelectObject.Call(hdc, old)
			uiDeleteObject.Call(pen)
		case "text":
			var font, oldFont uintptr
			if cmd.fontSize > 0 {
				face, _ := syscall.UTF16PtrFromString("Segoe UI")
				font, _, _ = uiCreateFont.Call(uintptr(int64(-float64(cmd.fontSize)*scale)), 0, 0, 0, uintptr(cmd.fontWeight), 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(face)))
				if font != 0 {
					oldFont, _, _ = uiSelectObject.Call(hdc, font)
				}
			}
			value, err := syscall.UTF16PtrFromString(cmd.text)
			if err != nil {
				if font != 0 {
					uiSelectObject.Call(hdc, oldFont)
					uiDeleteObject.Call(font)
				}
				continue
			}
			uiSetBkMode.Call(hdc, 1)
			uiSetTextColor.Call(hdc, color)
			r := windowsRect{x1, y1, x2, y2}
			flags := uintptr(0x0020 | 0x0004 | 0x0800)
			if cmd.fontSize > 0 {
				flags = 0x0010 | 0x0800
			}
			if cmd.textAlign == uiTextAlignCenter {
				flags |= 0x0001
			}
			if cmd.textAlign == uiTextAlignEnd {
				flags |= 0x0002
			}
			uiDrawText.Call(hdc, uintptr(unsafe.Pointer(value)), ^uintptr(0), uintptr(unsafe.Pointer(&r)), flags)
			if font != 0 {
				uiSelectObject.Call(hdc, oldFont)
				uiDeleteObject.Call(font)
			}
		}
	}
	if saved != 0 {
		uiRestoreDC.Call(hdc, ^uintptr(0))
	}
}

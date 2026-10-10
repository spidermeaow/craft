//go:build windows

package stdlib

import (
	"errors"
	"fmt"
	"math"
	"runtime"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

var errDirect2DDeviceLost = errors.New("Direct2D device lost")

// Direct2D and DirectWrite are Windows system APIs. All COM resources in this
// file are created, used, and released on the UI host's locked OS thread.
var (
	d2dDLL              = syscall.NewLazyDLL("d2d1.dll")
	dwriteDLL           = syscall.NewLazyDLL("dwrite.dll")
	d2dCreateFactory    = d2dDLL.NewProc("D2D1CreateFactory")
	dwriteCreateFactory = dwriteDLL.NewProc("DWriteCreateFactory")
)

type uiGUID struct {
	d1     uint32
	d2, d3 uint16
	d4     [8]byte
}

var d2dFactoryID = uiGUID{0x06152247, 0x6f50, 0x465a, [8]byte{0x92, 0x45, 0x11, 0x8b, 0xfd, 0x3b, 0x60, 0x07}}
var dwriteFactoryID = uiGUID{0xb859ee5a, 0xd838, 0x4b5b, [8]byte{0xa2, 0xe8, 0x1a, 0xdc, 0x7d, 0x93, 0xdb, 0x48}}

type d2dPixelFormat struct{ format, alpha uint32 }
type d2dTargetProperties struct {
	targetType      uint32
	pixel           d2dPixelFormat
	dpiX, dpiY      float32
	usage, minLevel uint32
}
type d2dHwndProperties struct {
	hwnd          uintptr
	width, height uint32
	present       uint32
}
type d2dRect struct{ left, top, right, bottom float32 }
type d2dRoundedRect struct {
	rect   d2dRect
	rx, ry float32
}
type d2dColor struct{ r, g, b, a float32 }
type d2dSize struct{ width, height uint32 }
type d2dBitmapProperties struct {
	pixel      d2dPixelFormat
	dpiX, dpiY float32
}
type dwriteTextMetrics struct {
	left, top, width, trailingWidth, height, layoutWidth, layoutHeight float32
	bidiDepth, lineCount                                               uint32
}

type direct2D struct {
	factory, writeFactory, target uintptr
	textFormat, buttonFormat      uintptr
	brushes                       map[uint32]uintptr
	images                        map[*uiImage]uintptr
	styledFormats                 map[[3]int64]uintptr
}

func comMethod(obj uintptr, index uintptr) uintptr {
	return *(*uintptr)(unsafe.Pointer(*(*uintptr)(unsafe.Pointer(obj)) + index*unsafe.Sizeof(uintptr(0))))
}

func comCall(obj uintptr, index uintptr, args ...uintptr) uintptr {
	var call [16]uintptr
	if len(args)+1 > len(call) {
		panic("Direct2D COM argument limit")
	}
	call[0] = obj
	copy(call[1:], args)
	r, _, _ := syscall.SyscallN(comMethod(obj, index), call[:len(args)+1]...)
	return r
}

func failedHR(hr uintptr) bool { return uint32(hr)&0x80000000 != 0 }

func releaseCOM(obj *uintptr) {
	if *obj != 0 {
		comCall(*obj, 2)
		*obj = 0
	}
}

const (
	dwriteTextAlignLeading  = 0
	dwriteTextAlignTrailing = 1
	dwriteTextAlignCenter   = 2
)

func dwriteTextAlignment(align int64) uintptr {
	switch align {
	case uiTextAlignCenter:
		return dwriteTextAlignCenter
	case uiTextAlignEnd:
		return dwriteTextAlignTrailing
	default:
		return dwriteTextAlignLeading
	}
}

func (d *direct2D) discardTarget() {
	for _, bitmap := range d.images {
		releaseCOM(&bitmap)
	}
	d.images = nil
	for _, brush := range d.brushes {
		releaseCOM(&brush)
	}
	d.brushes = nil
	releaseCOM(&d.target)
}

func (d *direct2D) close() {
	d.discardTarget()
	for _, format := range d.styledFormats {
		releaseCOM(&format)
	}
	d.styledFormats = nil
	releaseCOM(&d.textFormat)
	releaseCOM(&d.buttonFormat)
	releaseCOM(&d.writeFactory)
	releaseCOM(&d.factory)
}

func (d *direct2D) styledTextFormat(size, weight, align int64) uintptr {
	if size == 0 {
		return d.textFormat
	}
	if d.styledFormats == nil {
		d.styledFormats = map[[3]int64]uintptr{}
	}
	key := [3]int64{size, weight, align}
	if format := d.styledFormats[key]; format != 0 {
		return format
	}
	face, _ := syscall.UTF16PtrFromString("Segoe UI")
	locale, _ := syscall.UTF16PtrFromString("en-us")
	var format uintptr
	hr := comCall(d.writeFactory, 15, uintptr(unsafe.Pointer(face)), 0, uintptr(weight), 0, 5, uintptr(math.Float32bits(float32(size))), uintptr(unsafe.Pointer(locale)), uintptr(unsafe.Pointer(&format)))
	runtime.KeepAlive(face)
	runtime.KeepAlive(locale)
	if failedHR(hr) || format == 0 {
		return d.textFormat
	}
	comCall(format, 4, 2)
	comCall(format, 3, dwriteTextAlignment(align))
	comCall(format, 5, 0) // word wrap inside the supplied bounds
	d.styledFormats[key] = format
	return format
}

func newDirect2D() (*direct2D, error) {
	d := &direct2D{}
	hr, _, err := d2dCreateFactory.Call(0, uintptr(unsafe.Pointer(&d2dFactoryID)), 0, uintptr(unsafe.Pointer(&d.factory)))
	if failedHR(hr) || d.factory == 0 {
		return nil, fmt.Errorf("Direct2D factory: %v (HRESULT %#x)", err, uint32(hr))
	}
	hr, _, err = dwriteCreateFactory.Call(0, uintptr(unsafe.Pointer(&dwriteFactoryID)), uintptr(unsafe.Pointer(&d.writeFactory)))
	if failedHR(hr) || d.writeFactory == 0 {
		d.close()
		return nil, fmt.Errorf("DirectWrite factory: %v (HRESULT %#x)", err, uint32(hr))
	}
	face, _ := syscall.UTF16PtrFromString("Segoe UI")
	locale, _ := syscall.UTF16PtrFromString("en-us")
	createFormat := func(out *uintptr) error {
		hr := comCall(d.writeFactory, 15, uintptr(unsafe.Pointer(face)), 0, 400, 0, 5, uintptr(math.Float32bits(18)), uintptr(unsafe.Pointer(locale)), uintptr(unsafe.Pointer(out)))
		runtime.KeepAlive(face)
		runtime.KeepAlive(locale)
		if failedHR(hr) || *out == 0 {
			return fmt.Errorf("DirectWrite text format: HRESULT %#x", uint32(hr))
		}
		comCall(*out, 4, 2) // center vertically
		return nil
	}
	if err := createFormat(&d.textFormat); err != nil {
		d.close()
		return nil, err
	}
	if err := createFormat(&d.buttonFormat); err != nil {
		d.close()
		return nil, err
	}
	comCall(d.buttonFormat, 3, dwriteTextAlignment(uiTextAlignCenter))
	return d, nil
}

func (d *direct2D) ensureTarget(hwnd uintptr, width, height int32, scale float64) error {
	if d.target != 0 {
		return nil
	}
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	if scale <= 0 {
		scale = 1
	}
	props := d2dTargetProperties{targetType: 2, dpiX: float32(scale * 96), dpiY: float32(scale * 96)} // D2D1_RENDER_TARGET_TYPE_HARDWARE
	hwndProps := d2dHwndProperties{hwnd: hwnd, width: uint32(width), height: uint32(height)}
	hr := comCall(d.factory, 14, uintptr(unsafe.Pointer(&props)), uintptr(unsafe.Pointer(&hwndProps)), uintptr(unsafe.Pointer(&d.target)))
	runtime.KeepAlive(props)
	runtime.KeepAlive(hwndProps)
	if failedHR(hr) || d.target == 0 {
		return fmt.Errorf("Direct2D render target: HRESULT %#x", uint32(hr))
	}
	d.brushes = map[uint32]uintptr{}
	d.images = map[*uiImage]uintptr{}
	// The window target is opaque. ClearType preserves the weight and edge
	// definition of small Segoe UI text better than grayscale antialiasing.
	comCall(d.target, 34, 1) // SetTextAntialiasMode(D2D1_TEXT_ANTIALIAS_MODE_CLEARTYPE)
	return nil
}

func (d *direct2D) resize(width, height int32) {
	if d.target == 0 {
		return
	}
	if width < 1 || height < 1 {
		return
	}
	size := d2dSize{uint32(width), uint32(height)}
	hr := comCall(d.target, 58, uintptr(unsafe.Pointer(&size)))
	runtime.KeepAlive(size)
	if failedHR(hr) {
		d.discardTarget()
	}
}

func (d *direct2D) brush(color uint32) uintptr {
	return d.brushARGB(0xff000000 | color)
}

func (d *direct2D) brushARGB(color uint32) uintptr {
	if brush := d.brushes[color]; brush != 0 {
		return brush
	}
	c := d2dColor{float32((color>>16)&255) / 255, float32((color>>8)&255) / 255, float32(color&255) / 255, float32(color>>24) / 255}
	var brush uintptr
	hr := comCall(d.target, 8, uintptr(unsafe.Pointer(&c)), 0, uintptr(unsafe.Pointer(&brush)))
	runtime.KeepAlive(c)
	if failedHR(hr) {
		return 0
	}
	d.brushes[color] = brush
	return brush
}

func (d *direct2D) fill(rect d2dRect, color uint32, radius float32) {
	brush := d.brush(color)
	if brush == 0 {
		return
	}
	if radius > 0 {
		r := d2dRoundedRect{rect: rect, rx: radius, ry: radius}
		comCall(d.target, 19, uintptr(unsafe.Pointer(&r)), brush)
		runtime.KeepAlive(r)
	} else {
		comCall(d.target, 17, uintptr(unsafe.Pointer(&rect)), brush)
		runtime.KeepAlive(rect)
	}
}

func d2dPoint(x, y float32) uintptr {
	return uintptr(uint64(math.Float32bits(x)) | uint64(math.Float32bits(y))<<32)
}

func (d *direct2D) canvasBitmap(asset *uiImage) uintptr {
	if asset == nil || len(asset.pixels) == 0 {
		return 0
	}
	if bitmap := d.images[asset]; bitmap != 0 {
		return bitmap
	}
	props := d2dBitmapProperties{pixel: d2dPixelFormat{format: 87, alpha: 1}, dpiX: 96, dpiY: 96}
	size := uintptr(uint64(uint32(asset.width)) | uint64(uint32(asset.height))<<32)
	var bitmap uintptr
	hr := comCall(d.target, 4, size, uintptr(unsafe.Pointer(&asset.pixels[0])), uintptr(asset.width*4), uintptr(unsafe.Pointer(&props)), uintptr(unsafe.Pointer(&bitmap)))
	runtime.KeepAlive(asset)
	runtime.KeepAlive(props)
	if failedHR(hr) || bitmap == 0 {
		return 0
	}
	d.images[asset] = bitmap
	return bitmap
}

func (d *direct2D) releaseUnusedImages(root *UiNode) {
	if d == nil || len(d.images) == 0 {
		return
	}
	used := map[*uiImage]bool{}
	var walk func(*UiNode)
	walk = func(n *UiNode) {
		if n == nil {
			return
		}
		for _, cmd := range n.draws {
			if cmd.image != nil {
				used[cmd.image] = true
			}
		}
		for _, c := range n.children {
			walk(c)
		}
	}
	walk(root)
	for asset, bitmap := range d.images {
		if !used[asset] {
			releaseCOM(&bitmap)
			delete(d.images, asset)
		}
	}
}

// A stroke centered on an integer DIP can straddle physical pixels. Match its
// width to a whole physical pixel count, then place axis-aligned edges on the
// pixel center required by that width at the current DPI.
func d2dStrokePixels(thickness int64, scale float64) float64 {
	return math.Max(1, math.Round(float64(thickness)*scale))
}

func d2dStrokeCenter(coord, pixels, scale float64) float32 {
	phase := 0.0
	if int64(pixels)%2 != 0 {
		phase = 0.5
	}
	return float32((math.Round(coord*scale-phase) + phase) / scale)
}

func (d *direct2D) paintCanvas(box *uiBox, operations *int64, scale float64) bool {
	if box == nil || box.node.kind != "Canvas" {
		return true
	}
	clip := d2dRect{float32(box.x), float32(box.y), float32(box.x + box.w), float32(box.y + box.h)}
	comCall(d.target, 45, uintptr(unsafe.Pointer(&clip)), 0)
	for _, cmd := range box.node.draws {
		if cmd.kind == "hitRegion" {
			continue
		}
		if operations != nil {
			*operations = *operations + 1
		}
		if cmd.kind == "image" {
			if bitmap := d.canvasBitmap(cmd.image); bitmap != 0 {
				r := d2dRect{float32(box.x + cmd.x), float32(box.y + cmd.y), float32(box.x + cmd.x + cmd.w), float32(box.y + cmd.y + cmd.h)}
				comCall(d.target, 26, bitmap, uintptr(unsafe.Pointer(&r)), uintptr(math.Float32bits(1)), 1, 0)
				runtime.KeepAlive(r)
			} else {
				comCall(d.target, 46)
				return false
			}
			continue
		}
		if cmd.color>>24 == 0 {
			continue
		}
		brush := d.brushARGB(cmd.color)
		if brush == 0 {
			continue
		}
		r := d2dRect{float32(box.x + cmd.x), float32(box.y + cmd.y), float32(box.x + cmd.x + cmd.w), float32(box.y + cmd.y + cmd.h)}
		switch cmd.kind {
		case "fillPath", "strokePath":
			if !d.paintPath(box, cmd, scale) {
				comCall(d.target, 46)
				return false
			}
		case "fillRect":
			if cmd.radius > 0 {
				rounded := d2dRoundedRect{rect: r, rx: float32(cmd.radius), ry: float32(cmd.radius)}
				comCall(d.target, 19, uintptr(unsafe.Pointer(&rounded)), brush)
				runtime.KeepAlive(rounded)
			} else {
				comCall(d.target, 17, uintptr(unsafe.Pointer(&r)), brush)
			}
		case "strokeRect":
			pixels := d2dStrokePixels(cmd.thickness, scale)
			inset := pixels / (2 * scale)
			r.left = d2dStrokeCenter(float64(r.left)+inset, pixels, scale)
			r.top = d2dStrokeCenter(float64(r.top)+inset, pixels, scale)
			r.right = d2dStrokeCenter(float64(r.right)-inset, pixels, scale)
			r.bottom = d2dStrokeCenter(float64(r.bottom)-inset, pixels, scale)
			width := uintptr(math.Float32bits(float32(pixels / scale)))
			if cmd.radius > 0 {
				rounded := d2dRoundedRect{rect: r, rx: float32(cmd.radius), ry: float32(cmd.radius)}
				comCall(d.target, 18, uintptr(unsafe.Pointer(&rounded)), brush, width, 0)
				runtime.KeepAlive(rounded)
			} else {
				comCall(d.target, 16, uintptr(unsafe.Pointer(&r)), brush, width, 0)
			}
		case "line":
			pixels := d2dStrokePixels(cmd.thickness, scale)
			x1, y1 := float64(box.x+cmd.x), float64(box.y+cmd.y)
			x2, y2 := float64(box.x+cmd.w), float64(box.y+cmd.h)
			if x1 == x2 {
				x1 = float64(d2dStrokeCenter(x1, pixels, scale))
				x2 = x1
			} else if y1 == y2 {
				y1 = float64(d2dStrokeCenter(y1, pixels, scale))
				y2 = y1
			}
			comCall(d.target, 15, d2dPoint(float32(x1), float32(y1)), d2dPoint(float32(x2), float32(y2)), brush, uintptr(math.Float32bits(float32(pixels/scale))), 0)
		case "text":
			if len(cmd.textUTF16) > 0 {
				comCall(d.target, 27, uintptr(unsafe.Pointer(&cmd.textUTF16[0])), uintptr(len(cmd.textUTF16)), d.styledTextFormat(cmd.fontSize, cmd.fontWeight, cmd.textAlign), uintptr(unsafe.Pointer(&r)), brush, 0, 0)
				runtime.KeepAlive(cmd.textUTF16)
			}
		}
		runtime.KeepAlive(r)
	}
	comCall(d.target, 46)
	return true
}

func (d *direct2D) drawText(chars []uint16, rect d2dRect, node *UiNode) {
	if len(chars) == 0 {
		return
	}
	format := d.textFormat
	color := uint32(0x202020)
	if node.kind == "Button" {
		format = d.buttonFormat
		color = 0xffffff
	}
	if node.kind == "Text" && node.alignX == 2 && !node.customTextStyle {
		format = d.styledTextFormat(18, 400, uiTextAlignCenter)
	}
	if node.kind == "Text" && node.alignX == 3 && !node.customTextStyle {
		format = d.styledTextFormat(18, 400, uiTextAlignEnd)
	}
	if node.customTextStyle {
		alignment := uiTextAlignStart
		if node.kind == "Button" {
			alignment = uiTextAlignCenter
		}
		if node.kind == "Text" && node.alignX == 2 {
			alignment = uiTextAlignCenter
		}
		if node.kind == "Text" && node.alignX == 3 {
			alignment = uiTextAlignEnd
		}
		format = d.styledTextFormat(node.fontSize, node.fontWeight, alignment)
		color = node.textColor
	}
	if !node.customTextStyle {
		color |= 0xff000000
	}
	brush := d.brushARGB(color)
	if brush == 0 {
		return
	}
	comCall(d.target, 27, uintptr(unsafe.Pointer(&chars[0])), uintptr(len(chars)), format, uintptr(unsafe.Pointer(&rect)), brush, 0, 0)
	runtime.KeepAlive(chars)
	runtime.KeepAlive(rect)
}

// DirectWrite reports Unicode shaping metrics in logical DIPs. Text is kept
// single-line for the current Text/Button API; overflow is clipped by paint.
func (d *direct2D) measureText(chars []uint16) (int64, int64, bool) {
	if d == nil || d.writeFactory == 0 || d.textFormat == 0 || len(chars) == 0 {
		return 0, 22, d != nil && d.writeFactory != 0
	}
	var layout uintptr
	hr := comCall(d.writeFactory, 18, uintptr(unsafe.Pointer(&chars[0])), uintptr(len(chars)), d.textFormat, uintptr(math.Float32bits(100000)), uintptr(math.Float32bits(1000)), uintptr(unsafe.Pointer(&layout)))
	if failedHR(hr) || layout == 0 {
		return 0, 0, false
	}
	defer releaseCOM(&layout)
	var metrics dwriteTextMetrics
	hr = comCall(layout, 60, uintptr(unsafe.Pointer(&metrics)))
	if failedHR(hr) {
		return 0, 0, false
	}
	return int64(math.Ceil(float64(metrics.trailingWidth))), int64(math.Ceil(float64(metrics.height))), true
}

func (d *direct2D) measureTextStyled(chars []uint16, size, weight int64) (int64, int64, bool) {
	if d == nil || d.writeFactory == 0 || len(chars) == 0 {
		return 0, size + 4, d != nil && d.writeFactory != 0
	}
	format := d.styledTextFormat(size, weight, 0)
	var layout uintptr
	hr := comCall(d.writeFactory, 18, uintptr(unsafe.Pointer(&chars[0])), uintptr(len(chars)), format, uintptr(math.Float32bits(100000)), uintptr(math.Float32bits(1000)), uintptr(unsafe.Pointer(&layout)))
	if failedHR(hr) || layout == 0 {
		return 0, 0, false
	}
	defer releaseCOM(&layout)
	var metrics dwriteTextMetrics
	hr = comCall(layout, 60, uintptr(unsafe.Pointer(&metrics)))
	if failedHR(hr) {
		return 0, 0, false
	}
	return int64(math.Ceil(float64(metrics.trailingWidth))), int64(math.Ceil(float64(metrics.height))), true
}

func (d *direct2D) paint(root *uiBox, background uint32, focused, hovered, pressed string, focusVisible bool, dirty windowsRect, scale float64, painted *int64, operations *int64) error {
	if scale <= 0 {
		scale = 1
	}
	clip := d2dRect{float32(float64(dirty.left) / scale), float32(float64(dirty.top) / scale), float32(float64(dirty.right) / scale), float32(float64(dirty.bottom) / scale)}
	comCall(d.target, 48)                                    // BeginDraw
	comCall(d.target, 45, uintptr(unsafe.Pointer(&clip)), 0) // PushAxisAlignedClip, per primitive
	runtime.KeepAlive(clip)
	d.fill(clip, background, 0)
	var walk func(*uiBox, d2dRect)
	canvasFailed := false
	walk = func(box *uiBox, viewport d2dRect) {
		if canvasFailed || box == nil || box.w <= 0 || box.h <= 0 {
			return
		}
		r := d2dRect{float32(box.x), float32(box.y), float32(box.x + box.w), float32(box.y + box.h)}
		visible := d2dRect{left: maxFloat(r.left, viewport.left), top: maxFloat(r.top, viewport.top), right: minFloat(r.right, viewport.right), bottom: minFloat(r.bottom, viewport.bottom)}
		if visible.right <= visible.left || visible.bottom <= visible.top {
			return
		}
		*painted++
		switch box.node.kind {
		case "VirtualGrid":
			d.fill(r, 0xffffffff, 0)
		case "Text":
			if box.node.customSurface {
				d.fill(r, box.node.surfaceFill, float32(box.node.surfaceRadius))
			}
		case "Panel":
			if box.node.customSurface {
				d.fill(r, box.node.surfaceFill, float32(box.node.surfaceRadius))
			} else {
				d.fill(r, 0xf6f4f3, 10)
			}
		case "TextInput":
			if box.node.customSurface {
				d.fill(r, box.node.surfaceFill, float32(box.node.surfaceRadius))
			}
		case "Button":
			color := uint32(0x6366f1)
			if box.node.disabled {
				color = 0x9295a3
			}
			if box.node.id == hovered && !box.node.disabled {
				color = 0x7477fa
			}
			if box.node.id == focused && !box.node.disabled {
				color = 0x4f52dd
			}
			if box.node.customSurface {
				if box.node.disabled {
					d.fill(r, 0xff9295a3, float32(box.node.surfaceRadius))
				} else {
					fill := box.node.surfaceFill
					if box.node.id == pressed {
						fill = darkenUIFill(fill, 18)
					} else if box.node.id == hovered {
						fill = darkenUIFill(fill, 8)
					}
					d.fill(r, fill, float32(box.node.surfaceRadius))
				}
			} else {
				d.fill(r, color, 8)
			}
		}
		if box.node.kind == "VirtualGrid" {
			if x, y, width, height, ok := virtualGridRangeRect(box); ok {
				selection := d2dRect{float32(x), float32(y), float32(x + width), float32(y + height)}
				fill := uint32(0xffe8f0ff)
				if box.node.customRangeStyle {
					fill = box.node.rangeFill
				}
				comCall(d.target, 45, uintptr(unsafe.Pointer(&r)), 0)
				d.fill(selection, fill, 0)
				comCall(d.target, 46)
			}
		}
		if box.node.customSurface && box.node.surfaceBorder>>24 != 0 {
			borderColor := box.node.surfaceBorder
			if box.node.kind == "TextInput" && box.node.id == focused {
				borderColor = 0xff265ccc
			}
			brush := d.brushARGB(borderColor)
			if brush != 0 {
				if box.node.surfaceRadius > 0 {
					rounded := d2dRoundedRect{rect: r, rx: float32(box.node.surfaceRadius), ry: float32(box.node.surfaceRadius)}
					comCall(d.target, 18, uintptr(unsafe.Pointer(&rounded)), brush, uintptr(math.Float32bits(1)), 0)
					runtime.KeepAlive(rounded)
				} else {
					comCall(d.target, 16, uintptr(unsafe.Pointer(&r)), brush, uintptr(math.Float32bits(1)), 0)
				}
			}
		}
		if box.node.kind == "Text" || box.node.kind == "Button" {
			comCall(d.target, 45, uintptr(unsafe.Pointer(&r)), 0)
			textRect := r
			if box.gridCell && textRect.right-textRect.left > 16 {
				textRect.left += 8
				textRect.right -= 6
			}
			if box.node.wrapText {
				for i, line := range box.lines {
					lineRect := d2dRect{r.left, r.top + float32(int64(i)*box.lineHeight), r.right, r.top + float32(int64(i+1)*box.lineHeight)}
					d.drawText(utf16.Encode([]rune(line)), lineRect, box.node)
				}
			} else {
				d.drawText(box.node.textUTF16, textRect, box.node)
			}
			comCall(d.target, 46)
		}
		if box.node.kind == "Canvas" {
			if !d.paintCanvas(box, operations, scale) {
				canvasFailed = true
			}
		}
		if box.node.kind == "Button" && box.node.id == focused && focusVisible {
			d.focusRing(r)
		}
		if box.node.kind == "Canvas" {
			for _, draw := range box.node.draws {
				if draw.kind == "hitRegion" && draw.target == focused && focusVisible {
					d.focusRing(d2dRect{float32(box.x + draw.x), float32(box.y + draw.y), float32(box.x + draw.x + draw.w), float32(box.y + draw.y + draw.h)})
				}
			}
		}
		if len(box.children) > 0 {
			comCall(d.target, 45, uintptr(unsafe.Pointer(&r)), 0)
			for _, child := range box.children {
				walk(child, visible)
			}
			comCall(d.target, 46)
		}
		if box.node.kind == "VirtualGrid" {
			brush := d.brushARGB(0xffd8e0eb)
			if brush != 0 {
				comCall(d.target, 45, uintptr(unsafe.Pointer(&r)), 0)
				for _, cell := range box.children[:gridCellCount(box.node)] {
					if cell.x+cell.w <= box.x || cell.y+cell.h <= box.y || cell.x >= box.x+box.w || cell.y >= box.y+box.h {
						continue
					}
					border := d2dRect{float32(cell.x), float32(cell.y), float32(cell.x + cell.w), float32(cell.y + cell.h)}
					comCall(d.target, 16, uintptr(unsafe.Pointer(&border)), brush, uintptr(math.Float32bits(1)), 0)
				}
				comCall(d.target, 46)
			}
			if x, y, width, height, ok := virtualGridSelectionRect(box); ok {
				selection := d2dRect{float32(x + 1), float32(y + 1), float32(x + width - 1), float32(y + height - 1)}
				brush := d.brushARGB(0xff2563eb)
				if brush != 0 {
					comCall(d.target, 45, uintptr(unsafe.Pointer(&r)), 0)
					comCall(d.target, 16, uintptr(unsafe.Pointer(&selection)), brush, uintptr(math.Float32bits(2)), 0)
					comCall(d.target, 46)
				}
			}
			if x, y, width, height, ok := virtualGridRangeRect(box); ok && width > 2 && height > 2 {
				selection := d2dRect{float32(x + 1), float32(y + 1), float32(x + width - 1), float32(y + height - 1)}
				border := uint32(0xff2563eb)
				if box.node.customRangeStyle {
					border = box.node.rangeBorder
				}
				brush := d.brushARGB(border)
				if brush != 0 {
					comCall(d.target, 45, uintptr(unsafe.Pointer(&r)), 0)
					comCall(d.target, 16, uintptr(unsafe.Pointer(&selection)), brush, uintptr(math.Float32bits(2)), 0)
					comCall(d.target, 46)
				}
			}
		}
		if sx, sy, sw, sh, ok := scrollbarThumb(box); ok {
			d.fill(d2dRect{float32(sx), float32(sy), float32(sx + sw), float32(sy + sh)}, 0x929aa9, 3)
		}
		if sx, sy, sw, sh, ok := horizontalScrollbarThumb(box); ok {
			d.fill(d2dRect{float32(sx), float32(sy), float32(sx + sw), float32(sy + sh)}, 0x929aa9, 3)
		}
	}
	walk(root, clip)
	comCall(d.target, 46)             // PopAxisAlignedClip
	hr := comCall(d.target, 49, 0, 0) // EndDraw
	if uint32(hr) == 0x8899000c {
		d.discardTarget()
		return errDirect2DDeviceLost
	}
	if failedHR(hr) {
		return fmt.Errorf("Direct2D EndDraw: HRESULT %#x", uint32(hr))
	}
	if canvasFailed {
		return fmt.Errorf("UiRenderError: Direct2D Canvas command failed")
	}
	return nil
}

func (d *direct2D) focusRing(r d2dRect) {
	brush := d.brushARGB(0xff0b61d8)
	if brush == 0 {
		return
	}
	r.left += 2
	r.top += 2
	r.right -= 2
	r.bottom -= 2
	if r.right <= r.left || r.bottom <= r.top {
		return
	}
	comCall(d.target, 16, uintptr(unsafe.Pointer(&r)), brush, uintptr(math.Float32bits(2)), 0)
	runtime.KeepAlive(r)
}

func maxFloat(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
func minFloat(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

//go:build windows

package stdlib

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

const (
	wmClose            = 0x0010
	wmSize             = 0x0005
	wmKeyDown          = 0x0100
	wmKeyUp            = 0x0101
	wmSysKeyDown       = 0x0104
	wmSysKeyUp         = 0x0105
	wmLeftButtonDblClk = 0x0203
	wmCommand          = 0x0111
	wmLeftButtonDown   = 0x0201
	wmLeftButtonUp     = 0x0202
	wmMouseMove        = 0x0200
	wmMouseLeave       = 0x02A3
	wmMouseWheel       = 0x020A
	wmMouseHWheel      = 0x020E
	wmCaptureChanged   = 0x0215
	wmEraseBackground  = 0x0014
	wmPaint            = 0x000F
	wmDpiChanged       = 0x02E0
	pmRemove           = 0x0001
	qsAllInput         = 0x04ff
	mwmoInputAvailable = 0x0004
	wsOverlappedWindow = 0x00CF0000
	wsVisible          = 0x10000000
	wsClipChildren     = 0x02000000
)

var (
	uiUser32           = syscall.NewLazyDLL("user32.dll")
	uiKernel32         = syscall.NewLazyDLL("kernel32.dll")
	uiGdi32            = syscall.NewLazyDLL("gdi32.dll")
	uiDefWindowProc    = uiUser32.NewProc("DefWindowProcW")
	uiRegisterClass    = uiUser32.NewProc("RegisterClassExW")
	uiCreateWindow     = uiUser32.NewProc("CreateWindowExW")
	uiDestroyWindow    = uiUser32.NewProc("DestroyWindow")
	uiSetWindowText    = uiUser32.NewProc("SetWindowTextW")
	uiInvalidateRect   = uiUser32.NewProc("InvalidateRect")
	uiPeekMessage      = uiUser32.NewProc("PeekMessageW")
	uiTranslateMessage = uiUser32.NewProc("TranslateMessage")
	uiDispatchMessage  = uiUser32.NewProc("DispatchMessageW")
	uiGetClientRect    = uiUser32.NewProc("GetClientRect")
	uiScreenToClient   = uiUser32.NewProc("ScreenToClient")
	uiGetDC            = uiUser32.NewProc("GetDC")
	uiReleaseDC        = uiUser32.NewProc("ReleaseDC")
	uiGetDpiForWindow  = uiUser32.NewProc("GetDpiForWindow")
	uiGetKeyState      = uiUser32.NewProc("GetKeyState")
	uiSetCapture       = uiUser32.NewProc("SetCapture")
	uiReleaseCapture   = uiUser32.NewProc("ReleaseCapture")
	uiTrackMouseEvent  = uiUser32.NewProc("TrackMouseEvent")
	uiSetWindowPos     = uiUser32.NewProc("SetWindowPos")
	uiSetFocus         = uiUser32.NewProc("SetFocus")
	uiSetThreadDpi     = uiUser32.NewProc("SetThreadDpiAwarenessContext")
	uiFillRect         = uiUser32.NewProc("FillRect")
	uiBeginPaint       = uiUser32.NewProc("BeginPaint")
	uiEndPaint         = uiUser32.NewProc("EndPaint")
	uiCreateBrush      = uiGdi32.NewProc("CreateSolidBrush")
	uiDeleteObject     = uiGdi32.NewProc("DeleteObject")
	uiCreateFont       = uiGdi32.NewProc("CreateFontW")
	uiTextExtent       = uiGdi32.NewProc("GetTextExtentPoint32W")
	uiSelectObject     = uiGdi32.NewProc("SelectObject")
	uiSetTextColor     = uiGdi32.NewProc("SetTextColor")
	uiSetBkMode        = uiGdi32.NewProc("SetBkMode")
	uiDrawText         = uiUser32.NewProc("DrawTextW")
	uiSaveDC           = uiGdi32.NewProc("SaveDC")
	uiRestoreDC        = uiGdi32.NewProc("RestoreDC")
	uiIntersectClip    = uiGdi32.NewProc("IntersectClipRect")
	uiModuleHandle     = uiKernel32.NewProc("GetModuleHandleW")
	uiCreateEvent      = uiKernel32.NewProc("CreateEventW")
	uiSetEvent         = uiKernel32.NewProc("SetEvent")
	uiCloseHandle      = uiKernel32.NewProc("CloseHandle")
	uiWaitForMessages  = uiUser32.NewProc("MsgWaitForMultipleObjectsEx")
	uiClassOnce        sync.Once
	uiClassErr         error
	uiWndProc          = syscall.NewCallback(windowProcedure)
	uiWindows          sync.Map // hwnd -> *windowsState
)

type windowsClass struct {
	size, style                                                        uint32
	proc                                                               uintptr
	clsExtra, wndExtra                                                 int32
	instance, icon, cursor, background, menuName, className, smallIcon uintptr
}
type windowsPoint struct{ x, y int32 }
type windowsMessage struct {
	hwnd           uintptr
	message        uint32
	_              uint32
	wParam, lParam uintptr
	time           uint32
	point          windowsPoint
	private        uint32
}
type windowsRect struct{ left, top, right, bottom int32 }
type windowsTrackMouse struct {
	size, flags uint32
	hwnd        uintptr
	hoverTime   uint32
}
type windowsPaint struct {
	hdc             uintptr
	erase           int32
	paint           windowsRect
	restore, update int32
	reserved        [32]byte
}
type windowsState struct {
	ui                         *uiSession
	hwnd                       uintptr
	brush                      uintptr
	width, height              int64
	scale                      float64
	overflow                   bool
	root                       *UiNode
	layout                     *uiBox
	focused                    string
	hovered                    string
	pressed                    string
	pressedCanvas              string
	pressedTarget              string
	hoveredTarget              string
	tracking                   bool
	gpu                        *direct2D
	gpuUnavailable             bool
	background                 uint32
	trace                      uiTrace
	textSizes                  map[string][2]int64
	pendingOrigin              time.Time
	pendingScroll              time.Time
	activeScroll               string
	scrollDragID               string
	scrollDragOffset           int64
	scrollDragHorizontal       bool
	lastPointerX, lastPointerY int64
	paintVisible               int64
	lastMoveEvent              time.Time
	virtualPending             map[string]bool
	inputs                     map[string]*nativeTextInput
	inputFont                  uintptr
	inputFontScale             float64
	shortcuts                  map[shortcutKey]string
	scopedShortcuts            map[string]map[shortcutKey]string
	focusScopes                []focusScope
	uiaRoot                    *uiaProvider
	uiaProviders               map[string]*uiaProvider
	accProps                   uintptr
}

type shortcutKey struct{ key, mod int64 }

func (w *windowsState) invokeShortcut(key, lParam uintptr, inTextInput bool) bool {
	mod := keyModifiers()
	if inTextInput {
		// Native EDIT and IME own character editing, selection, and navigation.
		if key == 9 || key == 13 || key == 27 || (mod&2 != 0 && (key == 'A' || key == 'C' || key == 'V' || key == 'X' || key == 'Z' || key == 'Y')) {
			return false
		}
		if mod == 0 && (key < 0x70 || key > 0x87) {
			return false
		}
	}
	shortcut := shortcutKey{int64(key), mod}
	id := ""
	for i := len(w.focusScopes) - 1; i >= 0; i-- {
		if scoped := w.scopedShortcuts[w.focusScopes[i].id][shortcut]; scoped != "" {
			id = scoped
			break
		}
	}
	if id == "" {
		id = w.shortcuts[shortcut]
	}
	if id == "" {
		return false
	}
	if lParam&(1<<30) == 0 {
		w.emit(uiEvent{kind: "CommandInvoked", nodeID: id, key: int64(key), mod: mod})
	}
	return true
}

func countVirtualRows(b *uiBox) int64 {
	if b == nil {
		return 0
	}
	var count int64
	if b.node.kind == "VirtualList" {
		count = int64(len(b.children))
	}
	for _, child := range b.children {
		count += countVirtualRows(child)
	}
	return count
}

func (w *windowsState) requestVirtualRanges(b *uiBox) {
	if b == nil {
		return
	}
	if b.node.kind == "VirtualList" {
		first, end := virtualRange(b)
		visibleFirst := b.scrollY / b.node.rowHeight
		visibleEnd := min64(b.node.itemCount, (b.scrollY+b.h+b.node.rowHeight-1)/b.node.rowHeight)
		providedEnd := b.node.firstIndex + int64(len(b.children))
		if (b.node.firstIndex > visibleFirst || providedEnd < visibleEnd) && !w.virtualPending[b.node.id] {
			w.virtualPending[b.node.id] = true
			w.trace.rangeRequests++
			w.emit(uiEvent{kind: "VisibleRangeChanged", nodeID: b.node.id, firstIndex: first, endIndex: end})
		}
	}
	for _, child := range b.children {
		w.requestVirtualRanges(child)
	}
}

func (w *windowsState) finishPaintTrace(started time.Time) {
	if !w.trace.enabled {
		return
	}
	w.trace.frames++
	w.trace.record(&w.trace.paint, time.Since(started))
	if !w.pendingOrigin.IsZero() {
		w.trace.record(&w.trace.inputToEndDraw, time.Since(w.pendingOrigin))
		w.pendingOrigin = time.Time{}
	}
	if !w.pendingScroll.IsZero() {
		w.trace.record(&w.trace.scrollToEndDraw, time.Since(w.pendingScroll))
		w.pendingScroll = time.Time{}
	}
	w.trace.visibleNodes = w.paintVisible
}

func (w *windowsState) initRenderer() {
	if w.gpu != nil || w.gpuUnavailable || w.hwnd == 0 {
		return
	}
	if os.Getenv("CRAFT_UI_RENDERER") == "gdi" {
		w.gpuUnavailable = true
		return
	}
	d, err := newDirect2D()
	if err != nil {
		w.gpuUnavailable = true
		return
	}
	var rect windowsRect
	uiGetClientRect.Call(w.hwnd, uintptr(unsafe.Pointer(&rect)))
	if err := d.ensureTarget(w.hwnd, rect.right-rect.left, rect.bottom-rect.top, w.scale); err != nil {
		d.close()
		w.gpuUnavailable = true
		return
	}
	w.gpu = d
}

func (w *windowsState) invalidateButton(id string) {
	if w.hwnd == 0 || id == "" {
		return
	}
	b := nodeBox(w.layout, id)
	if b == nil {
		var findCanvas func(*uiBox) *uiBox
		findCanvas = func(box *uiBox) *uiBox {
			if box == nil {
				return nil
			}
			for _, draw := range box.node.draws {
				if draw.kind == "hitRegion" && draw.target == id {
					return box
				}
			}
			for _, child := range box.children {
				if result := findCanvas(child); result != nil {
					return result
				}
			}
			return nil
		}
		b = findCanvas(w.layout)
	}
	w.invalidateBox(b)
}

func (w *windowsState) invalidateBox(b *uiBox) {
	if w.hwnd == 0 || b == nil {
		return
	}
	s := w.scale
	if s <= 0 {
		s = 1
	}
	margin := int64(2)
	r := windowsRect{int32(float64(b.x-margin) * s), int32(float64(b.y-margin) * s), int32(float64(b.x+b.w+margin) * s), int32(float64(b.y+b.h+margin) * s)}
	uiInvalidateRect.Call(w.hwnd, uintptr(unsafe.Pointer(&r)), 0)
}

func (w *windowsState) relayout() {
	if w.root != nil {
		offsets := make(map[string][4]int64)
		scrollOffsets(w.layout, offsets)
		w.textSizes = make(map[string][2]int64)
		measure, closeMeasure := w.textMeasurer()
		measureStarted := time.Now()
		box := measureNode(w.root, measure)
		restoreScroll(box, offsets)
		var overflow int64
		reflowBox(box, w.width)
		w.trace.record(&w.trace.measure, time.Since(measureStarted))
		layoutStarted := time.Now()
		placeNode(box, 0, 0, w.width, w.height, &overflow)
		closeMeasure()
		w.layout = box
		if w.trace.enabled {
			w.trace.virtualRows = countVirtualRows(box)
		}
		w.requestVirtualRanges(box)
		if scrollByID(w.layout, w.activeScroll) == nil {
			w.activeScroll = ""
			if first := firstScroll(w.layout); first != nil {
				w.activeScroll = first.node.id
			}
		}
		w.trace.record(&w.trace.layout, time.Since(layoutStarted))
		w.trace.overflowNodes = overflow
		if w.focused != "" {
			w.scrollFocusIntoView(w.focused)
		}
		w.positionInputs()
	}
	if w.hwnd != 0 {
		uiInvalidateRect.Call(w.hwnd, 0, 1)
	}
}

func (w *windowsState) applyScroll(box *uiBox, dx, dy int64) bool {
	started := time.Now()
	if box == nil || !scrollBy(box, dx, dy) {
		return false
	}
	w.activeScroll = box.node.id
	w.requestVirtualRanges(box)
	w.trace.scrollEvents++
	if w.pendingScroll.IsZero() {
		w.pendingScroll = started
	}
	w.invalidateBox(box)
	w.positionInputs()
	if id := hitButton(w.layout, w.lastPointerX, w.lastPointerY); id != w.hovered {
		old := w.hovered
		w.hovered = id
		w.invalidateButton(old)
		w.invalidateButton(id)
	}
	return true
}

type uiPixelSize struct{ width, height int32 }

// Text measurement uses the same Segoe UI size as the GDI paint path. The
// cache is scoped to a layout pass, so DPI and font changes cannot reuse stale
// metrics. DirectWrite painting is clipped to the resulting node bounds.
func (w *windowsState) textMeasurer() (uiTextMeasure, func()) {
	w.initRenderer()
	if w.gpu != nil {
		return func(n *UiNode) (int64, int64) {
			key := fmt.Sprintf("%d/%d/%s", n.fontSize, n.fontWeight, n.text)
			if size, ok := w.textSizes[key]; ok {
				return size[0], size[1]
			}
			var width, height int64
			var ok bool
			if n.customTextStyle {
				width, height, ok = w.gpu.measureTextStyled(n.textUTF16, n.fontSize, n.fontWeight)
			} else {
				width, height, ok = w.gpu.measureText(n.textUTF16)
			}
			if !ok {
				width, height = int64(len(n.textUTF16))*10, 22
			}
			if height < 22 {
				height = 22
			}
			w.textSizes[key] = [2]int64{width, height}
			return width, height
		}, func() {}
	}
	hdc, _, _ := uiGetDC.Call(w.hwnd)
	if hdc == 0 {
		return func(n *UiNode) (int64, int64) { return int64(len(n.textUTF16)) * 10, 22 }, func() {}
	}
	face, _ := syscall.UTF16PtrFromString("Segoe UI")
	font, _, _ := uiCreateFont.Call(uintptr(int64(-18*w.scale)), 0, 0, 0, 400, 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(face)))
	var oldFont uintptr
	if font != 0 {
		oldFont, _, _ = uiSelectObject.Call(hdc, font)
	}
	closeMeasure := func() {
		if font != 0 {
			uiSelectObject.Call(hdc, oldFont)
			uiDeleteObject.Call(font)
		}
		uiReleaseDC.Call(w.hwnd, hdc)
	}
	return func(n *UiNode) (int64, int64) {
		key := fmt.Sprintf("%d/%d/%s", n.fontSize, n.fontWeight, n.text)
		if size, ok := w.textSizes[key]; ok {
			return size[0], size[1]
		}
		var styledFont, oldStyled uintptr
		if n.customTextStyle {
			styledFont, _, _ = uiCreateFont.Call(uintptr(int64(-float64(n.fontSize)*w.scale)), 0, 0, 0, uintptr(n.fontWeight), 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(face)))
			if styledFont != 0 {
				oldStyled, _, _ = uiSelectObject.Call(hdc, styledFont)
			}
		}
		var px uiPixelSize
		if len(n.textUTF16) > 0 {
			uiTextExtent.Call(hdc, uintptr(unsafe.Pointer(&n.textUTF16[0])), uintptr(len(n.textUTF16)), uintptr(unsafe.Pointer(&px)))
		}
		if styledFont != 0 {
			uiSelectObject.Call(hdc, oldStyled)
			uiDeleteObject.Call(styledFont)
		}
		scale := w.scale
		if scale <= 0 {
			scale = 1
		}
		result := [2]int64{int64(float64(px.width)/scale + 0.999), int64(float64(px.height)/scale + 0.999)}
		if result[1] < 22 {
			result[1] = 22
		}
		w.textSizes[key] = result
		return result[0], result[1]
	}, closeMeasure
}

func firstButton(n *UiNode) string {
	if n == nil {
		return ""
	}
	if n.kind == "Button" {
		return n.id
	}
	for _, c := range n.children {
		if id := firstButton(c); id != "" {
			return id
		}
	}
	return ""
}

func buttonIDs(n *UiNode, ids *[]string) {
	if n == nil {
		return
	}
	if n.kind == "Button" {
		*ids = append(*ids, n.id)
	}
	for _, c := range n.children {
		buttonIDs(c, ids)
	}
}

func (w *windowsState) paintBox(hdc uintptr, b *uiBox, dirty windowsRect) {
	if b == nil || b.w <= 0 || b.h <= 0 {
		return
	}
	scale := w.scale
	if scale <= 0 {
		scale = 1
	}
	r := windowsRect{int32(float64(b.x) * scale), int32(float64(b.y) * scale), int32(float64(b.x+b.w) * scale), int32(float64(b.y+b.h) * scale)}
	visible := windowsRect{left: max32(r.left, dirty.left), top: max32(r.top, dirty.top), right: min32(r.right, dirty.right), bottom: min32(r.bottom, dirty.bottom)}
	if visible.right <= visible.left || visible.bottom <= visible.top {
		return
	}
	w.paintVisible++
	if b.node.kind == "Panel" || b.node.kind == "Button" {
		if b.node.customSurface {
			fill := b.node.surfaceFill
			if b.node.kind == "Button" && b.node.disabled {
				fill = 0xffa39592
			}
			brush, _, _ := uiCreateBrush.Call(gdiColor(fill))
			if brush != 0 {
				if b.node.surfaceRadius == 0 {
					uiFillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), brush)
				} else {
					oldBrush, _, _ := uiSelectObject.Call(hdc, brush)
					noPen, _, _ := uiGetStockObject.Call(8)
					oldPen, _, _ := uiSelectObject.Call(hdc, noPen)
					radius := int32(float64(b.node.surfaceRadius*2) * scale)
					uiRoundRect.Call(hdc, uintptr(r.left), uintptr(r.top), uintptr(r.right), uintptr(r.bottom), uintptr(radius), uintptr(radius))
					uiSelectObject.Call(hdc, oldPen)
					uiSelectObject.Call(hdc, oldBrush)
				}
				uiDeleteObject.Call(brush)
			}
			if b.node.surfaceBorder>>24 != 0 {
				pen, _, _ := uiCreatePen.Call(0, uintptr(max64(1, int64(scale))), gdiColor(b.node.surfaceBorder))
				if pen != 0 {
					oldPen, _, _ := uiSelectObject.Call(hdc, pen)
					hollow, _, _ := uiGetStockObject.Call(5)
					oldBrush, _, _ := uiSelectObject.Call(hdc, hollow)
					if b.node.surfaceRadius > 0 {
						radius := int32(float64(b.node.surfaceRadius*2) * scale)
						uiRoundRect.Call(hdc, uintptr(r.left), uintptr(r.top), uintptr(r.right), uintptr(r.bottom), uintptr(radius), uintptr(radius))
					} else {
						uiRectangle.Call(hdc, uintptr(r.left), uintptr(r.top), uintptr(r.right), uintptr(r.bottom))
					}
					uiSelectObject.Call(hdc, oldBrush)
					uiSelectObject.Call(hdc, oldPen)
					uiDeleteObject.Call(pen)
				}
			}
		}
		if !b.node.customSurface {
			color := uintptr(0x00F3F4F6)
			if b.node.kind == "Button" {
				color = 0x00F16663
				if b.node.disabled {
					color = 0x00A39592
				}
				if w.hovered == b.node.id && !b.node.disabled {
					color = 0x00DD554F
				}
				if w.focused == b.node.id && !b.node.disabled {
					color = 0x00CA4742
				}
			}
			brush, _, _ := uiCreateBrush.Call(color)
			if brush != 0 {
				uiFillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), brush)
				uiDeleteObject.Call(brush)
			}
		}
	}
	if b.node.kind == "Text" || b.node.kind == "Button" {
		saved, _, _ := uiSaveDC.Call(hdc)
		uiIntersectClip.Call(hdc, uintptr(r.left), uintptr(r.top), uintptr(r.right), uintptr(r.bottom))
		var styledFont, oldStyled uintptr
		if b.node.customTextStyle {
			face, _ := syscall.UTF16PtrFromString("Segoe UI")
			styledFont, _, _ = uiCreateFont.Call(uintptr(int64(-float64(b.node.fontSize)*scale)), 0, 0, 0, uintptr(b.node.fontWeight), 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(face)))
			if styledFont != 0 {
				oldStyled, _, _ = uiSelectObject.Call(hdc, styledFont)
			}
		}
		value, err := syscall.UTF16PtrFromString(b.node.text)
		if err == nil {
			uiSetBkMode.Call(hdc, 1)
			color := uintptr(0x00202020)
			if b.node.kind == "Button" {
				color = 0x00FFFFFF
			}
			if b.node.customTextStyle {
				color = gdiColor(b.node.textColor)
			}
			uiSetTextColor.Call(hdc, color)
			flags := uintptr(0x0020 | 0x0004 | 0x0800) // single line, vertical center, no prefix
			if b.node.kind == "Button" {
				flags |= 0x0001
			} // horizontal center
			if b.node.wrapText {
				for i, line := range b.lines {
					text, textErr := syscall.UTF16PtrFromString(line)
					if textErr != nil {
						continue
					}
					top := r.top + int32(float64(int64(i)*b.lineHeight)*scale)
					lineRect := windowsRect{r.left, top, r.right, top + int32(float64(b.lineHeight)*scale)}
					uiDrawText.Call(hdc, uintptr(unsafe.Pointer(text)), ^uintptr(0), uintptr(unsafe.Pointer(&lineRect)), 0x0020|0x0800)
				}
			} else {
				uiDrawText.Call(hdc, uintptr(unsafe.Pointer(value)), ^uintptr(0), uintptr(unsafe.Pointer(&r)), flags)
			}
		}
		if styledFont != 0 {
			uiSelectObject.Call(hdc, oldStyled)
			uiDeleteObject.Call(styledFont)
		}
		if saved != 0 {
			uiRestoreDC.Call(hdc, ^uintptr(0))
		}
	}
	if b.node.kind == "Canvas" {
		w.paintCanvasGDI(hdc, b)
	}
	if b.node.kind == "Button" && w.focused == b.node.id {
		w.paintFocusRingGDI(hdc, r)
	}
	if b.node.kind == "Canvas" {
		for _, draw := range b.node.draws {
			if draw.kind == "hitRegion" && draw.target == w.focused {
				w.paintFocusRingGDI(hdc, windowsRect{int32(float64(b.x+draw.x) * scale), int32(float64(b.y+draw.y) * scale), int32(float64(b.x+draw.x+draw.w) * scale), int32(float64(b.y+draw.y+draw.h) * scale)})
			}
		}
	}
	if len(b.children) > 0 {
		saved, _, _ := uiSaveDC.Call(hdc)
		uiIntersectClip.Call(hdc, uintptr(r.left), uintptr(r.top), uintptr(r.right), uintptr(r.bottom))
		for _, child := range b.children {
			w.paintBox(hdc, child, visible)
		}
		if saved != 0 {
			uiRestoreDC.Call(hdc, ^uintptr(0))
		}
	}
	if sx, sy, sw, sh, ok := scrollbarThumb(b); ok {
		bar := windowsRect{int32(float64(sx) * scale), int32(float64(sy) * scale), int32(float64(sx+sw) * scale), int32(float64(sy+sh) * scale)}
		brush, _, _ := uiCreateBrush.Call(0x00A99A92)
		if brush != 0 {
			uiFillRect.Call(hdc, uintptr(unsafe.Pointer(&bar)), brush)
			uiDeleteObject.Call(brush)
		}
	}
	if sx, sy, sw, sh, ok := horizontalScrollbarThumb(b); ok {
		bar := windowsRect{int32(float64(sx) * scale), int32(float64(sy) * scale), int32(float64(sx+sw) * scale), int32(float64(sy+sh) * scale)}
		brush, _, _ := uiCreateBrush.Call(0x00A99A92)
		if brush != 0 {
			uiFillRect.Call(hdc, uintptr(unsafe.Pointer(&bar)), brush)
			uiDeleteObject.Call(brush)
		}
	}
}

func (w *windowsState) paintFocusRingGDI(hdc uintptr, r windowsRect) {
	thickness := max64(1, int64(w.scale*2))
	pen, _, _ := uiCreatePen.Call(0, uintptr(thickness), 0x00D8610B)
	if pen == 0 {
		return
	}
	oldPen, _, _ := uiSelectObject.Call(hdc, pen)
	hollow, _, _ := uiGetStockObject.Call(5)
	oldBrush, _, _ := uiSelectObject.Call(hdc, hollow)
	inset := int32(thickness + 1)
	if r.right-r.left > 2*inset && r.bottom-r.top > 2*inset {
		uiRectangle.Call(hdc, uintptr(r.left+inset), uintptr(r.top+inset), uintptr(r.right-inset), uintptr(r.bottom-inset))
	}
	uiSelectObject.Call(hdc, oldBrush)
	uiSelectObject.Call(hdc, oldPen)
	uiDeleteObject.Call(pen)
}

func max32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}
func min32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func (w *windowsState) emit(event uiEvent) {
	event.windowID = 1
	event.scale = w.scale
	event.at = time.Now()
	select {
	case w.ui.events <- event:
	default:
		w.overflow = true
	}
	if event.kind == "ButtonPressed" || event.kind == "CanvasTargetPressed" {
		w.raiseUIAInvoke(event.nodeID)
	}
}

func (w *windowsState) logical(v int64) int64 {
	if w.scale <= 0 {
		return v
	}
	return int64(float64(v) / w.scale)
}

func keyModifiers() int64 {
	var flags int64
	for _, key := range []struct {
		code uintptr
		flag int64
	}{{0x10, 1}, {0x11, 2}, {0x12, 4}} {
		state, _, _ := uiGetKeyState.Call(key.code)
		if state&0x8000 != 0 {
			flags |= key.flag
		}
	}
	return flags
}

func windowProcedure(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	v, ok := uiWindows.Load(hwnd)
	if !ok {
		r, _, _ := uiDefWindowProc.Call(hwnd, uintptr(msg), wParam, lParam)
		return r
	}
	w := v.(*windowsState)
	switch msg {
	case wmGetObject:
		if uint32(lParam) == uiaRootObjectID {
			p := w.uiaRootProvider()
			result, _, _ := uiaReturn.Call(hwnd, wParam, lParam, p.ptr())
			return result
		}
	case wmUiaInvoke:
		if p := uiaOf(wParam); p != nil {
			if p.actionable() && w.allowedInScope(p.id) {
				if p.draw() != nil {
					w.emit(uiEvent{kind: "CanvasTargetPressed", nodeID: p.id})
				} else {
					w.emit(uiEvent{kind: "ButtonPressed", nodeID: p.id})
				}
			}
			p.release()
		}
		return 0
	case wmUiaFocus:
		if p := uiaOf(wParam); p != nil {
			if w.focusCandidate(p.id) == nil {
				w.focusNode(p.id)
			}
			p.release()
		}
		return 0
	case wmClose:
		w.emit(uiEvent{kind: "CloseRequested"})
		return 0
	case wmSize:
		w.width, w.height = w.logical(int64(uint16(lParam))), w.logical(int64(uint16(lParam>>16)))
		if w.gpu != nil {
			w.gpu.resize(int32(uint16(lParam)), int32(uint16(lParam>>16)))
		}
		w.relayout()
		w.emit(uiEvent{kind: "Resized", width: w.width, height: w.height})
	case wmLeftButtonDown:
		x, y := w.logical(int64(int16(lParam))), w.logical(int64(int16(lParam>>16)))
		w.emit(uiEvent{kind: "PointerDown", x: x, y: y, mod: keyModifiers()})
		for _, box := range scrollChainAt(w.layout, x, y) {
			hx, hy, hw, hh, horizontal := horizontalScrollbarThumb(box)
			if horizontal && y >= hy-2 && y < hy+hh+2 && x >= box.x && x < box.x+box.w {
				w.activeScroll, w.scrollDragID, w.scrollDragHorizontal = box.node.id, box.node.id, true
				if x >= hx && x < hx+hw {
					w.scrollDragOffset = x - hx
				} else {
					w.scrollDragID = ""
					if x < hx {
						w.applyScroll(box, -box.w, 0)
					} else {
						w.applyScroll(box, box.w, 0)
					}
				}
				uiSetCapture.Call(hwnd)
				return 0
			}
			sx, sy, sw, sh, visible := scrollbarThumb(box)
			if !visible || x < sx-2 || x >= sx+sw+2 || y < box.y || y >= box.y+box.h {
				continue
			}
			w.activeScroll = box.node.id
			if y >= sy && y < sy+sh {
				w.scrollDragID, w.scrollDragOffset = box.node.id, y-sy
				w.scrollDragHorizontal = false
				uiSetCapture.Call(hwnd)
			} else if y < sy {
				w.applyScroll(box, 0, -box.h)
			} else {
				w.applyScroll(box, 0, box.h)
			}
			return 0
		}
		w.pressed = hitButton(w.layout, x, y)
		if !w.allowedInScope(w.pressed) {
			w.pressed = ""
		}
		w.pressedCanvas = ""
		w.pressedTarget = ""
		if box := hitCanvas(w.layout, x, y); box != nil {
			w.pressedCanvas = box.node.id
			w.pressedTarget = hitCanvasTarget(box, x, y)
			if !w.allowedInScope(w.pressedTarget) {
				w.pressedTarget = ""
			}
			id := w.pressedTarget
			if id == "" {
				id = box.node.id
			}
			w.emit(uiEvent{kind: "CanvasPointerDown", nodeID: id, x: x - box.x, y: y - box.y, mod: keyModifiers()})
		}
		if w.pressed == "" && w.inputs[w.focused] != nil {
			w.focusNode("")
		}
		uiSetCapture.Call(hwnd)
	case wmLeftButtonUp:
		x, y := w.logical(int64(int16(lParam))), w.logical(int64(int16(lParam>>16)))
		w.emit(uiEvent{kind: "PointerUp", x: x, y: y, mod: keyModifiers()})
		if w.scrollDragID != "" {
			w.scrollDragID = ""
			w.scrollDragHorizontal = false
			uiReleaseCapture.Call()
			return 0
		}
		pressed := w.pressed
		w.pressed = ""
		canvasPressed := w.pressedCanvas
		w.pressedCanvas = ""
		pressedTarget := w.pressedTarget
		w.pressedTarget = ""
		uiReleaseCapture.Call()
		if box := hitCanvas(w.layout, x, y); box != nil {
			id := hitCanvasTarget(box, x, y)
			if id == "" {
				id = box.node.id
			}
			w.emit(uiEvent{kind: "CanvasPointerUp", nodeID: id, x: x - box.x, y: y - box.y, mod: keyModifiers()})
		}
		w.emit(uiEvent{kind: "PointerClicked", x: x, y: y, mod: keyModifiers()})
		if pressed != "" && hitButton(w.layout, x, y) == pressed {
			w.focusNode(pressed)
			w.emit(uiEvent{kind: "ButtonPressed", nodeID: pressed, x: x, y: y, mod: keyModifiers()})
		}
		if box := hitCanvas(w.layout, x, y); box != nil && canvasPressed == box.node.id && w.allowedInScope(box.node.id) {
			w.emit(uiEvent{kind: "CanvasPressed", nodeID: box.node.id, x: x - box.x, y: y - box.y, mod: keyModifiers()})
			if target := hitCanvasTarget(box, x, y); target != "" && target == pressedTarget {
				w.focusNode(target)
				w.emit(uiEvent{kind: "CanvasTargetPressed", nodeID: target, x: x - box.x, y: y - box.y, mod: keyModifiers()})
			}
		}
	case wmLeftButtonDblClk:
		x, y := w.logical(int64(int16(lParam))), w.logical(int64(int16(lParam>>16)))
		w.emit(uiEvent{kind: "PointerDoubleClicked", x: x, y: y, mod: keyModifiers()})
	case wmCaptureChanged:
		w.pressed = ""
		w.pressedCanvas = ""
		w.pressedTarget = ""
		w.scrollDragID = ""
		w.scrollDragHorizontal = false
	case wmMouseMove:
		x, y := w.logical(int64(int16(lParam))), w.logical(int64(int16(lParam>>16)))
		w.lastPointerX, w.lastPointerY = x, y
		var target string
		var canvasBox *uiBox
		if box := hitCanvas(w.layout, x, y); box != nil {
			canvasBox = box
			target = hitCanvasTarget(box, x, y)
		}
		if target != w.hoveredTarget {
			if w.hoveredTarget != "" {
				w.emit(uiEvent{kind: "CanvasTargetLeft", nodeID: w.hoveredTarget, x: x, y: y})
			}
			if target != "" {
				w.emit(uiEvent{kind: "CanvasTargetEntered", nodeID: target, x: x, y: y})
			}
			w.hoveredTarget = target
		}
		if canvasBox != nil && time.Since(w.lastMoveEvent) >= 16*time.Millisecond {
			if target == "" {
				target = canvasBox.node.id
			}
			// Motion is coalescible; never close the session because a motion event filled the queue.
			event := uiEvent{kind: "CanvasPointerMoved", nodeID: target, x: x - canvasBox.x, y: y - canvasBox.y, mod: keyModifiers(), windowID: 1, scale: w.scale, at: time.Now()}
			select {
			case w.ui.events <- event:
				w.lastMoveEvent = event.at
			default:
			}
		}
		if w.scrollDragID != "" {
			if box := scrollByID(w.layout, w.scrollDragID); box != nil {
				if w.scrollDragHorizontal {
					_, _, thumb, _, visible := horizontalScrollbarThumb(box)
					content := box.children[0].w
					if visible && box.w-8-thumb > 0 {
						position := clamp64(x-box.x-4-w.scrollDragOffset, 0, box.w-8-thumb)
						next := position * (content - box.w) / (box.w - 8 - thumb)
						w.applyScroll(box, next-box.scrollX, 0)
					}
					return 0
				}
				_, _, _, thumb, visible := scrollbarThumb(box)
				content := scrollContentHeight(box)
				if visible && box.h-8-thumb > 0 {
					position := clamp64(y-box.y-4-w.scrollDragOffset, 0, box.h-8-thumb)
					next := position * (content - box.h) / (box.h - 8 - thumb)
					w.applyScroll(box, 0, next-box.scrollY)
				}
			}
			return 0
		}
		if box := scrollAt(w.layout, x, y); box != nil {
			w.activeScroll = box.node.id
		}
		id := hitButton(w.layout, x, y)
		if !w.tracking {
			track := windowsTrackMouse{size: uint32(unsafe.Sizeof(windowsTrackMouse{})), flags: 2, hwnd: hwnd}
			ok, _, _ := uiTrackMouseEvent.Call(uintptr(unsafe.Pointer(&track)))
			w.tracking = ok != 0
		}
		if id != w.hovered {
			old := w.hovered
			w.hovered = id
			w.invalidateButton(old)
			w.invalidateButton(id)
		}
	case wmMouseLeave:
		if w.hoveredTarget != "" {
			w.emit(uiEvent{kind: "CanvasTargetLeft", nodeID: w.hoveredTarget})
		}
		w.hoveredTarget = ""
		old := w.hovered
		w.hovered = ""
		w.tracking = false
		w.invalidateButton(old)
	case wmMouseWheel, wmMouseHWheel:
		point := windowsPoint{x: int32(int16(lParam)), y: int32(int16(lParam >> 16))}
		uiScreenToClient.Call(hwnd, uintptr(unsafe.Pointer(&point)))
		x, y := w.logical(int64(point.x)), w.logical(int64(point.y))
		w.lastPointerX, w.lastPointerY = x, y
		chain := scrollChainAt(w.layout, x, y)
		if len(chain) == 0 {
			break
		}
		box := chain[0]
		delta := int64(int16(wParam >> 16))
		if msg == wmMouseHWheel || keyModifiers()&1 != 0 {
			box.wheelX += delta * 48
			step := box.wheelX / 120
			box.wheelX %= 120
			if msg == wmMouseWheel {
				step = -step
			}
			for _, candidate := range chain {
				if w.applyScroll(candidate, step, 0) {
					w.activeScroll = candidate.node.id
					break
				}
			}
		} else {
			box.wheelY += delta * 48
			step := box.wheelY / 120
			box.wheelY %= 120
			for _, candidate := range chain {
				if w.applyScroll(candidate, 0, -step) {
					w.activeScroll = candidate.node.id
					break
				}
			}
		}
		return 0
	case wmKeyDown, wmSysKeyDown:
		w.emit(uiEvent{kind: "KeyPressed", nodeID: w.focused, key: int64(wParam), mod: keyModifiers()})
		if w.invokeShortcut(wParam, lParam, false) {
			return 0
		}
		if (wParam == 13 || wParam == 32) && hasCanvasTarget(w.root, w.focused) && w.allowedInScope(w.focused) {
			w.emit(uiEvent{kind: "CanvasTargetPressed", nodeID: w.focused, key: int64(wParam), mod: keyModifiers()})
			return 0
		}
		if box := scrollByID(w.layout, w.activeScroll); box != nil {
			switch wParam {
			case 0x21:
				w.applyScroll(box, 0, -box.h)
				return 0 // PageUp
			case 0x22:
				w.applyScroll(box, 0, box.h)
				return 0 // PageDown
			case 0x24:
				w.applyScroll(box, 0, -scrollContentHeight(box))
				return 0 // Home
			case 0x23:
				w.applyScroll(box, 0, scrollContentHeight(box))
				return 0 // End
			case 0x26:
				w.applyScroll(box, 0, -40)
				return 0 // Up
			case 0x28:
				w.applyScroll(box, 0, 40)
				return 0 // Down
			case 0x25:
				w.applyScroll(box, -40, 0)
				return 0 // Left
			case 0x27:
				w.applyScroll(box, 40, 0)
				return 0 // Right
			}
		}
		if wParam == 9 && w.root != nil {
			w.advanceFocus(keyModifiers()&1 != 0)
			return 0
		}
		if (wParam == 13 || wParam == 32) && hasButton(w.root, w.focused) && w.allowedInScope(w.focused) {
			w.emit(uiEvent{kind: "ButtonPressed", nodeID: w.focused, key: int64(wParam), mod: keyModifiers()})
		}
	case wmKeyUp, wmSysKeyUp:
		w.emit(uiEvent{kind: "KeyReleased", nodeID: w.focused, key: int64(wParam), mod: keyModifiers()})
	case wmDpiChanged:
		w.scale = float64(uint16(wParam)) / 96
		if lParam != 0 {
			suggested := (*windowsRect)(unsafe.Pointer(lParam))
			uiSetWindowPos.Call(hwnd, 0, uintptr(suggested.left), uintptr(suggested.top), uintptr(suggested.right-suggested.left), uintptr(suggested.bottom-suggested.top), 0x0014)
		}
		var client windowsRect
		uiGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&client)))
		w.width, w.height = w.logical(int64(client.right-client.left)), w.logical(int64(client.bottom-client.top))
		if w.gpu != nil {
			w.gpu.discardTarget()
		}
		w.relayout()
		w.emit(uiEvent{kind: "DpiChanged", width: w.width, height: w.height})
	case wmEraseBackground:
		return 1
	case wmCommand:
		w.inputCommand(wParam, lParam)
		return 0
	case wmPaint:
		var paint windowsPaint
		hdc, _, _ := uiBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&paint)))
		if hdc != 0 {
			started := time.Now()
			w.paintVisible = 0
			w.initRenderer()
			if w.gpu != nil {
				var rect windowsRect
				uiGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rect)))
				if err := w.gpu.ensureTarget(hwnd, rect.right-rect.left, rect.bottom-rect.top, w.scale); err == nil {
					if err := w.gpu.paint(w.layout, w.background, w.focused, w.hovered, paint.paint, w.scale, &w.paintVisible); err == nil {
						w.finishPaintTrace(started)
						uiEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&paint)))
						return 0
					} else if errors.Is(err, errDirect2DDeviceLost) {
						uiEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&paint)))
						uiInvalidateRect.Call(hwnd, 0, 1)
						return 0
					}
				}
				w.gpu.close()
				w.gpu = nil
				w.gpuUnavailable = true
			}
			uiFillRect.Call(hdc, uintptr(unsafe.Pointer(&paint.paint)), w.brush)
			face, _ := syscall.UTF16PtrFromString("Segoe UI")
			height := int64(-18 * w.scale)
			font, _, _ := uiCreateFont.Call(uintptr(height), 0, 0, 0, 400, 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(face)))
			var oldFont uintptr
			if font != 0 {
				oldFont, _, _ = uiSelectObject.Call(hdc, font)
			}
			w.paintBox(hdc, w.layout, paint.paint)
			w.finishPaintTrace(started)
			if font != 0 {
				uiSelectObject.Call(hdc, oldFont)
				uiDeleteObject.Call(font)
			}
			uiEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&paint)))
		}
		return 0
	}
	r, _, _ := uiDefWindowProc.Call(hwnd, uintptr(msg), wParam, lParam)
	return r
}

func registerUIClass() error {
	uiClassOnce.Do(func() {
		class, _ := syscall.UTF16PtrFromString("CraftUiBridge13")
		module, _, _ := uiModuleHandle.Call(0)
		wc := windowsClass{size: uint32(unsafe.Sizeof(windowsClass{})), style: 0x0008, proc: uiWndProc, instance: module, className: uintptr(unsafe.Pointer(class))}
		result, _, err := uiRegisterClass.Call(uintptr(unsafe.Pointer(&wc)))
		if result == 0 {
			uiClassErr = fmt.Errorf("UiUnavailable: register window class: %v", err)
		}
	})
	return uiClassErr
}

func startUI(ctx context.Context) (*uiSession, error) {
	if err := registerUIClass(); err != nil {
		return nil, err
	}
	wake, _, err := uiCreateEvent.Call(0, 0, 0, 0)
	if wake == 0 {
		return nil, fmt.Errorf("UiUnavailable: create wake event: %v", err)
	}
	u := &uiSession{commands: make(chan uiCommand, 64), events: make(chan uiEvent, 1024), done: make(chan struct{}), wake: wake}
	go runUIHost(ctx, u)
	return u, nil
}

func (u *uiSession) wakeHost() {
	u.wakeMu.Lock()
	if u.wake != 0 {
		uiSetEvent.Call(u.wake)
	}
	u.wakeMu.Unlock()
}

func (u *uiSession) closeWake() {
	u.wakeMu.Lock()
	if u.wake != 0 {
		uiCloseHandle.Call(u.wake)
		u.wake = 0
	}
	u.wakeMu.Unlock()
}

func runUIHost(ctx context.Context, u *uiSession) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	uiOle32 := syscall.NewLazyDLL("ole32.dll")
	coInit := uiOle32.NewProc("CoInitializeEx")
	coUninit := uiOle32.NewProc("CoUninitialize")
	if result, _, _ := coInit.Call(0, 2); result == 0 || result == 1 {
		defer coUninit.Call()
	}
	if uiSetThreadDpi.Find() == nil {
		previous, _, _ := uiSetThreadDpi.Call(^uintptr(3)) // PER_MONITOR_AWARE_V2 (-4)
		if previous != 0 {
			defer uiSetThreadDpi.Call(previous)
		}
	}
	defer close(u.done)
	defer u.closeWake()
	stopWake := context.AfterFunc(ctx, u.wakeHost)
	defer stopWake()
	state := &windowsState{ui: u, scale: 1, background: 0xffffff, trace: newUITrace(), virtualPending: make(map[string]bool), inputs: make(map[string]*nativeTextInput)}
	defer func() {
		state.trace.report()
		state.closeUIA()
		state.destroyInputs()
		state.closeAccessibilityAnnotation()
		if state.gpu != nil {
			state.gpu.close()
		}
		if state.hwnd != 0 {
			uiWindows.Delete(state.hwnd)
			uiDestroyWindow.Call(state.hwnd)
		}
		if state.brush != 0 {
			uiDeleteObject.Call(state.brush)
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		// Drain pending commands before sleeping. A bounded batch leaves room for input.
		for i := 0; i < 64; i++ {
			select {
			case cmd, ok := <-u.commands:
				if !ok {
					return
				}
				state.trace.record(&state.trace.queue, time.Since(cmd.queuedAt))
				cmd.answer <- executeUICommand(state, cmd)
			default:
				i = 64
			}
		}
		var msg windowsMessage
		for {
			found, _, _ := uiPeekMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0, pmRemove)
			if found == 0 {
				break
			}
			uiTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
			uiDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
		}
		if state.overflow {
			// No event may be silently lost. Stop the host and surface a failure.
			u.fail(fmt.Errorf("UiQueueOverflow: 1024 pending events exceeded; UI session closed"))
			return
		}
		// The auto-reset event wakes for Go commands; Win32 messages wake directly.
		// Check both queues before waiting so a command arriving during dispatch
		// cannot be stranded until another OS event arrives.
		select {
		case <-ctx.Done():
			return
		case cmd, ok := <-u.commands:
			if !ok {
				return
			}
			state.trace.record(&state.trace.queue, time.Since(cmd.queuedAt))
			cmd.answer <- executeUICommand(state, cmd)
			continue
		default:
		}
		result, _, waitErr := uiWaitForMessages.Call(1, uintptr(unsafe.Pointer(&u.wake)), ^uintptr(0), qsAllInput, mwmoInputAvailable)
		if uint32(result) == 0xffffffff {
			u.fail(fmt.Errorf("UiHostError: wait for input: %v", waitErr))
			return
		}
	}
}

func executeUICommand(state *windowsState, cmd uiCommand) uiReply {
	if cmd.op == "std.ui.mount" || cmd.op == "std.ui.update" {
		defer state.rebuildUIA()
	}
	switch cmd.op {
	case "open":
		if state.hwnd != 0 {
			return uiReply{err: fmt.Errorf("UiSessionError: window already open")}
		}
		title, err := syscall.UTF16PtrFromString(cmd.title)
		if err != nil {
			return uiReply{err: fmt.Errorf("UiArgumentError: title contains NUL")}
		}
		class, _ := syscall.UTF16PtrFromString("CraftUiBridge13")
		brush, _, e := uiCreateBrush.Call(0x00FFFFFF)
		if brush == 0 {
			return uiReply{err: fmt.Errorf("UiUnavailable: create brush: %v", e)}
		}
		state.brush = brush
		module, _, _ := uiModuleHandle.Call(0)
		hwnd, _, e := uiCreateWindow.Call(0, uintptr(unsafe.Pointer(class)), uintptr(unsafe.Pointer(title)), wsOverlappedWindow|wsVisible|wsClipChildren, 100, 100, uintptr(cmd.w), uintptr(cmd.h), 0, 0, module, 0)
		if hwnd == 0 {
			return uiReply{err: fmt.Errorf("UiUnavailable: create window: %v", e)}
		}
		state.hwnd = hwnd
		uiWindows.Store(hwnd, state)
		uiInvalidateRect.Call(hwnd, 0, 1)
		if err := uiGetDpiForWindow.Find(); err == nil {
			dpi, _, _ := uiGetDpiForWindow.Call(hwnd)
			if dpi >= 96 {
				state.scale = float64(dpi) / 96
			}
		}
		var rect windowsRect
		uiGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rect)))
		state.width, state.height = state.logical(int64(rect.right-rect.left)), state.logical(int64(rect.bottom-rect.top))
		state.initRenderer()
		return uiReply{id: 1}
	case "std.ui.closeWindow":
		if state.hwnd == 0 {
			return uiReply{err: fmt.Errorf("UiInvalidHandle: window already closed")}
		}
		if state.gpu != nil {
			state.gpu.close()
			state.gpu = nil
		}
		state.closeUIA()
		state.destroyInputs()
		state.closeAccessibilityAnnotation()
		uiWindows.Delete(state.hwnd)
		uiDestroyWindow.Call(state.hwnd)
		state.hwnd = 0
		state.root, state.layout = nil, nil
		state.focused, state.hovered, state.pressed = "", "", ""
		state.shortcuts = nil
		state.scopedShortcuts = nil
		state.focusScopes = nil
		state.tracking = false
		return uiReply{}
	case "std.ui.setTitle":
		if state.hwnd == 0 {
			return uiReply{err: fmt.Errorf("UiInvalidHandle: window already closed")}
		}
		title, err := syscall.UTF16PtrFromString(cmd.title)
		if err != nil {
			return uiReply{err: fmt.Errorf("UiArgumentError: title contains NUL")}
		}
		ok, _, e := uiSetWindowText.Call(state.hwnd, uintptr(unsafe.Pointer(title)))
		if ok == 0 {
			return uiReply{err: fmt.Errorf("UiHostError: set title: %v", e)}
		}
		return uiReply{}
	case "std.ui.setBackground":
		if state.hwnd == 0 {
			return uiReply{err: fmt.Errorf("UiInvalidHandle: window already closed")}
		}
		color := uintptr(cmd.r | cmd.g<<8 | cmd.b<<16)
		brush, _, e := uiCreateBrush.Call(color)
		if brush == 0 {
			return uiReply{err: fmt.Errorf("UiHostError: create brush: %v", e)}
		}
		old := state.brush
		state.brush = brush
		state.background = uint32(cmd.r<<16 | cmd.g<<8 | cmd.b)
		uiInvalidateRect.Call(state.hwnd, 0, 1)
		if old != 0 {
			uiDeleteObject.Call(old)
		}
		return uiReply{}
	case "std.ui.renderer":
		if state.hwnd == 0 {
			return uiReply{err: fmt.Errorf("UiInvalidHandle: window already closed")}
		}
		if state.gpu != nil {
			return uiReply{text: "direct2d-hardware"}
		}
		return uiReply{text: "gdi"}
	case "std.ui.viewportWidth":
		if state.hwnd == 0 {
			return uiReply{err: fmt.Errorf("UiInvalidHandle: window already closed")}
		}
		return uiReply{value: state.width}
	case "std.ui.viewportHeight":
		if state.hwnd == 0 {
			return uiReply{err: fmt.Errorf("UiInvalidHandle: window already closed")}
		}
		return uiReply{value: state.height}
	case "std.ui.measureTextWidth":
		if state.hwnd == 0 {
			return uiReply{err: fmt.Errorf("UiInvalidHandle: window already closed")}
		}
		if state.textSizes == nil {
			state.textSizes = make(map[string][2]int64)
		}
		node, err := makeNode("Text", "measurement", cmd.value, nil)
		if err != nil {
			return uiReply{err: err}
		}
		node.fontSize, node.fontWeight, node.customTextStyle = cmd.w, cmd.h, true
		measure, closeMeasure := state.textMeasurer()
		width, _ := measure(node)
		closeMeasure()
		return uiReply{value: width}
	case "std.ui.setText":
		return state.setInputText(cmd.nodeID, cmd.value)
	case "std.ui.focus":
		if err := state.focusCandidate(cmd.nodeID); err != nil {
			return uiReply{err: err}
		}
		state.focusNode(cmd.nodeID)
		return uiReply{}
	case "std.ui.pushFocusScope":
		return uiReply{err: state.pushFocusScope(cmd.scopeID)}
	case "std.ui.popFocusScope":
		return uiReply{err: state.popFocusScope(cmd.scopeID)}
	case "std.ui.bindScopedShortcut":
		return uiReply{err: state.bindScopedShortcut(cmd.scopeID, cmd.nodeID, shortcutKey{cmd.key, cmd.mod})}
	case "std.ui.unbindScopedShortcut":
		state.unbindScopedShortcut(cmd.scopeID, cmd.nodeID)
		return uiReply{}
	case "std.ui.bindShortcut":
		if state.hwnd == 0 {
			return uiReply{err: fmt.Errorf("UiInvalidHandle: window already closed")}
		}
		if state.shortcuts == nil {
			state.shortcuts = make(map[shortcutKey]string)
		}
		key := shortcutKey{cmd.key, cmd.mod}
		if previous := state.shortcuts[key]; previous != "" && previous != cmd.nodeID {
			return uiReply{err: fmt.Errorf("UiShortcutError: key %d/mod %d already bound to %q", cmd.key, cmd.mod, previous)}
		}
		if len(state.shortcuts) >= 128 && state.shortcuts[key] == "" {
			return uiReply{err: fmt.Errorf("UiShortcutError: limit of 128 shortcuts exceeded")}
		}
		state.shortcuts[key] = cmd.nodeID
		return uiReply{}
	case "std.ui.unbindShortcut":
		for key, id := range state.shortcuts {
			if id == cmd.nodeID {
				delete(state.shortcuts, key)
			}
		}
		return uiReply{}
	case "std.ui.mount", "std.ui.update":
		if state.hwnd == 0 {
			return uiReply{err: fmt.Errorf("UiInvalidHandle: window already closed")}
		}
		if cmd.op == "std.ui.mount" && state.root != nil {
			return uiReply{err: fmt.Errorf("UiTreeError: tree already mounted")}
		}
		if cmd.op == "std.ui.update" && state.root == nil {
			return uiReply{err: fmt.Errorf("UiTreeError: mount tree before update")}
		}
		if sameTree(state.root, cmd.root) {
			return uiReply{}
		}
		state.pendingOrigin = cmd.origin
		oldFocus := state.focused
		oldHover := state.hovered
		state.root = cmd.root
		state.reconcileFocusScopes()
		if state.trace.enabled {
			state.trace.drawCommands, state.trace.imageBytes = canvasMetrics(cmd.root)
		}
		if state.gpu != nil {
			state.gpu.releaseUnusedImages(cmd.root)
		}
		if err := state.reconcileInputs(cmd.root); err != nil {
			return uiReply{err: err}
		}
		state.syncInputScope()
		clear(state.virtualPending)
		state.focused = oldFocus
		if !hasFocusable(cmd.root, oldFocus) || !state.allowedInScope(oldFocus) {
			state.focusNode(firstFocusable(state.activeFocusRoot()))
		}
		state.hovered = ""
		if state.layout != nil {
			var changed []*uiBox
			if rebindLayout(state.layout, cmd.root, &changed) {
				measure, closeMeasure := state.textMeasurer()
				geometryChanged := false
				for _, box := range changed {
					if box.node.kind != "Text" && box.node.kind != "Button" && box.node.kind != "TextInput" {
						continue
					}
					mw, mh := measure(box.node)
					if box.node.kind == "Button" {
						mw += 20
						mh += 12
						if mh < 32 {
							mh = 32
						}
					} else {
						mw += 4
						mh += 4
						if mh < 24 {
							mh = 24
						}
					}
					if mw != box.prefW || mh != box.prefH {
						geometryChanged = true
						break
					}
				}
				closeMeasure()
				if geometryChanged {
					state.relayout()
					return uiReply{}
				}
				var dirtyArea int64
				for _, box := range changed {
					dirtyArea += box.w * box.h
				}
				if len(changed) > 16 || (state.width > 0 && state.height > 0 && dirtyArea > state.width*state.height/2) {
					uiInvalidateRect.Call(state.hwnd, 0, 0)
				} else {
					for _, box := range changed {
						state.invalidateBox(box)
					}
				}
				if oldFocus != state.focused {
					state.invalidateButton(oldFocus)
					state.invalidateButton(state.focused)
				}
				state.invalidateButton(oldHover)
				state.positionInputs()
				return uiReply{}
			}
		}
		state.relayout()
		return uiReply{}
	}
	return uiReply{err: fmt.Errorf("UiCapabilityError: unknown command")}
}

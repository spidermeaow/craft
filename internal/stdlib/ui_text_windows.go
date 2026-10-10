//go:build windows

package stdlib

import (
	"fmt"
	"math"
	"sync"
	"syscall"
	"unsafe"
)

const (
	wmSetFocus            = 0x0007
	wmGetText             = 0x000D
	wmGetTextLength       = 0x000E
	wmChar                = 0x0102
	wmImeStartComposition = 0x010D
	wmImeEndComposition   = 0x010E
	wmNcDestroy           = 0x0082
	wmSetFont             = 0x0030
	wmCtlColorEdit        = 0x0133
	wmCtlColorStatic      = 0x0138
	emLimitText           = 0x00C5
	emSetCueBanner        = 0x1501
	enChange              = 0x0300
	wsChild               = 0x40000000
	wsTabStop             = 0x00010000
	wsBorder              = 0x00800000
	esAutoHScroll         = 0x0080
	swHide                = 0
	swShow                = 5
	gwlpWndProc           = ^uintptr(3) // -4
	maxInputUnits         = 1024
	swpNoZOrder           = 0x0004
	swpNoActivate         = 0x0010
	swpNoCopyBits         = 0x0100
)

var (
	uiTextSetWindowLong = uiUser32.NewProc("SetWindowLongPtrW")
	uiTextCallWindow    = uiUser32.NewProc("CallWindowProcW")
	uiTextSendMessage   = uiUser32.NewProc("SendMessageW")
	uiTextShowWindow    = uiUser32.NewProc("ShowWindow")
	uiTextEnableWindow  = uiUser32.NewProc("EnableWindow")
	uiTextSetBkColor    = uiGdi32.NewProc("SetBkColor")
	uiTextSetRegion     = uiUser32.NewProc("SetWindowRgn")
	uiTextCreateRegion  = uiGdi32.NewProc("CreateRectRgn")
	uiTextWndProc       = syscall.NewCallback(textInputProcedure)
	uiTextWindows       sync.Map // child HWND -> *nativeTextInput
	uiCoCreateInstance  = syscall.NewLazyDLL("ole32.dll").NewProc("CoCreateInstance")
	uiAccCLSID          = syscall.GUID{Data1: 0xb5f8350b, Data2: 0x0548, Data3: 0x48b1, Data4: [8]byte{0xa6, 0xee, 0x88, 0xbd, 0x00, 0xb4, 0xa5, 0xe7}}
	uiAccIID            = syscall.GUID{Data1: 0x6e26e776, Data2: 0x04f0, Data3: 0x495d, Data4: [8]byte{0x80, 0xe4, 0x33, 0x30, 0x35, 0x2e, 0x31, 0x69}}
	uiAccNameGUID       = syscall.GUID{Data1: 0xc3a6921b, Data2: 0x4a99, Data3: 0x44f1, Data4: [8]byte{0xbc, 0xa6, 0x61, 0x18, 0x70, 0x52, 0xc4, 0x31}}
)

type nativeTextInput struct {
	owner               *windowsState
	hwnd                uintptr
	oldProc             uintptr
	id                  string
	placeholder         string
	accessibleName      string
	lastText            string
	composing           bool
	updating            bool
	disabled            bool
	fill                uint32
	fillBrush           uintptr
	shown               bool
	positioned          bool
	x, y, width, height int32
	clip                windowsRect
}

func (input *nativeTextInput) setFill(node *UiNode) {
	fill := uint32(0xffffffff)
	if node.customSurface {
		fill = node.surfaceFill
	}
	if input.fillBrush != 0 && input.fill == fill {
		return
	}
	if input.fillBrush != 0 {
		uiDeleteObject.Call(input.fillBrush)
	}
	input.fill = fill
	input.fillBrush, _, _ = uiCreateBrush.Call(gdiColor(fill))
	if input.hwnd != 0 {
		uiInvalidateRect.Call(input.hwnd, 0, 1)
	}
}

func collectInputs(n *UiNode, result map[string]*UiNode) {
	if n == nil {
		return
	}
	if n.kind == "TextInput" {
		result[n.id] = n
	}
	for _, child := range n.children {
		collectInputs(child, result)
	}
}

func (w *windowsState) reconcileInputs(root *UiNode) error {
	if w.inputs == nil {
		w.inputs = make(map[string]*nativeTextInput)
	}
	wanted := make(map[string]*UiNode)
	collectInputs(root, wanted)
	for id, input := range w.inputs {
		if wanted[id] == nil {
			input.updating = true
			w.clearInputAccessibleName(input.hwnd)
			uiDestroyWindow.Call(input.hwnd)
			if input.fillBrush != 0 {
				uiDeleteObject.Call(input.fillBrush)
			}
			uiTextWindows.Delete(input.hwnd)
			delete(w.inputs, id)
		}
	}
	for id, node := range wanted {
		if input := w.inputs[id]; input != nil {
			input.setFill(node)
			if input.accessibleName != node.accessibleName {
				w.clearInputAccessibleName(input.hwnd)
				w.setInputAccessibleName(input.hwnd, node.accessibleName)
				input.accessibleName = node.accessibleName
			}
			if input.disabled != node.disabled {
				if node.disabled {
					uiTextEnableWindow.Call(input.hwnd, 0)
				} else {
					uiTextEnableWindow.Call(input.hwnd, 1)
				}
				input.disabled = node.disabled
			}
			if input.placeholder != node.placeholder {
				input.placeholder = node.placeholder
				input.setPlaceholder(node.placeholder)
			}
			continue
		}
		if err := validateInputValue(node.text); err != nil {
			return err
		}
		initial, _ := syscall.UTF16PtrFromString(node.text)
		class, _ := syscall.UTF16PtrFromString("EDIT")
		module, _, _ := uiModuleHandle.Call(0)
		// Keep the child hidden until positionInputs has assigned its final box and clip.
		style := uintptr(wsChild | wsTabStop | esAutoHScroll)
		if !node.customSurface {
			style |= wsBorder
		}
		hwnd, _, err := uiCreateWindow.Call(0, uintptr(unsafe.Pointer(class)), uintptr(unsafe.Pointer(initial)), style, 0, 0, 1, 1, w.hwnd, 0, module, 0)
		if hwnd == 0 {
			return fmt.Errorf("UiUnavailable: create text input: %v", err)
		}
		input := &nativeTextInput{owner: w, hwnd: hwnd, id: id, placeholder: node.placeholder, accessibleName: node.accessibleName, lastText: node.text, disabled: node.disabled}
		input.setFill(node)
		old, _, setErr := uiTextSetWindowLong.Call(hwnd, gwlpWndProc, uiTextWndProc)
		if old == 0 {
			if input.fillBrush != 0 {
				uiDeleteObject.Call(input.fillBrush)
			}
			uiDestroyWindow.Call(hwnd)
			return fmt.Errorf("UiUnavailable: subclass text input: %v", setErr)
		}
		input.oldProc = old
		uiTextWindows.Store(hwnd, input)
		w.inputs[id] = input
		w.setInputAccessibleName(hwnd, node.accessibleName)
		if node.disabled {
			uiTextEnableWindow.Call(hwnd, 0)
		}
		uiTextSendMessage.Call(hwnd, emLimitText, maxInputUnits, 0)
		if w.inputFont != 0 {
			uiTextSendMessage.Call(hwnd, wmSetFont, w.inputFont, 1)
		}
		input.setPlaceholder(node.placeholder)
	}
	w.updateInputFont()
	return nil
}

func (w *windowsState) destroyInputs() {
	for id, input := range w.inputs {
		input.updating = true
		w.clearInputAccessibleName(input.hwnd)
		uiDestroyWindow.Call(input.hwnd)
		if input.fillBrush != 0 {
			uiDeleteObject.Call(input.fillBrush)
		}
		uiTextWindows.Delete(input.hwnd)
		delete(w.inputs, id)
	}
	if w.inputFont != 0 {
		uiDeleteObject.Call(w.inputFont)
		w.inputFont = 0
	}
}

func (w *windowsState) accessibilityAnnotation() uintptr {
	if w.accProps != 0 {
		return w.accProps
	}
	var result uintptr
	hr, _, _ := uiCoCreateInstance.Call(uintptr(unsafe.Pointer(&uiAccCLSID)), 0, 1, uintptr(unsafe.Pointer(&uiAccIID)), uintptr(unsafe.Pointer(&result)))
	if int32(hr) < 0 {
		return 0
	}
	w.accProps = result
	return result
}

func (w *windowsState) setInputAccessibleName(hwnd uintptr, name string) {
	if hwnd == 0 || name == "" {
		return
	}
	service := w.accessibilityAnnotation()
	if service == 0 {
		return
	}
	wide, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return
	}
	vtbl := *(*uintptr)(unsafe.Pointer(service))
	method := *(*uintptr)(unsafe.Pointer(vtbl + 7*unsafe.Sizeof(uintptr(0))))
	syscall.SyscallN(method, service, hwnd, uintptr(uint32(0xfffffffc)), 0, uintptr(unsafe.Pointer(&uiAccNameGUID)), uintptr(unsafe.Pointer(wide)))
}

func (w *windowsState) clearInputAccessibleName(hwnd uintptr) {
	if w.accProps == 0 || hwnd == 0 {
		return
	}
	vtbl := *(*uintptr)(unsafe.Pointer(w.accProps))
	method := *(*uintptr)(unsafe.Pointer(vtbl + 9*unsafe.Sizeof(uintptr(0))))
	syscall.SyscallN(method, w.accProps, hwnd, uintptr(uint32(0xfffffffc)), 0, uintptr(unsafe.Pointer(&uiAccNameGUID)), 1)
}

func (w *windowsState) closeAccessibilityAnnotation() {
	if w.accProps == 0 {
		return
	}
	vtbl := *(*uintptr)(unsafe.Pointer(w.accProps))
	method := *(*uintptr)(unsafe.Pointer(vtbl + 2*unsafe.Sizeof(uintptr(0))))
	syscall.SyscallN(method, w.accProps)
	w.accProps = 0
}

func (w *windowsState) updateInputFont() {
	if len(w.inputs) == 0 || (w.inputFont != 0 && w.inputFontScale == w.scale) {
		return
	}
	face, _ := syscall.UTF16PtrFromString("Segoe UI")
	height := int64(math.Round(-18 * w.scale))
	font, _, _ := uiCreateFont.Call(uintptr(height), 0, 0, 0, 400, 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(face)))
	if font == 0 {
		return
	}
	for _, input := range w.inputs {
		uiTextSendMessage.Call(input.hwnd, wmSetFont, font, 1)
	}
	if w.inputFont != 0 {
		uiDeleteObject.Call(w.inputFont)
	}
	w.inputFont = font
	w.inputFontScale = w.scale
}

func (input *nativeTextInput) setPlaceholder(value string) {
	placeholder, _ := syscall.UTF16PtrFromString(value)
	uiTextSendMessage.Call(input.hwnd, emSetCueBanner, 1, uintptr(unsafe.Pointer(placeholder)))
}

func (input *nativeTextInput) text() string {
	// Read the live EDIT contents for change and submit events.
	length, _, _ := uiTextSendMessage.Call(input.hwnd, wmGetTextLength, 0, 0)
	if length > maxInputUnits {
		length = maxInputUnits
	}
	buffer := make([]uint16, int(length)+1)
	uiTextSendMessage.Call(input.hwnd, wmGetText, uintptr(len(buffer)), uintptr(unsafe.Pointer(&buffer[0])))
	return syscall.UTF16ToString(buffer)
}

func (w *windowsState) inputCommand(wParam, lParam uintptr) {
	if uint16(wParam>>16) != enChange || lParam == 0 {
		return
	}
	value, ok := uiTextWindows.Load(lParam)
	if !ok {
		return
	}
	input := value.(*nativeTextInput)
	if input.updating || input.composing {
		return
	}
	input.changed()
}

func (input *nativeTextInput) changed() {
	text := input.text()
	if text == input.lastText {
		return
	}
	input.lastText = text
	input.owner.emit(uiEvent{kind: "TextChanged", nodeID: input.id, text: text})
	input.owner.invalidateButton(input.id)
}

func (w *windowsState) setInputText(id, value string) uiReply {
	input := w.inputs[id]
	if input == nil {
		return uiReply{err: fmt.Errorf("UiInvalidHandle: text input %q is not mounted", id)}
	}
	if input.composing {
		return uiReply{err: fmt.Errorf("UiStateError: cannot set text during IME composition")}
	}
	if err := validateInputValue(value); err != nil {
		return uiReply{err: err}
	}
	if value == input.text() {
		return uiReply{}
	}
	input.updating = true
	defer func() { input.updating = false }()
	text, _ := syscall.UTF16PtrFromString(value)
	if ok, _, err := uiSetWindowText.Call(input.hwnd, uintptr(unsafe.Pointer(text))); ok == 0 {
		return uiReply{err: fmt.Errorf("UiHostError: set input text: %v", err)}
	}
	input.lastText = value
	w.invalidateButton(id)
	return uiReply{}
}

func textInputProcedure(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	value, ok := uiTextWindows.Load(hwnd)
	if !ok {
		r, _, _ := uiDefWindowProc.Call(hwnd, uintptr(msg), wParam, lParam)
		return r
	}
	input := value.(*nativeTextInput)
	w := input.owner
	switch msg {
	case wmSetFocus:
		w.focusNode(input.id)
	case wmImeStartComposition:
		input.composing = true
	case wmImeEndComposition:
		result, _, _ := uiTextCallWindow.Call(input.oldProc, hwnd, uintptr(msg), wParam, lParam)
		input.composing = false
		if !input.updating {
			input.changed()
		}
		return result
	case wmLeftButtonDown:
		if grid := gridEditorBoxForInput(w.layout, input.id); grid != nil {
			w.gridDragID = grid.node.id
			w.gridDragAnchorRow, w.gridDragRow = grid.node.editorRow, grid.node.editorRow
			w.gridDragAnchorColumn, w.gridDragColumn = grid.node.editorColumn, grid.node.editorColumn
		}
	case wmMouseMove:
		if w.gridDragID != "" && wParam&1 != 0 {
			x := w.logical(int64(input.x) + int64(int16(lParam)))
			y := w.logical(int64(input.y) + int64(int16(lParam>>16)))
			if w.dragGridTo(x, y) {
				uiSetCapture.Call(w.hwnd)
				return 0
			}
		}
	case wmLeftButtonUp:
		if w.gridDragID != "" {
			w.gridDragID = ""
		}
	case wmKeyDown, wmSysKeyDown:
		w.emit(uiEvent{kind: "KeyPressed", nodeID: input.id, key: int64(wParam), mod: keyModifiers()})
		if !input.composing && keyModifiers()&1 != 0 && (wParam == 0x25 || wParam == 0x26 || wParam == 0x27 || wParam == 0x28) {
			if grid := gridEditorBoxForInput(w.layout, input.id); grid != nil {
				w.focusNode(grid.node.id)
				if w.gridNavigate(grid, wParam) {
					return 0
				}
			}
		}
		if !input.composing && w.invokeShortcut(wParam, lParam, true) {
			return 0
		}
		if wParam == 9 {
			w.advanceFocus(keyModifiers()&1 != 0)
			return 0
		}
		if wParam == 13 {
			w.emit(uiEvent{kind: "TextSubmitted", nodeID: input.id, text: input.text()})
			return 0
		}
	case wmChar:
		if wParam == 9 || wParam == 13 {
			return 0
		}
	case wmNcDestroy:
		r, _, _ := uiTextCallWindow.Call(input.oldProc, hwnd, uintptr(msg), wParam, lParam)
		uiTextWindows.Delete(hwnd)
		return r
	}
	r, _, _ := uiTextCallWindow.Call(input.oldProc, hwnd, uintptr(msg), wParam, lParam)
	return r
}

func gridEditorBoxForInput(box *uiBox, inputID string) *uiBox {
	if box == nil {
		return nil
	}
	if box.node.kind == "VirtualGrid" && box.node.hasGridEditor && len(box.children) > 0 && box.children[len(box.children)-1].node.id == inputID {
		return box
	}
	for _, child := range box.children {
		if found := gridEditorBoxForInput(child, inputID); found != nil {
			return found
		}
	}
	return nil
}

func nodeBox(box *uiBox, id string) *uiBox {
	if box == nil || id == "" {
		return nil
	}
	if box.node.id == id {
		return box
	}
	for _, child := range box.children {
		if result := nodeBox(child, id); result != nil {
			return result
		}
	}
	return nil
}

type inputClip struct{ left, top, right, bottom int64 }

func intersectInputClip(a, b inputClip) inputClip {
	return inputClip{max64(a.left, b.left), max64(a.top, b.top), min64(a.right, b.right), min64(a.bottom, b.bottom)}
}

func visibleInputBox(box *uiBox, id string, clip inputClip) (*uiBox, inputClip) {
	if box == nil {
		return nil, clip
	}
	right := box.x + box.w
	if _, _, _, _, visible := scrollbarThumb(box); visible {
		right -= 12
	}
	clip = intersectInputClip(clip, inputClip{box.x, box.y, right, box.y + box.h})
	if box.node.id == id {
		return box, intersectInputClip(clip, inputClip{box.x, box.y, box.x + box.w, box.y + box.h})
	}
	for _, child := range box.children {
		if found, visible := visibleInputBox(child, id, clip); found != nil {
			return found, visible
		}
	}
	return nil, clip
}

func hasFocusable(n *UiNode, id string) bool {
	if n == nil || id == "" {
		return false
	}
	if n.id == id && (n.kind == "Button" || n.kind == "TextInput" || n.kind == "VirtualGrid") && !n.disabled {
		return true
	}
	for _, draw := range n.draws {
		if draw.kind == "hitRegion" && !draw.disabled && draw.target == id {
			return true
		}
	}
	for _, child := range n.children {
		if hasFocusable(child, id) {
			return true
		}
	}
	return false
}

func firstFocusable(n *UiNode) string {
	if n == nil {
		return ""
	}
	if (n.kind == "Button" || n.kind == "TextInput" || n.kind == "VirtualGrid") && n.tabStop && !n.disabled {
		return n.id
	}
	for _, draw := range n.draws {
		if draw.kind == "hitRegion" && draw.tabStop && !draw.disabled {
			return draw.target
		}
	}
	for _, child := range n.children {
		if id := firstFocusable(child); id != "" {
			return id
		}
	}
	return ""
}

func focusableIDs(n *UiNode, ids *[]string) {
	if n == nil {
		return
	}
	if (n.kind == "Button" || n.kind == "TextInput" || n.kind == "VirtualGrid") && n.tabStop && !n.disabled {
		*ids = append(*ids, n.id)
	}
	for _, draw := range n.draws {
		if draw.kind == "hitRegion" && draw.tabStop && !draw.disabled {
			*ids = append(*ids, draw.target)
		}
	}
	for _, child := range n.children {
		focusableIDs(child, ids)
	}
}

func (w *windowsState) focusNode(id string) {
	if len(w.focusScopes) > 0 && (id == "" || !w.allowedInScope(id)) {
		return
	}
	if id == w.focused {
		if id != "" {
			w.scrollFocusIntoView(id)
			if input := w.inputs[id]; input != nil {
				// The retained focus ID can outlive native focus changes caused by
				// a tree update or another child window receiving focus.
				focused, _, _ := uiGetFocus.Call()
				if focused != input.hwnd {
					uiSetFocus.Call(input.hwnd)
				}
			}
		}
		return
	}
	old := w.focused
	w.focused = id
	if id != "" {
		w.scrollFocusIntoView(id)
	}
	w.invalidateButton(old)
	w.invalidateButton(id)
	if input := w.inputs[id]; input != nil {
		uiSetFocus.Call(input.hwnd)
	} else if w.hwnd != 0 {
		uiSetFocus.Call(w.hwnd)
	}
	w.emit(uiEvent{kind: "FocusChanged", nodeID: id})
	w.raiseUIAFocus(id)
}

func (w *windowsState) scrollFocusIntoView(id string) {
	var visit func(*uiBox) (bool, int64, int64, int64, int64)
	visit = func(box *uiBox) (bool, int64, int64, int64, int64) {
		if box == nil {
			return false, 0, 0, 0, 0
		}
		if box.node.id == id {
			if box.node.kind == "VirtualGrid" {
				if x, y, width, height, ok := virtualGridSelectionRect(box); ok {
					return true, x, y, width, height
				}
				return true, box.x, box.y, min64(box.w, box.node.columnWidth), min64(box.h, box.node.rowHeight)
			}
			return true, box.x, box.y, box.w, box.h
		}
		for _, draw := range box.node.draws {
			if draw.kind == "hitRegion" && draw.target == id {
				return true, box.x + draw.x, box.y + draw.y, draw.w, draw.h
			}
		}
		for _, child := range box.children {
			found, x, y, width, height := visit(child)
			if !found {
				continue
			}
			if box.node.kind == "ScrollView" || box.node.kind == "VirtualList" || box.node.kind == "VirtualGrid" {
				var dx, dy int64
				if x < box.x {
					dx = x - box.x
				} else if x+width > box.x+box.w {
					dx = x + width - box.x - box.w
				}
				if y < box.y {
					dy = y - box.y
				} else if y+height > box.y+box.h {
					dy = y + height - box.y - box.h
				}
				oldX, oldY := box.scrollX, box.scrollY
				if scrollBy(box, dx, dy) {
					w.invalidateBox(box)
					x -= box.scrollX - oldX
					y -= box.scrollY - oldY
				}
			}
			return true, x, y, width, height
		}
		return false, 0, 0, 0, 0
	}
	visit(w.layout)
	w.positionInputs()
}

func (w *windowsState) advanceFocus(reverse bool) {
	w.focusVisible = true
	var ids []string
	focusableIDs(w.activeFocusRoot(), &ids)
	if len(ids) == 0 {
		return
	}
	index := -1
	for i, id := range ids {
		if id == w.focused {
			index = i
			break
		}
	}
	if reverse {
		if index < 0 {
			index = 0
		}
		index = (index - 1 + len(ids)) % len(ids)
	} else {
		index = (index + 1) % len(ids)
	}
	w.focusNode(ids[index])
}

func (w *windowsState) positionInputs() {
	w.updateInputFont()
	if w.layout == nil {
		return
	}
	for id, input := range w.inputs {
		box, visible := visibleInputBox(w.layout, id, inputClip{0, 0, w.width, w.height})
		if box == nil || box.w <= 0 || box.h <= 0 || visible.right <= visible.left || visible.bottom <= visible.top {
			hideNativeInput(input)
			continue
		}
		x, y, width, height := box.x, box.y, box.w, box.h
		if box.node.customSurface {
			x += 10
			y += 6
			width -= 20
			height -= 12
			if width < 1 || height < 1 {
				hideNativeInput(input)
				continue
			}
		}
		if x+width <= 0 || y+height <= 0 || x >= w.width || y >= w.height {
			hideNativeInput(input)
			continue
		}
		scale := w.scale
		if scale <= 0 {
			scale = 1
		}
		px, py := int32(float64(x)*scale), int32(float64(y)*scale)
		pw, ph := int32(float64(width)*scale), int32(float64(height)*scale)
		clip := windowsRect{
			left:   int32(float64(max64(0, visible.left-x)) * scale),
			top:    int32(float64(max64(0, visible.top-y)) * scale),
			right:  int32(float64(min64(width, visible.right-x)) * scale),
			bottom: int32(float64(min64(height, visible.bottom-y)) * scale),
		}
		if clip.right <= clip.left || clip.bottom <= clip.top {
			hideNativeInput(input)
			continue
		}
		moved := !input.positioned || input.x != px || input.y != py || input.width != pw || input.height != ph
		clipChanged := !input.positioned || input.clip != clip
		if moved {
			// A scrolling EDIT must paint its new content; copied pixels can leave text from the old row.
			uiSetWindowPos.Call(input.hwnd, 0, uintptr(px), uintptr(py), uintptr(pw), uintptr(ph), swpNoZOrder|swpNoActivate|swpNoCopyBits)
		}
		if clipChanged {
			// Win32 owns this region after a successful SetWindowRgn call.
			region, _, _ := uiTextCreateRegion.Call(uintptr(clip.left), uintptr(clip.top), uintptr(clip.right), uintptr(clip.bottom))
			if region != 0 {
				if ok, _, _ := uiTextSetRegion.Call(input.hwnd, region, 1); ok == 0 {
					uiDeleteObject.Call(region)
				}
			}
		}
		if !input.shown {
			uiTextShowWindow.Call(input.hwnd, swShow)
			input.shown = true
		}
		if moved || clipChanged {
			uiInvalidateRect.Call(input.hwnd, 0, 1)
		}
		input.x, input.y, input.width, input.height = px, py, pw, ph
		input.clip, input.positioned = clip, true
	}
}

func hideNativeInput(input *nativeTextInput) {
	if input.shown {
		uiTextShowWindow.Call(input.hwnd, swHide)
		input.shown = false
	}
	input.positioned = false
}

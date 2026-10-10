//go:build windows

package stdlib

import (
	"fmt"
	"hash/fnv"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

// A small, window-owned UIA fragment. UIA may retain providers after an update,
// so removed elements stay allocated until their final COM Release and report
// element-not-available instead of exposing a stale UiNode.
const (
	wmGetObject     = 0x003d
	wmUiaInvoke     = 0x8001
	wmUiaFocus      = 0x8002
	uiaRootObjectID = uint32(0xffffffE7) // UiaRootObjectId (-25)
	uiaNoInterface  = uintptr(0x80004002)
	uiaUnavailable  = uintptr(0x80040201)
)

type uiaRect struct{ left, top, width, height float64 }
type uiaVariant struct {
	vt       uint16
	reserved [6]byte
	value    uintptr
	extra    uintptr
}
type uiaProvider struct {
	simple       uiaSimpleProvider
	fragment     uiaFragmentProvider
	fragmentRoot uiaFragmentRootProvider
	invoke       uiaInvokeProvider
	refs         atomic.Int32
	pin          runtime.Pinner
	w            *windowsState
	id           string
	parent       *uiaProvider
	children     []*uiaProvider
	live         bool
	root         bool
	gridCell     bool
	gridRow      int64
	gridColumn   int64
}
type uiaSimpleProvider struct{ vtbl *[7]uintptr }
type uiaFragmentProvider struct{ vtbl *[9]uintptr }
type uiaFragmentRootProvider struct{ vtbl *[5]uintptr }
type uiaInvokeProvider struct {
	vtbl   *[4]uintptr
	parent *uiaProvider
}

var (
	uiaCore           = syscall.NewLazyDLL("uiautomationcore.dll")
	uiaReturn         = uiaCore.NewProc("UiaReturnRawElementProvider")
	uiaHost           = uiaCore.NewProc("UiaHostProviderFromHwnd")
	uiaRaise          = uiaCore.NewProc("UiaRaiseAutomationEvent")
	uiaStructure      = uiaCore.NewProc("UiaRaiseStructureChangedEvent")
	uiaOle            = syscall.NewLazyDLL("oleaut32.dll")
	uiaSysAlloc       = uiaOle.NewProc("SysAllocStringLen")
	uiaSafeArray      = uiaOle.NewProc("SafeArrayCreateVector")
	uiaSafePut        = uiaOle.NewProc("SafeArrayPutElement")
	uiaClientToScreen = uiUser32.NewProc("ClientToScreen")
	uiaPostMessage    = uiUser32.NewProc("PostMessageW")
	uiaObjects        sync.Map // COM interface pointer -> *uiaProvider; roots are held by windowsState too.
	uiaSimpleVtable   = [7]uintptr{
		syscall.NewCallback(uiaQueryInterface), syscall.NewCallback(uiaAddRef), syscall.NewCallback(uiaRelease),
		syscall.NewCallback(uiaOptions), syscall.NewCallback(uiaPattern), syscall.NewCallback(uiaProperty), syscall.NewCallback(uiaHostProvider),
	}
	uiaFragmentVtable = [9]uintptr{syscall.NewCallback(uiaQueryInterface), syscall.NewCallback(uiaAddRef), syscall.NewCallback(uiaRelease), syscall.NewCallback(uiaNavigate), syscall.NewCallback(uiaRuntimeID), syscall.NewCallback(uiaBounds), syscall.NewCallback(uiaEmbedded), syscall.NewCallback(uiaSetFocus), syscall.NewCallback(uiaFragmentRoot)}
	uiaRootVtable     = [5]uintptr{syscall.NewCallback(uiaQueryInterface), syscall.NewCallback(uiaAddRef), syscall.NewCallback(uiaRelease), syscall.NewCallback(uiaFromPoint), syscall.NewCallback(uiaGetFocus)}
	uiaInvokeVtable   = [4]uintptr{syscall.NewCallback(uiaQueryInterface), syscall.NewCallback(uiaAddRef), syscall.NewCallback(uiaRelease), syscall.NewCallback(uiaInvoke)}
)

var (
	iidIUnknown = syscall.GUID{Data1: 0x00000000, Data2: 0x0000, Data3: 0x0000, Data4: [8]byte{0xc0, 0, 0, 0, 0, 0, 0, 0x46}}
	iidSimple   = syscall.GUID{Data1: 0xd6dd68d1, Data2: 0x86fd, Data3: 0x4332, Data4: [8]byte{0x86, 0x66, 0x9a, 0xbe, 0xde, 0xa2, 0xd2, 0x4c}}
	iidFragment = syscall.GUID{Data1: 0xf7063da8, Data2: 0x8359, Data3: 0x439c, Data4: [8]byte{0x92, 0x97, 0xbb, 0xc5, 0x29, 0x9a, 0x7d, 0x87}}
	iidRoot     = syscall.GUID{Data1: 0x620ce2a5, Data2: 0xab8f, Data3: 0x40a9, Data4: [8]byte{0x86, 0xcb, 0xde, 0x3c, 0x75, 0x59, 0x9b, 0x58}}
	iidInvoke   = syscall.GUID{Data1: 0x54fcb24b, Data2: 0xe18e, Data3: 0x47a2, Data4: [8]byte{0xb4, 0xd3, 0xec, 0xcb, 0xe7, 0x75, 0x99, 0xa2}}
)

func uiaOf(ptr uintptr) *uiaProvider {
	if value, ok := uiaObjects.Load(ptr); ok {
		return value.(*uiaProvider)
	}
	return nil
}
func (p *uiaProvider) ptr() uintptr         { return uintptr(unsafe.Pointer(&p.simple)) }
func (p *uiaProvider) fragmentPtr() uintptr { return uintptr(unsafe.Pointer(&p.fragment)) }
func (p *uiaProvider) rootPtr() uintptr     { return uintptr(unsafe.Pointer(&p.fragmentRoot)) }
func (p *uiaProvider) invokePtr() uintptr   { return uintptr(unsafe.Pointer(&p.invoke)) }
func (p *uiaProvider) addRef() uintptr      { return uintptr(p.refs.Add(1)) }
func (p *uiaProvider) release() uintptr {
	n := p.refs.Add(-1)
	if n == 0 {
		p.pin.Unpin()
		uiaObjects.Delete(p.ptr())
		uiaObjects.Delete(p.fragmentPtr())
		uiaObjects.Delete(p.rootPtr())
		uiaObjects.Delete(p.invokePtr())
		p.w = nil
		p.parent = nil
		p.children = nil
	}
	if n < 0 {
		return 0
	}
	return uintptr(n)
}
func newUIAProvider(w *windowsState, id string, root bool) *uiaProvider {
	p := &uiaProvider{w: w, id: id, live: true, root: root}
	p.pin.Pin(p)
	p.simple.vtbl = &uiaSimpleVtable
	p.fragment.vtbl = &uiaFragmentVtable
	p.fragmentRoot.vtbl = &uiaRootVtable
	p.invoke = uiaInvokeProvider{vtbl: &uiaInvokeVtable, parent: p}
	p.refs.Store(1)
	uiaObjects.Store(p.ptr(), p)
	uiaObjects.Store(p.fragmentPtr(), p)
	uiaObjects.Store(p.rootPtr(), p)
	uiaObjects.Store(p.invokePtr(), p)
	return p
}
func uiaGUIDEqual(a, b *syscall.GUID) bool {
	return a.Data1 == b.Data1 && a.Data2 == b.Data2 && a.Data3 == b.Data3 && a.Data4 == b.Data4
}
func uiaQueryInterface(this, riid, out uintptr) uintptr {
	if out == 0 || riid == 0 {
		return uiaNoInterface
	}
	*(*uintptr)(unsafe.Pointer(out)) = 0
	p := uiaOf(this)
	if p == nil {
		return uiaNoInterface
	}
	id := (*syscall.GUID)(unsafe.Pointer(riid))
	result := uintptr(0)
	switch {
	case uiaGUIDEqual(id, &iidIUnknown), uiaGUIDEqual(id, &iidSimple):
		result = p.ptr()
	case uiaGUIDEqual(id, &iidFragment):
		result = p.fragmentPtr()
	case uiaGUIDEqual(id, &iidRoot) && p.root:
		result = p.rootPtr()
	case uiaGUIDEqual(id, &iidInvoke) && p.actionable():
		result = p.invokePtr()
	}
	if result == 0 {
		return uiaNoInterface
	}
	*(*uintptr)(unsafe.Pointer(out)) = result
	p.addRef()
	return 0
}
func uiaAddRef(this uintptr) uintptr {
	if p := uiaOf(this); p != nil {
		return p.addRef()
	}
	return 0
}
func uiaRelease(this uintptr) uintptr {
	if p := uiaOf(this); p != nil {
		return p.release()
	}
	return 0
}
func uiaOptions(this, out uintptr) uintptr {
	if out != 0 {
		*(*int32)(unsafe.Pointer(out)) = 1
	}
	return 0
} // ServerSideProvider
func (p *uiaProvider) node() *UiNode {
	if p == nil || !p.live || p.w == nil {
		return nil
	}
	return uiNodeByID(p.w.root, p.id)
}
func (p *uiaProvider) draw() *UiDraw {
	if p == nil || !p.live || p.w == nil {
		return nil
	}
	var visit func(*UiNode) *UiDraw
	visit = func(n *UiNode) *UiDraw {
		if n == nil {
			return nil
		}
		for i := range n.draws {
			if n.draws[i].kind == "hitRegion" && n.draws[i].target == p.id {
				return &n.draws[i]
			}
		}
		for _, c := range n.children {
			if d := visit(c); d != nil {
				return d
			}
		}
		return nil
	}
	return visit(p.w.root)
}
func (p *uiaProvider) valid() bool {
	return p != nil && p.live && p.w != nil && p.w.hwnd != 0 && (p.root || p.node() != nil || p.draw() != nil)
}
func (p *uiaProvider) actionable() bool {
	if !p.valid() || p.root {
		return false
	}
	if n := p.node(); n != nil {
		return n.kind == "Button" && !n.disabled
	}
	if d := p.draw(); d != nil {
		return !d.disabled
	}
	return false
}
func uiaPattern(this, pattern, out uintptr) uintptr {
	if out == 0 {
		return uiaNoInterface
	}
	*(*uintptr)(unsafe.Pointer(out)) = 0
	p := uiaOf(this)
	if p != nil && pattern == 10000 && p.actionable() {
		*(*uintptr)(unsafe.Pointer(out)) = p.invokePtr()
		p.addRef()
	}
	return 0
}
func uiaSetVariantBool(v *uiaVariant, b bool) {
	v.vt = 11
	if b {
		v.value = 0xffff
	}
}
func uiaSetVariantString(v *uiaVariant, s string) {
	chars := utf16.Encode([]rune(s))
	if len(chars) == 0 {
		v.vt = 8
		v.value, _, _ = uiaSysAlloc.Call(0, 0)
		return
	}
	v.vt = 8
	v.value, _, _ = uiaSysAlloc.Call(uintptr(unsafe.Pointer(&chars[0])), uintptr(len(chars)))
	runtime.KeepAlive(chars)
}
func uiaProperty(this, property, out uintptr) uintptr {
	if out == 0 {
		return uiaNoInterface
	}
	v := (*uiaVariant)(unsafe.Pointer(out))
	*v = uiaVariant{}
	p := uiaOf(this)
	if p == nil || !p.valid() {
		return uiaUnavailable
	}
	n := p.node()
	d := p.draw()
	switch property {
	case 30003: // ControlType
		v.vt = 3
		v.value = 50033 // Pane
		if n != nil {
			switch n.kind {
			case "Button":
				v.value = 50000
			case "Text":
				v.value = 50020
			case "VirtualGrid":
				v.value = 50028 // DataGrid
			default:
				v.value = 50026
			}
		}
		if p.gridCell {
			v.value = 50029 // DataItem
		}
		if d != nil {
			v.value = 50000
		}
	case 30005: // Name
		name := "Craft UI"
		if n != nil {
			name = n.accessibleName
			if name == "" {
				name = n.text
			}
			if n.kind == "VirtualGrid" && name == "" {
				name = fmt.Sprintf("Grid, %d rows, %d columns", n.itemCount, n.columnCount)
			}
			if p.gridCell {
				name = fmt.Sprintf("Row %d, column %d", p.gridRow+1, p.gridColumn+1)
				if n.text != "" {
					name += ": " + n.text
				}
			}
		}
		if d != nil {
			name = d.accessibleName
		}
		uiaSetVariantString(v, name)
	case 30011:
		uiaSetVariantString(v, p.id) // AutomationId
	case 30010: // IsEnabled
		enabled := true
		if n != nil {
			enabled = !n.disabled
		}
		if d != nil {
			enabled = !d.disabled
		}
		uiaSetVariantBool(v, enabled)
	case 30009:
		uiaSetVariantBool(v, !p.root && p.w.focusCandidate(p.id) == nil) // IsKeyboardFocusable
	case 30008:
		uiaSetVariantBool(v, p.w.focused == p.id) // HasKeyboardFocus
	case 30016, 30017:
		uiaSetVariantBool(v, true) // IsControlElement / IsContentElement
	}
	return 0
}
func uiaHostProvider(this, out uintptr) uintptr {
	if out == 0 {
		return uiaNoInterface
	}
	*(*uintptr)(unsafe.Pointer(out)) = 0
	p := uiaOf(this)
	if p == nil || !p.valid() {
		return uiaUnavailable
	}
	if p.root {
		result, _, _ := uiaHost.Call(p.w.hwnd, out)
		return result
	}
	return 0
}
func uiaReturnPtr(p *uiaProvider, out uintptr, rootInterface bool) uintptr {
	if out == 0 {
		return uiaNoInterface
	}
	*(*uintptr)(unsafe.Pointer(out)) = 0
	if p != nil && p.valid() {
		if rootInterface {
			*(*uintptr)(unsafe.Pointer(out)) = p.rootPtr()
		} else {
			*(*uintptr)(unsafe.Pointer(out)) = p.fragmentPtr()
		}
		p.addRef()
	}
	return 0
}
func uiaNavigate(this, direction, out uintptr) uintptr {
	p := uiaOf(this)
	if p == nil || !p.valid() {
		return uiaUnavailable
	}
	var dest *uiaProvider
	switch direction {
	case 0:
		dest = p.parent // Parent
	case 1:
		if p.parent != nil {
			for i, sibling := range p.parent.children {
				if sibling == p && i+1 < len(p.parent.children) {
					dest = p.parent.children[i+1]
					break
				}
			}
		}
	case 2:
		if p.parent != nil {
			for i, sibling := range p.parent.children {
				if sibling == p && i > 0 {
					dest = p.parent.children[i-1]
					break
				}
			}
		}
	case 3:
		if len(p.children) > 0 {
			dest = p.children[0]
		}
	case 4:
		if len(p.children) > 0 {
			dest = p.children[len(p.children)-1]
		}
	}
	return uiaReturnPtr(dest, out, false)
}
func uiaRuntimeID(this, out uintptr) uintptr {
	if out == 0 {
		return uiaNoInterface
	}
	*(*uintptr)(unsafe.Pointer(out)) = 0
	p := uiaOf(this)
	if p == nil || !p.valid() {
		return uiaUnavailable
	}
	if p.root {
		return 0
	}
	sa, _, _ := uiaSafeArray.Call(3, 0, 2)
	if sa == 0 {
		return uintptr(0x8007000e)
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(p.id))
	values := [2]int32{3, int32(h.Sum32())}
	for i := int32(0); i < 2; i++ {
		index := i
		uiaSafePut.Call(sa, uintptr(unsafe.Pointer(&index)), uintptr(unsafe.Pointer(&values[i])))
	}
	*(*uintptr)(unsafe.Pointer(out)) = sa
	return 0
}
func (p *uiaProvider) geometry() (int64, int64, int64, int64, bool) {
	if p.root {
		return 0, 0, p.w.width, p.w.height, true
	}
	if box := nodeBox(p.w.layout, p.id); box != nil {
		return box.x, box.y, box.w, box.h, true
	}
	var visit func(*uiBox) (int64, int64, int64, int64, bool)
	visit = func(b *uiBox) (int64, int64, int64, int64, bool) {
		if b == nil {
			return 0, 0, 0, 0, false
		}
		for _, d := range b.node.draws {
			if d.kind == "hitRegion" && d.target == p.id {
				return b.x + d.x, b.y + d.y, d.w, d.h, true
			}
		}
		for _, c := range b.children {
			if x, y, w, h, ok := visit(c); ok {
				return x, y, w, h, true
			}
		}
		return 0, 0, 0, 0, false
	}
	return visit(p.w.layout)
}
func uiaBounds(this, out uintptr) uintptr {
	if out == 0 {
		return uiaNoInterface
	}
	*(*uiaRect)(unsafe.Pointer(out)) = uiaRect{}
	p := uiaOf(this)
	if p == nil || !p.valid() {
		return uiaUnavailable
	}
	x, y, w, h, ok := p.geometry()
	if !ok {
		return uiaUnavailable
	}
	point := windowsPoint{}
	uiaClientToScreen.Call(p.w.hwnd, uintptr(unsafe.Pointer(&point)))
	s := p.w.scale
	if s <= 0 {
		s = 1
	}
	*(*uiaRect)(unsafe.Pointer(out)) = uiaRect{float64(point.x) + float64(x)*s, float64(point.y) + float64(y)*s, float64(w) * s, float64(h) * s}
	return 0
}
func uiaEmbedded(this, out uintptr) uintptr {
	if out != 0 {
		*(*uintptr)(unsafe.Pointer(out)) = 0
	}
	return 0
}
func uiaSetFocus(this uintptr) uintptr {
	p := uiaOf(this)
	if p == nil || !p.valid() || p.root || p.w.focusCandidate(p.id) != nil {
		return uiaUnavailable
	}
	p.addRef()
	ok, _, _ := uiaPostMessage.Call(p.w.hwnd, wmUiaFocus, p.ptr(), 0)
	if ok == 0 {
		p.release()
		return uiaUnavailable
	}
	return 0
}
func uiaFragmentRoot(this, out uintptr) uintptr {
	p := uiaOf(this)
	if p == nil || !p.valid() {
		return uiaUnavailable
	}
	return uiaReturnPtr(p.w.uiaRoot, out, true)
}
func uiaFromPoint(this, x, y, out uintptr) uintptr {
	p := uiaOf(this)
	if p == nil || !p.valid() {
		return uiaUnavailable
	}
	return uiaReturnPtr(p, out, false)
}
func uiaGetFocus(this, out uintptr) uintptr {
	p := uiaOf(this)
	if p == nil || !p.valid() {
		return uiaUnavailable
	}
	return uiaReturnPtr(p.w.uiaProviders[p.w.focused], out, false)
}
func uiaInvoke(this uintptr) uintptr {
	p := uiaOf(this)
	if p == nil || !p.actionable() {
		return uiaUnavailable
	}
	p.addRef()
	ok, _, _ := uiaPostMessage.Call(p.w.hwnd, wmUiaInvoke, p.ptr(), 0)
	if ok == 0 {
		p.release()
		return uiaUnavailable
	}
	return 0
}

func (w *windowsState) uiaRootProvider() *uiaProvider {
	if w.uiaRoot == nil {
		w.uiaRoot = newUIAProvider(w, "@window", true)
	}
	return w.uiaRoot
}
func (w *windowsState) rebuildUIA() {
	root := w.uiaRootProvider()
	old := w.uiaProviders
	next := make(map[string]*uiaProvider)
	oldCount := len(old)
	root.children = nil
	var walk func(*UiNode, *uiaProvider)
	walk = func(n *UiNode, parent *uiaProvider) {
		if n == nil {
			return
		}
		include := n.kind == "Button" || n.kind == "Text" || n.kind == "VirtualGrid" || (n.accessibleName != "" && n.kind != "TextInput")
		if include {
			p := old[n.id]
			if p == nil {
				p = newUIAProvider(w, n.id, false)
			}
			p.live = true
			p.gridCell = false
			p.parent = parent
			p.children = nil
			parent.children = append(parent.children, p)
			next[n.id] = p
			parent = p
		}
		for _, d := range n.draws {
			if d.kind == "hitRegion" && d.accessibleName != "" {
				p := old[d.target]
				if p == nil {
					p = newUIAProvider(w, d.target, false)
				}
				p.live = true
				p.parent = parent
				p.children = nil
				parent.children = append(parent.children, p)
				next[d.target] = p
			}
		}
		for index, child := range n.children {
			walk(child, parent)
			if n.kind == "VirtualGrid" && int64(index) < int64(gridCellCount(n)) && n.loadedColumns > 0 {
				if cell := next[child.id]; cell != nil {
					cell.gridCell = true
					cell.gridRow = n.firstIndex + int64(index)/n.loadedColumns
					cell.gridColumn = n.firstColumn + int64(index)%n.loadedColumns
				}
			}
		}
	}
	walk(w.root, root)
	for id, p := range old {
		if next[id] == nil {
			p.live = false
			p.parent = nil
			p.children = nil
			p.release()
		}
	}
	w.uiaProviders = next
	structureChanged := oldCount != len(next)
	if !structureChanged {
		for id := range old {
			if next[id] == nil {
				structureChanged = true
				break
			}
		}
	}
	if structureChanged {
		uiaStructure.Call(root.ptr(), 2, 0, 0)
	}
	w.raiseUIAFocus(w.focused)
}
func (w *windowsState) closeUIA() {
	if w.hwnd != 0 && w.uiaRoot != nil {
		uiaReturn.Call(w.hwnd, 0, 0, 0)
	}
	for _, p := range w.uiaProviders {
		p.live = false
		p.release()
	}
	w.uiaProviders = nil
	if w.uiaRoot != nil {
		w.uiaRoot.live = false
		w.uiaRoot.release()
		w.uiaRoot = nil
	}
}
func (w *windowsState) raiseUIAFocus(id string) {
	if p := w.uiaProviders[id]; p != nil && p.valid() {
		uiaRaise.Call(p.ptr(), 20005)
	}
}
func (w *windowsState) raiseUIAInvoke(id string) {
	if p := w.uiaProviders[id]; p != nil && p.valid() {
		uiaRaise.Call(p.ptr(), 20009)
	}
}

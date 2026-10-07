//go:build windows

package stdlib

import (
	"context"
	"os"
	"testing"
	"time"
	"unsafe"
)

func TestNativeScrollViewInputAndClick(t *testing.T) {
	if os.Getenv("CRAFT_UI_INTEGRATION") != "1" {
		t.Skip("set CRAFT_UI_INTEGRATION=1 to open a native window")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	u, err := startUI(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { close(u.commands); u.wakeHost(); <-u.done }()
	if _, err := u.send(ctx, uiCommand{op: "open", title: "Craft scroll integration test", w: 640, h: 480}); err != nil {
		t.Fatal(err)
	}
	var state *windowsState
	uiWindows.Range(func(_, value any) bool { state = value.(*windowsState); return false })
	if state == nil || state.hwnd == 0 {
		t.Fatal("window host was not registered")
	}
	if _, err := u.send(ctx, uiCommand{op: "std.ui.mount", root: testGrid(20, 20)}); err != nil {
		t.Fatal(err)
	}
	if state.layout == nil {
		t.Fatal("mount did not create layout")
	}
	send := uiUser32.NewProc("SendMessageW")
	send.Call(state.hwnd, wmKeyDown, 0x22, 0) // PageDown
	if state.layout.scrollY <= 0 {
		t.Fatal("PageDown did not move the viewport")
	}
	before := state.layout.scrollY
	point := windowsPoint{x: 100, y: 100}
	uiUser32.NewProc("ClientToScreen").Call(state.hwnd, uintptr(unsafe.Pointer(&point)))
	position := uintptr(uint16(point.x)) | uintptr(uint16(point.y))<<16
	wheelDown := uintptr(uint16(0xff88)) << 16 // -120
	send.Call(state.hwnd, wmMouseWheel, wheelDown, position)
	if state.layout.scrollY <= before {
		t.Fatal("mouse wheel did not move the viewport")
	}
	send.Call(state.hwnd, wmMouseHWheel, uintptr(uint16(120))<<16, position)
	if state.layout.scrollX <= 0 {
		t.Fatal("horizontal wheel did not move the viewport")
	}
	var pointX, pointY int64
	var want string
	for y := int64(10); y < state.height && want == ""; y += 10 {
		for x := int64(10); x < state.width; x += 10 {
			if id := hitButton(state.layout, x, y); id != "" {
				pointX, pointY, want = x, y, id
				break
			}
		}
	}
	if want == "" {
		t.Fatalf("no visible button after scrolling: viewport=%dx%d offset=%d,%d content=%dx%d", state.width, state.height, state.layout.scrollX, state.layout.scrollY, state.layout.children[0].w, state.layout.children[0].h)
	}
	click := uintptr(uint16(int64(float64(pointX)*state.scale))) | uintptr(uint16(int64(float64(pointY)*state.scale)))<<16
	send.Call(state.hwnd, wmLeftButtonDown, 0, click)
	send.Call(state.hwnd, wmLeftButtonUp, 0, click)
	for len(u.events) > 0 {
		event := <-u.events
		if event.kind == "ButtonPressed" {
			if event.nodeID != want {
				t.Fatalf("clicked %q, expected %q", event.nodeID, want)
			}
			if _, err := u.send(ctx, uiCommand{op: "std.ui.update", root: testGrid(2, 2)}); err != nil {
				t.Fatal(err)
			}
			if state.layout.scrollX != 0 || state.layout.scrollY != 0 {
				t.Fatal("scroll offset did not clamp after update shrank content")
			}
			return
		}
	}
	t.Fatal("click after scroll did not emit ButtonPressed")
}

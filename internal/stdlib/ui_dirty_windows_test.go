//go:build windows

package stdlib

import "testing"

func TestDirtyRegionsMergeAndCanvasOldNewBounds(t *testing.T) {
	w := &windowsState{width: 500, height: 400, scale: 1}
	dirty := newUIDirtyRegions(w)
	dirty.addLogical(10, 10, 20, 20)
	dirty.addLogical(30, 10, 20, 20)
	if len(dirty.regions) != 1 {
		t.Fatalf("expected merged regions, got %d", len(dirty.regions))
	}
	box := &uiBox{x: 0, y: 0, w: 500, h: 200}
	old := []UiDraw{{kind: "fillRect", x: 40, y: 80, w: 20, h: 20}}
	next := []UiDraw{{kind: "fillRect", x: 360, y: 80, w: 20, h: 20}}
	if !dirty.addCanvasChanges(box, old, next) {
		t.Fatal("canvas diff should be bounded")
	}
	if len(dirty.regions) != 3 {
		t.Fatalf("old and new positions both need invalidation, got %d regions", len(dirty.regions))
	}
}

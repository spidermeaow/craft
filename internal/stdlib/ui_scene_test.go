package stdlib

import "testing"

func testTextMeasure(n *UiNode) (int64, int64) { return int64(len(n.text)) * 8, 22 }

func testGrid(rows, columns int) *UiNode {
	grid := &UiNode{kind: "Column", id: "grid"}
	for row := 0; row < rows; row++ {
		line := &UiNode{kind: "Row", id: "row-" + string(rune(row+65))}
		for column := 0; column < columns; column++ {
			line.children = append(line.children, &UiNode{kind: "Button", id: "cell-" + string(rune(row+65)) + string(rune(column+65)), text: "cell"})
		}
		grid.children = append(grid.children, line)
	}
	return &UiNode{kind: "ScrollView", id: "scroll", children: []*UiNode{grid}}
}

func TestRev23SplitAndWeightedShell(t *testing.T) {
	left := &UiNode{kind: "ScrollView", id: "left", children: []*UiNode{{kind: "Column", id: "left-content", height: 900}}}
	right := &UiNode{kind: "ScrollView", id: "right", children: []*UiNode{{kind: "Column", id: "right-content", height: 900}}}
	split := &UiNode{kind: "SplitView", id: "split", ratio: 30, splitMinFirst: 220, splitMinSecond: 260, weight: 1, children: []*UiNode{left, right}}
	root := &UiNode{kind: "Column", id: "root", children: []*UiNode{{kind: "Text", id: "header", height: 50}, split, {kind: "Text", id: "status", height: 34}}}
	if err := validateTree(root); err != nil {
		t.Fatal(err)
	}
	for _, width := range []int64{1000, 420} {
		box := measureNode(root, testTextMeasure)
		var overflow int64
		placeNode(box, 0, 0, width, 700, &overflow)
		pane := box.children[1]
		if pane.children[0].w+pane.children[1].w+uiGap != pane.w {
			t.Fatalf("width %d: split children exceed pane", width)
		}
		if pane.children[1].x != pane.children[0].x+pane.children[0].w+uiGap {
			t.Fatalf("width %d: split gap incorrect", width)
		}
		if box.children[2].y+box.children[2].h > 700 {
			t.Fatalf("width %d: status escaped viewport", width)
		}
		if _, _, _, _, ok := scrollbarThumb(pane.children[0]); !ok {
			t.Fatalf("width %d: missing vertical scrollbar", width)
		}
		chain := scrollChainAt(box, pane.children[0].x+10, pane.children[0].y+10)
		if len(chain) != 1 || chain[0].node.id != "left" {
			t.Fatalf("width %d: wrong scroll target", width)
		}
	}
}

func TestRev23LayoutSpecInvalidatesRebind(t *testing.T) {
	first := &UiNode{kind: "Column", id: "root", children: []*UiNode{{kind: "Text", id: "a", text: "same", width: 100}}}
	box := measureNode(first, testTextMeasure)
	second := cloneNode(first)
	second.children[0].width = 200
	var changed []*uiBox
	if rebindLayout(box, second, &changed) {
		t.Fatal("layout reused after width changed")
	}
}

func TestAdaptiveTextElsewhereDoesNotForceCanvasRelayout(t *testing.T) {
	first := &UiNode{kind: "Column", id: "root", children: []*UiNode{
		{kind: "Text", id: "help", text: "Unchanged wrapping help", wrapText: true},
		{kind: "Canvas", id: "draw", width: 200, height: 100, draws: []UiDraw{{kind: "fillRect", x: 10, y: 10, w: 20, h: 20, color: 0xff123456}}},
	}}
	box := measureNode(first, testTextMeasure)
	next := cloneNode(first)
	next.children[1].draws[0].x = 80
	var changed []uiChange
	if !rebindLayoutDetailed(box, next, &changed) || len(changed) != 1 || changed[0].box.node.id != "draw" {
		t.Fatalf("canvas change should rebind with stable adaptive text; changes=%d", len(changed))
	}
	wrapped := cloneNode(next)
	wrapped.children[0].text = "A different wrapping help string"
	if rebindLayoutDetailed(box, wrapped, &changed) {
		t.Fatal("changed wrapping text must relayout")
	}
}

func TestScrollViewLayoutHitAndClamp(t *testing.T) {
	root := testGrid(20, 20)
	if err := validateTree(root); err != nil {
		t.Fatal(err)
	}
	box := measureNode(root, testTextMeasure)
	var overflow int64
	placeNode(box, 0, 0, 320, 240, &overflow)
	if box.scrollX != 0 || box.scrollY != 0 {
		t.Fatal("initial scroll is not zero")
	}
	first := box.children[0].children[0].children[0]
	if got := hitButton(box, first.x+1, first.y+1); got != first.node.id {
		t.Fatalf("initial hit=%q, want %q", got, first.node.id)
	}
	if !scrollBy(box, 1<<20, 1<<20) {
		t.Fatal("large grid did not scroll")
	}
	if box.scrollX != box.children[0].w-box.w || box.scrollY != box.children[0].h-box.h {
		t.Fatal("offset was not clamped to content edge")
	}
	last := box.children[0].children[19].children[19]
	if got := hitButton(box, last.x+last.w-1, last.y+last.h-1); got != last.node.id {
		t.Fatalf("last visible hit=%q, want %q", got, last.node.id)
	}
	if hitButton(box, box.w+1, box.h-1) != "" {
		t.Fatal("hit escaped the viewport")
	}
	if scrollBy(box, 1, 1) {
		t.Fatal("scroll moved past content edge")
	}
	offsets := map[string][4]int64{}
	scrollOffsets(box, offsets)
	smaller := measureNode(testGrid(2, 2), testTextMeasure)
	restoreScroll(smaller, offsets)
	placeNode(smaller, 0, 0, 320, 240, &overflow)
	if smaller.scrollX != 0 || smaller.scrollY != 0 {
		t.Fatal("offset did not clamp after content shrank")
	}
}

func TestScrollViewOwnsSpareSpaceAndFillsContentWidth(t *testing.T) {
	controls := &UiNode{kind: "Column", id: "controls", children: []*UiNode{{kind: "Text", id: "title", text: "Stress Lab"}, {kind: "Button", id: "action", text: "Update all"}}}
	scroll := testGrid(8, 8)
	root := &UiNode{kind: "Row", id: "root", children: []*UiNode{controls, scroll}}
	box := measureNode(root, testTextMeasure)
	var overflow int64
	placeNode(box, 0, 0, 1100, 700, &overflow)
	left, viewport := box.children[0], box.children[1]
	if left.w != left.prefW {
		t.Fatalf("controls stretched from %d to %d although ScrollView needs the spare space", left.prefW, left.w)
	}
	if viewport.w <= viewport.prefW {
		t.Fatalf("ScrollView did not receive spare width: %d", viewport.w)
	}
	content := viewport.children[0]
	if content.w != viewport.w || content.h != content.prefH {
		t.Fatalf("scroll content should fill width and keep intrinsic height: viewport %dx%d, content %dx%d", viewport.w, viewport.h, content.w, content.h)
	}
	if content.w <= viewport.w && viewport.scrollX != 0 {
		t.Fatal("horizontal offset exists without horizontal overflow")
	}
}

func TestVirtualListRangeLayoutAndClamp(t *testing.T) {
	list := &UiNode{kind: "VirtualList", id: "list", itemCount: 100000, rowHeight: 40}
	if err := validateTree(list); err != nil {
		t.Fatal(err)
	}
	box := measureNode(list, testTextMeasure)
	var overflow int64
	placeNode(box, 0, 0, 320, 240, &overflow)
	first, end := virtualRange(box)
	if first != 0 || end != 9 {
		t.Fatalf("initial range=%d..%d", first, end)
	}
	list.children = make([]*UiNode, 0, end-first)
	for i := first; i < end; i++ {
		list.children = append(list.children, &UiNode{kind: "Button", id: "item-" + string(rune(i+65)), text: "item"})
	}
	box = measureNode(list, testTextMeasure)
	placeNode(box, 0, 0, 320, 240, &overflow)
	if got := hitButton(box, 10, 10); got != "item-A" {
		t.Fatalf("first hit=%q", got)
	}
	if !scrollBy(box, 0, 1<<30) {
		t.Fatal("did not scroll to end")
	}
	if box.scrollY != virtualExtent(list)-box.h {
		t.Fatalf("end offset=%d", box.scrollY)
	}
	first, end = virtualRange(box)
	if end != list.itemCount || end-first > 12 {
		t.Fatalf("end range=%d..%d", first, end)
	}
	offsets := map[string][4]int64{}
	scrollOffsets(box, offsets)
	short := &UiNode{kind: "VirtualList", id: "list", itemCount: 3, rowHeight: 40}
	shortBox := measureNode(short, testTextMeasure)
	restoreScroll(shortBox, offsets)
	placeNode(shortBox, 0, 0, 320, 240, &overflow)
	if shortBox.scrollY != 0 {
		t.Fatalf("shrink offset=%d", shortBox.scrollY)
	}
}

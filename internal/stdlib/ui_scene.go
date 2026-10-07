package stdlib

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

// UiNode is an immutable description copied into a UI session on mount/update.
type UiNode struct {
	kind, id, text, placeholder                                                     string
	textUTF16                                                                       []uint16
	children                                                                        []*UiNode
	draws                                                                           []UiDraw
	itemCount, rowHeight, firstIndex                                                int64
	width, height, minWidth, minHeight, maxWidth, maxHeight, weight, alignX, alignY int64
	padding, gap, ratio, splitMinFirst, splitMinSecond                              int64
	customSpacing                                                                   bool
	fontSize, fontWeight                                                            int64
	textColor                                                                       uint32
	customTextStyle                                                                 bool
	wrapText                                                                        bool
	gridMinCell, gridMaxColumns, gridGap                                            int64
	surfaceFill, surfaceBorder                                                      uint32
	surfaceRadius                                                                   int64
	customSurface                                                                   bool
	tabStop, disabled                                                               bool
	accessibleName, accessibleRole                                                  string
}

func (n *UiNode) String() string { return fmt.Sprintf("UiNode(%s:%s)", n.kind, n.id) }

type uiBox struct {
	node             *UiNode
	x, y, w, h       int64
	prefW, prefH     int64
	scrollX, scrollY int64
	wheelX, wheelY   int64
	hasScroll        bool
	hasAdaptive      bool
	measure          uiTextMeasure
	lines            []string
	lineHeight       int64
	children         []*uiBox
}

const virtualOverscan int64 = 3

func virtualExtent(n *UiNode) int64 { return n.itemCount * n.rowHeight }

func virtualRange(b *uiBox) (int64, int64) {
	if b == nil || b.node.kind != "VirtualList" || b.node.itemCount == 0 || b.h <= 0 {
		return 0, 0
	}
	first := max64(0, b.scrollY/b.node.rowHeight-virtualOverscan)
	last := min64(b.node.itemCount, (b.scrollY+b.h+b.node.rowHeight-1)/b.node.rowHeight+virtualOverscan)
	return first, max64(first, last)
}

func makeNode(kind, id, label string, children []*UiNode) (*UiNode, error) {
	if id == "" || len(id) > 128 || !utf8.ValidString(id) || strings.ContainsRune(id, 0) || len(label) > 4096 || !utf8.ValidString(label) || strings.ContainsRune(label, 0) {
		return nil, fmt.Errorf("UiArgumentError: invalid node id or text")
	}
	node := &UiNode{kind: kind, id: id, text: label, children: children, tabStop: true}
	if kind == "TextInput" {
		if err := validateInputValue(label); err != nil {
			return nil, err
		}
	}
	if kind == "Text" || kind == "Button" || kind == "TextInput" {
		node.textUTF16 = utf16.Encode([]rune(label))
	}
	return node, nil
}

func validateInputValue(value string) error {
	if !utf8.ValidString(value) || strings.ContainsRune(value, 0) || strings.ContainsAny(value, "\r\n") || len(utf16.Encode([]rune(value))) > 1024 {
		return fmt.Errorf("UiArgumentError: text input must be one line of at most 1024 UTF-16 units")
	}
	return nil
}

func validateTree(root *UiNode) error {
	if root == nil {
		return fmt.Errorf("UiTreeError: root is nil")
	}
	seen := map[string]bool{}
	targets := map[string]bool{}
	drawCount := 0
	var walk func(*UiNode, int) error
	walk = func(n *UiNode, depth int) error {
		if n == nil || depth > 64 || len(seen) >= 4096 {
			return fmt.Errorf("UiTreeError: tree exceeds limits")
		}
		if seen[n.id] {
			return fmt.Errorf("UiTreeError: duplicate node id %q", n.id)
		}
		seen[n.id] = true
		if n.kind == "Canvas" {
			regions := map[string]bool{}
			for _, draw := range n.draws {
				if draw.kind == "hitRegion" {
					if regions[draw.target] || targets[draw.target] || seen[draw.target] {
						return fmt.Errorf("UiTreeError: duplicate Canvas target %q", draw.target)
					}
					regions[draw.target] = true
					targets[draw.target] = true
				}
			}
		}
		drawCount += len(n.draws)
		if drawCount > 8192 || len(n.draws) > maxCanvasCommands || (n.kind != "Canvas" && len(n.draws) != 0) {
			return fmt.Errorf("UiTreeError: Canvas command limit exceeded")
		}
		if n.kind == "ScrollView" && len(n.children) != 1 {
			return fmt.Errorf("UiTreeError: ScrollView requires exactly one child")
		}
		if n.kind == "SplitView" && len(n.children) != 2 {
			return fmt.Errorf("UiTreeError: SplitView requires exactly two children")
		}
		if n.kind == "VirtualList" {
			if n.itemCount < 0 || n.itemCount > 100000000 || n.rowHeight < 16 || n.rowHeight > 8192 || n.itemCount > (1<<40)/n.rowHeight || n.firstIndex < 0 || n.firstIndex > n.itemCount || int64(len(n.children)) > n.itemCount-n.firstIndex || len(n.children) > 1024 {
				return fmt.Errorf("UiTreeError: invalid VirtualList range or dimensions")
			}
		}
		if n.kind == "Grid" && (n.gridMinCell < 1 || n.gridMinCell > 8192 || n.gridMaxColumns < 1 || n.gridMaxColumns > 64 || n.gridGap < 0 || n.gridGap > 256) {
			return fmt.Errorf("UiTreeError: invalid Grid constraints")
		}
		if n.kind != "Row" && n.kind != "Column" && n.kind != "Panel" && n.kind != "SplitView" && n.kind != "ScrollView" && n.kind != "VirtualList" && n.kind != "Grid" && len(n.children) != 0 {
			return fmt.Errorf("UiTreeError: leaf node has children")
		}
		for _, child := range n.children {
			if err := walk(child, depth+1); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(root, 0); err != nil {
		return err
	}
	for id := range targets {
		if seen[id] {
			return fmt.Errorf("UiTreeError: Canvas target collides with node id %q", id)
		}
	}
	return nil
}

func cloneNode(n *UiNode) *UiNode {
	if n == nil {
		return nil
	}
	c := &UiNode{kind: n.kind, id: n.id, text: n.text, placeholder: n.placeholder, textUTF16: append([]uint16(nil), n.textUTF16...), itemCount: n.itemCount, rowHeight: n.rowHeight, firstIndex: n.firstIndex,
		width: n.width, height: n.height, minWidth: n.minWidth, minHeight: n.minHeight, maxWidth: n.maxWidth, maxHeight: n.maxHeight, weight: n.weight, alignX: n.alignX, alignY: n.alignY, padding: n.padding, gap: n.gap, fontSize: n.fontSize, fontWeight: n.fontWeight, textColor: n.textColor, customTextStyle: n.customTextStyle, wrapText: n.wrapText, gridMinCell: n.gridMinCell, gridMaxColumns: n.gridMaxColumns, gridGap: n.gridGap, surfaceFill: n.surfaceFill, surfaceBorder: n.surfaceBorder, surfaceRadius: n.surfaceRadius, customSurface: n.customSurface,
		ratio: n.ratio, splitMinFirst: n.splitMinFirst, splitMinSecond: n.splitMinSecond, customSpacing: n.customSpacing,
		tabStop: n.tabStop, disabled: n.disabled, accessibleName: n.accessibleName, accessibleRole: n.accessibleRole}
	for _, draw := range n.draws {
		c.draws = append(c.draws, cloneDraw(draw))
	}
	for _, child := range n.children {
		c.children = append(c.children, cloneNode(child))
	}
	return c
}

type uiTextMeasure func(*UiNode) (int64, int64)

const uiPad, uiGap int64 = 6, 4

func nodeSpacing(n *UiNode) (int64, int64) {
	if n.customSpacing {
		return n.padding, n.gap
	}
	return uiPad, uiGap
}

func applyPreferredSize(b *uiBox) {
	n := b.node
	if n.width > 0 {
		b.prefW = n.width
	}
	if n.height > 0 {
		b.prefH = n.height
	}
	b.prefW = max64(b.prefW, n.minWidth)
	b.prefH = max64(b.prefH, n.minHeight)
	if n.maxWidth > 0 {
		b.prefW = min64(b.prefW, n.maxWidth)
	}
	if n.maxHeight > 0 {
		b.prefH = min64(b.prefH, n.maxHeight)
	}
}

func measureNode(n *UiNode, measure uiTextMeasure) *uiBox {
	b := &uiBox{node: n, hasScroll: n.kind == "ScrollView" || n.kind == "VirtualList", hasAdaptive: n.kind == "Grid" || n.wrapText, measure: measure}
	switch n.kind {
	case "Canvas":
		b.prefW, b.prefH = n.width, n.height
		applyPreferredSize(b)
		return b
	case "Text", "Button", "TextInput":
		b.prefW, b.prefH = measure(n)
		if n.kind == "TextInput" {
			b.prefW = max64(240, b.prefW+20)
			b.prefH = max64(34, b.prefH+12)
		} else if n.kind == "Button" {
			b.prefW += 20
			b.prefH += 12
			if b.prefH < 32 {
				b.prefH = 32
			}
		} else {
			b.prefW += 4
			b.prefH += 4
			if b.prefH < 24 {
				b.prefH = 24
			}
		}
		if n.kind == "Text" && n.wrapText {
			b.lineHeight = max64(1, b.prefH-4)
			b.prefW = 1
		}
		applyPreferredSize(b)
		return b
	}
	for _, child := range n.children {
		measured := measureNode(child, measure)
		b.children = append(b.children, measured)
		b.hasScroll = b.hasScroll || measured.hasScroll
		b.hasAdaptive = b.hasAdaptive || measured.hasAdaptive
	}
	if n.kind == "VirtualList" {
		for _, child := range b.children {
			b.prefW = max64(b.prefW, child.prefW)
		}
		b.prefW = max64(240, b.prefW)
		b.prefH = min64(160, virtualExtent(n))
		applyPreferredSize(b)
		return b
	}
	if n.kind == "ScrollView" {
		b.prefW = min64(240, b.children[0].prefW)
		b.prefH = min64(160, b.children[0].prefH)
		applyPreferredSize(b)
		return b
	}
	if n.kind == "Grid" {
		b.prefW = n.gridMinCell
		for _, child := range b.children {
			b.prefH += child.prefH
		}
		b.prefH += n.gridGap * int64(max64(0, int64(len(b.children))-1))
		applyPreferredSize(b)
		return b
	}
	if len(b.children) == 0 {
		applyPreferredSize(b)
		return b
	}
	pad, gap := nodeSpacing(n)
	if n.kind == "Column" {
		for _, child := range b.children {
			if child.prefW > b.prefW {
				b.prefW = child.prefW
			}
			b.prefH += child.prefH
		}
		b.prefH += gap * int64(len(b.children)-1)
	} else if n.kind == "Row" || n.kind == "SplitView" {
		for _, child := range b.children {
			b.prefW += child.prefW
			if child.prefH > b.prefH {
				b.prefH = child.prefH
			}
		}
		b.prefW += gap * int64(len(b.children)-1)
	} else {
		for _, child := range b.children {
			if child.prefW > b.prefW {
				b.prefW = child.prefW
			}
			if child.prefH > b.prefH {
				b.prefH = child.prefH
			}
		}
	}
	if n.kind != "SplitView" {
		b.prefW += 2 * pad
		b.prefH += 2 * pad
	}
	if n.kind == "SplitView" {
		b.prefW = max64(b.prefW, n.splitMinFirst+n.splitMinSecond+gap)
	}
	applyPreferredSize(b)
	return b
}

// Adaptive nodes are measured a second time after their parent supplies width.
// The same stored line breaks are used for painting, so geometry and pixels agree.
func reflowBox(b *uiBox, width int64) {
	if b == nil || !b.hasAdaptive {
		return
	}
	n := b.node
	if n.kind == "VirtualList" {
		for _, child := range b.children {
			reflowBox(child, width)
		}
		b.prefH = min64(160, virtualExtent(n))
		applyPreferredSize(b)
		return
	}
	if n.kind == "Text" && n.wrapText {
		b.lines = wrapLines(n, max64(1, width-4), b.measure)
		b.prefW = max64(1, width)
		b.prefH = max64(24, int64(len(b.lines))*b.lineHeight+4)
		applyPreferredSize(b)
		return
	}
	if len(b.children) == 0 {
		return
	}
	if n.kind == "ScrollView" {
		child := b.children[0]
		contentWidth := max64(width, child.prefW)
		reflowBox(child, contentWidth)
		b.prefH = min64(160, child.prefH)
		applyPreferredSize(b)
		return
	}
	if n.kind == "Grid" {
		columns, cell := gridDimensions(max64(width, n.gridMinCell), n.gridMinCell, n.gridMaxColumns, n.gridGap)
		var total, rowHeight int64
		for i, child := range b.children {
			reflowBox(child, cell)
			rowHeight = max64(rowHeight, child.prefH)
			if int64(i+1)%columns == 0 || i == len(b.children)-1 {
				if total > 0 {
					total += n.gridGap
				}
				total += rowHeight
				rowHeight = 0
			}
		}
		b.prefW, b.prefH = n.gridMinCell, total
		applyPreferredSize(b)
		return
	}
	pad, gap := nodeSpacing(n)
	inner := max64(0, width-2*pad)
	if n.kind == "Row" || n.kind == "SplitView" {
		if n.kind == "SplitView" {
			available := max64(0, width-uiGap)
			first := clamp64(available*n.ratio/100, n.splitMinFirst, max64(n.splitMinFirst, available-n.splitMinSecond))
			first = min64(first, available)
			reflowBox(b.children[0], first)
			reflowBox(b.children[1], available-first)
			b.prefH = max64(b.children[0].prefH, b.children[1].prefH)
			applyPreferredSize(b)
			return
		}
		available := max64(0, inner-gap*int64(len(b.children)-1))
		widths := make([]int64, len(b.children))
		var fixed, totalWeight, flexible, scrolls int64
		custom := false
		for i, child := range b.children {
			widths[i] = child.prefW
			fixed += child.prefW
			totalWeight += child.node.weight
			if child.node.width > 0 || child.node.height > 0 || child.node.minWidth > 0 || child.node.minHeight > 0 || child.node.weight > 0 || child.node.maxWidth > 0 || child.node.maxHeight > 0 || child.node.alignX > 0 || child.node.alignY > 0 {
				custom = true
			}
			if child.node.kind == "ScrollView" || child.node.kind == "VirtualList" {
				scrolls++
			}
			if child.node.kind == "Row" || child.node.kind == "Column" || child.node.kind == "Panel" || child.node.kind == "Grid" || child.node.kind == "ScrollView" || child.node.kind == "VirtualList" || child.node.wrapText {
				flexible++
			}
		}
		if fixed < available {
			extra := available - fixed
			if custom && totalWeight > 0 {
				for i, child := range b.children {
					if child.node.weight > 0 {
						widths[i] += extra * child.node.weight / totalWeight
					}
				}
			} else if !custom {
				if scrolls > 0 {
					flexible = scrolls
				}
				if flexible > 0 {
					for i, child := range b.children {
						if (scrolls > 0 && (child.node.kind == "ScrollView" || child.node.kind == "VirtualList")) || (scrolls == 0 && (child.node.kind == "Row" || child.node.kind == "Column" || child.node.kind == "Panel" || child.node.kind == "Grid" || child.node.wrapText)) {
							widths[i] += extra / flexible
						}
					}
				}
			}
		} else if fixed > available && custom {
			deficit := fixed - available
			for pass := 0; pass < 2 && deficit > 0; pass++ {
				for i, child := range b.children {
					if pass == 0 && child.node.weight == 0 {
						continue
					}
					cut := min64(deficit, max64(0, widths[i]-child.node.minWidth))
					widths[i] -= cut
					deficit -= cut
				}
			}
		}
		for i, child := range b.children {
			reflowBox(child, widths[i])
		}
		b.prefH = 0
		for _, child := range b.children {
			b.prefH = max64(b.prefH, child.prefH)
		}
		if n.kind != "SplitView" {
			b.prefH += 2 * pad
		}
	} else {
		b.prefH = 2 * pad
		if n.kind == "Panel" {
			b.prefH = 0
		}
		for i, child := range b.children {
			childWidth := inner
			if child.node.width > 0 {
				childWidth = child.node.width
			}
			reflowBox(child, childWidth)
			if n.kind == "Column" {
				b.prefH += child.prefH
				if i > 0 {
					b.prefH += gap
				}
			} else {
				b.prefH = max64(b.prefH, child.prefH+2*pad)
			}
		}
	}
	applyPreferredSize(b)
}

func gridDimensions(width, minimum, maximum, gap int64) (int64, int64) {
	columns := clamp64((width+gap)/(minimum+gap), 1, maximum)
	cell := max64(1, (max64(width, minimum)-(columns-1)*gap)/columns)
	return columns, cell
}

func wrapLines(n *UiNode, width int64, measure uiTextMeasure) []string {
	var lines []string
	for _, paragraph := range strings.Split(strings.ReplaceAll(n.text, "\r\n", "\n"), "\n") {
		runes := []rune(paragraph)
		if len(runes) == 0 {
			lines = append(lines, "")
			continue
		}
		for len(runes) > 0 {
			fits := func(count int) bool {
				candidate := *n
				candidate.text = string(runes[:count])
				candidate.textUTF16 = utf16.Encode(runes[:count])
				measured, _ := measure(&candidate)
				return measured <= width
			}
			fit := 1
			if fits(1) {
				probe := 2
				for probe <= len(runes) && fits(probe) {
					fit = probe
					probe *= 2
				}
				lo, hi := fit+1, minInt(len(runes), probe-1)
				for lo <= hi {
					mid := (lo + hi) / 2
					if fits(mid) {
						fit = mid
						lo = mid + 1
					} else {
						hi = mid - 1
					}
				}
			}
			if fit < len(runes) && unicode.IsMark(runes[fit]) {
				start := fit - 1
				for start > 0 && unicode.IsMark(runes[start]) {
					start--
				}
				if start > 0 {
					fit = start
				} else {
					for fit < len(runes) && unicode.IsMark(runes[fit]) {
						fit++
					}
				}
			}
			if fit < len(runes) {
				for i := fit - 1; i > 0; i-- {
					if runes[i] == ' ' || runes[i] == '\t' {
						fit = i + 1
						break
					}
				}
			}
			lines = append(lines, strings.TrimRight(string(runes[:fit]), " \t"))
			runes = runes[fit:]
			for len(runes) > 0 && (runes[0] == ' ' || runes[0] == '\t') {
				runes = runes[1:]
			}
		}
	}
	return lines
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func placeNode(b *uiBox, x, y, w, h int64, overflow *int64) {
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	b.x, b.y, b.w, b.h = x, y, w, h
	if w < b.prefW || h < b.prefH {
		*overflow++
	}
	if b.node.kind == "VirtualList" {
		b.scrollY = clamp64(b.scrollY, 0, max64(0, virtualExtent(b.node)-h))
		for i, child := range b.children {
			row := b.node.firstIndex + int64(i)
			placeNode(child, x, y+row*b.node.rowHeight-b.scrollY, max64(w, child.prefW), b.node.rowHeight, overflow)
		}
		return
	}
	if len(b.children) == 0 {
		return
	}
	if b.node.kind == "ScrollView" {
		child := b.children[0]
		// Fill the viewport horizontally while retaining intrinsic overflow.
		// Keep the intrinsic height so vertical scrolling remains available.
		contentW := max64(w, child.prefW)
		b.scrollX = clamp64(b.scrollX, 0, max64(0, contentW-w))
		b.scrollY = clamp64(b.scrollY, 0, max64(0, child.prefH-h))
		placeNode(child, x-b.scrollX, y-b.scrollY, contentW, child.prefH, overflow)
		return
	}
	if b.node.kind == "Grid" {
		n := b.node
		columns, cell := gridDimensions(max64(w, n.gridMinCell), n.gridMinCell, n.gridMaxColumns, n.gridGap)
		cursorY := y
		for start := 0; start < len(b.children); start += int(columns) {
			end := int64(start) + columns
			if end > int64(len(b.children)) {
				end = int64(len(b.children))
			}
			var rowHeight int64
			for i := start; i < int(end); i++ {
				rowHeight = max64(rowHeight, b.children[i].prefH)
			}
			for i := start; i < int(end); i++ {
				placeNode(b.children[i], x+int64(i-start)*(cell+n.gridGap), cursorY, cell, rowHeight, overflow)
			}
			cursorY += rowHeight + n.gridGap
		}
		return
	}
	if b.node.kind == "SplitView" {
		gap := uiGap
		available := max64(0, w-gap)
		first := available * b.node.ratio / 100
		minimumFirst := max64(b.node.splitMinFirst, b.children[0].node.minWidth)
		minimumSecond := max64(b.node.splitMinSecond, b.children[1].node.minWidth)
		if available >= minimumFirst+minimumSecond {
			first = clamp64(first, minimumFirst, available-minimumSecond)
		} else if minimumFirst+minimumSecond > 0 {
			// When the viewport cannot fit both minima, keep both panes visible.
			first = available * minimumFirst / (minimumFirst + minimumSecond)
		}
		placeNode(b.children[0], x, y, first, h, overflow)
		placeNode(b.children[1], x+first+gap, y, available-first, h, overflow)
		return
	}
	pad, gap := nodeSpacing(b.node)
	innerW, innerH := w-2*pad, h-2*pad
	if innerW < 0 {
		innerW = 0
	}
	if innerH < 0 {
		innerH = 0
	}
	if b.node.kind == "Panel" {
		for _, child := range b.children {
			ox, cw := alignedSize(innerW, child.prefW, child.node.maxWidth, child.node.alignX)
			oy, ch := alignedSize(innerH, child.prefH, child.node.maxHeight, child.node.alignY)
			placeNode(child, x+pad+ox, y+pad+oy, cw, ch, overflow)
		}
		return
	}
	mainSpace := innerW
	if b.node.kind == "Column" {
		mainSpace = innerH
	}
	mainSpace = max64(0, mainSpace-gap*int64(len(b.children)-1))
	custom := false
	for _, child := range b.children {
		if child.node.width > 0 || child.node.height > 0 || child.node.minWidth > 0 || child.node.minHeight > 0 || child.node.maxWidth > 0 || child.node.maxHeight > 0 || child.node.weight > 0 || child.node.alignX > 0 || child.node.alignY > 0 {
			custom = true
			break
		}
	}
	if custom {
		placeConstrainedChildren(b, x+pad, y+pad, innerW, innerH, mainSpace, gap, overflow)
		return
	}
	var preferred int64
	var flexible int64
	var scrollViews int64
	for _, child := range b.children {
		if b.node.kind == "Column" {
			preferred += child.prefH
		} else {
			preferred += child.prefW
		}
		if child.node.kind == "Row" || child.node.kind == "Column" || child.node.kind == "Panel" || child.node.kind == "Grid" || child.node.kind == "ScrollView" || child.node.kind == "VirtualList" || (b.node.kind == "Row" && child.node.wrapText) {
			flexible++
		}
		if child.node.kind == "ScrollView" || child.node.kind == "VirtualList" {
			scrollViews++
		}
	}
	if scrollViews > 0 {
		flexible = scrollViews
	}
	extra := int64(0)
	if mainSpace > preferred && flexible > 0 {
		extra = (mainSpace - preferred) / flexible
	}
	cursor := int64(0)
	for _, child := range b.children {
		grow := int64(0)
		if (scrollViews > 0 && (child.node.kind == "ScrollView" || child.node.kind == "VirtualList")) || (scrollViews == 0 && (child.node.kind == "Row" || child.node.kind == "Column" || child.node.kind == "Panel" || child.node.kind == "Grid" || (b.node.kind == "Row" && child.node.wrapText))) {
			grow = extra
		}
		if b.node.kind == "Column" {
			ox, cw := alignedSize(innerW, child.prefW, child.node.maxWidth, child.node.alignX)
			placeNode(child, x+pad+ox, y+pad+cursor, cw, child.prefH+grow, overflow)
			cursor += child.prefH + grow + gap
		} else {
			oy, ch := alignedSize(innerH, child.prefH, child.node.maxHeight, child.node.alignY)
			placeNode(child, x+pad+cursor, y+pad+oy, child.prefW+grow, ch, overflow)
			cursor += child.prefW + grow + gap
		}
	}
}

func alignedSize(available, preferred, maximum, alignment int64) (int64, int64) {
	size := available
	if alignment != 0 {
		size = min64(preferred, available)
	}
	if maximum > 0 {
		size = min64(size, maximum)
	}
	size = max64(0, size)
	switch alignment {
	case 2:
		return max64(0, (available-size)/2), size
	case 3:
		return max64(0, available-size), size
	default:
		return 0, size
	}
}

func placeConstrainedChildren(b *uiBox, x, y, innerW, innerH, available, gap int64, overflow *int64) {
	column := b.node.kind == "Column"
	sizes := make([]int64, len(b.children))
	var total, totalWeight int64
	for i, child := range b.children {
		if column {
			sizes[i] = child.prefH
		} else {
			sizes[i] = child.prefW
		}
		if child.node.weight > 0 {
			totalWeight += child.node.weight
		}
		total += sizes[i]
	}
	if total < available && totalWeight > 0 {
		extra := available - total
		remaining := extra
		for i, child := range b.children {
			if child.node.weight == 0 {
				continue
			}
			grow := extra * child.node.weight / totalWeight
			sizes[i] += grow
			remaining -= grow
		}
		for i, child := range b.children {
			if remaining == 0 {
				break
			}
			if child.node.weight > 0 {
				sizes[i]++
				remaining--
			}
		}
		for i, child := range b.children {
			maximum := child.node.maxWidth
			if column {
				maximum = child.node.maxHeight
			}
			if maximum > 0 {
				sizes[i] = min64(sizes[i], maximum)
			}
		}
	}
	if total > available {
		// Reduce flexible children first, then all children if the viewport is too narrow.
		deficit := total - available
		for pass := 0; pass < 2 && deficit > 0; pass++ {
			for i, child := range b.children {
				if deficit == 0 {
					break
				}
				if pass == 0 && child.node.weight == 0 {
					continue
				}
				minimum := child.node.minWidth
				if column {
					minimum = child.node.minHeight
				}
				cut := min64(deficit, max64(0, sizes[i]-minimum))
				sizes[i] -= cut
				deficit -= cut
			}
		}
		// An impossible set of minima is clipped within the viewport.
		for i := len(sizes) - 1; i >= 0 && deficit > 0; i-- {
			cut := min64(deficit, sizes[i])
			sizes[i] -= cut
			deficit -= cut
		}
	}
	cursor := int64(0)
	for i, child := range b.children {
		if column {
			ox, cw := alignedSize(innerW, child.prefW, child.node.maxWidth, child.node.alignX)
			placeNode(child, x+ox, y+cursor, cw, sizes[i], overflow)
		} else {
			oy, ch := alignedSize(innerH, child.prefH, child.node.maxHeight, child.node.alignY)
			placeNode(child, x+cursor, y+oy, sizes[i], ch, overflow)
		}
		cursor += sizes[i] + gap
	}
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
func clamp64(v, lo, hi int64) int64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func scrollOffsets(box *uiBox, offsets map[string][4]int64) {
	if box == nil {
		return
	}
	if box.node.kind == "ScrollView" || box.node.kind == "VirtualList" {
		offsets[box.node.id] = [4]int64{box.scrollX, box.scrollY, box.wheelX, box.wheelY}
	}
	for _, child := range box.children {
		scrollOffsets(child, offsets)
	}
}

func restoreScroll(box *uiBox, offsets map[string][4]int64) {
	if box == nil {
		return
	}
	if box.node.kind == "ScrollView" || box.node.kind == "VirtualList" {
		if old, ok := offsets[box.node.id]; ok {
			box.scrollX, box.scrollY, box.wheelX, box.wheelY = old[0], old[1], old[2], old[3]
		}
	}
	for _, child := range box.children {
		restoreScroll(child, offsets)
	}
}

func scrollAt(box *uiBox, x, y int64) *uiBox {
	if box == nil || !box.hasScroll || x < box.x || y < box.y || x >= box.x+box.w || y >= box.y+box.h {
		return nil
	}
	for i := len(box.children) - 1; i >= 0; i-- {
		if found := scrollAt(box.children[i], x, y); found != nil {
			return found
		}
	}
	if box.node.kind == "ScrollView" || box.node.kind == "VirtualList" {
		return box
	}
	return nil
}

func scrollChainAt(box *uiBox, x, y int64) []*uiBox {
	if box == nil || !box.hasScroll || x < box.x || y < box.y || x >= box.x+box.w || y >= box.y+box.h {
		return nil
	}
	for i := len(box.children) - 1; i >= 0; i-- {
		if chain := scrollChainAt(box.children[i], x, y); len(chain) > 0 {
			if box.node.kind == "ScrollView" || box.node.kind == "VirtualList" {
				return append(chain, box)
			}
			return chain
		}
	}
	if box.node.kind == "ScrollView" || box.node.kind == "VirtualList" {
		return []*uiBox{box}
	}
	return nil
}

func firstScroll(box *uiBox) *uiBox {
	if box == nil || !box.hasScroll {
		return nil
	}
	if box.node.kind == "ScrollView" || box.node.kind == "VirtualList" {
		return box
	}
	for _, child := range box.children {
		if found := firstScroll(child); found != nil {
			return found
		}
	}
	return nil
}

func scrollByID(box *uiBox, id string) *uiBox {
	if box == nil || !box.hasScroll || id == "" {
		return nil
	}
	if (box.node.kind == "ScrollView" || box.node.kind == "VirtualList") && box.node.id == id {
		return box
	}
	for _, child := range box.children {
		if found := scrollByID(child, id); found != nil {
			return found
		}
	}
	return nil
}

func scrollBy(box *uiBox, dx, dy int64) bool {
	if box == nil {
		return false
	}
	if box.node.kind == "VirtualList" {
		nextY := clamp64(box.scrollY+dy, 0, max64(0, virtualExtent(box.node)-box.h))
		if nextY == box.scrollY {
			return false
		}
		for _, child := range box.children {
			shiftBox(child, 0, box.scrollY-nextY)
		}
		box.scrollY = nextY
		return true
	}
	if box.node.kind != "ScrollView" || len(box.children) != 1 {
		return false
	}
	child := box.children[0]
	nextX := clamp64(box.scrollX+dx, 0, max64(0, child.w-box.w))
	nextY := clamp64(box.scrollY+dy, 0, max64(0, child.h-box.h))
	if nextX == box.scrollX && nextY == box.scrollY {
		return false
	}
	shiftBox(child, box.scrollX-nextX, box.scrollY-nextY)
	box.scrollX, box.scrollY = nextX, nextY
	return true
}

func scrollContentHeight(box *uiBox) int64 {
	if box == nil {
		return 0
	}
	if box.node.kind == "VirtualList" {
		return virtualExtent(box.node)
	}
	if len(box.children) > 0 {
		return box.children[0].h
	}
	return 0
}

func scrollbarThumb(box *uiBox) (int64, int64, int64, int64, bool) {
	if box == nil || (box.node.kind != "ScrollView" && box.node.kind != "VirtualList") || box.w < 16 || box.h < 32 {
		return 0, 0, 0, 0, false
	}
	content := scrollContentHeight(box)
	if content <= box.h {
		return 0, 0, 0, 0, false
	}
	width := int64(7)
	track := box.h - 8
	thumb := clamp64(track*box.h/content, 24, track)
	travel := track - thumb
	y := box.y + 4 + travel*box.scrollY/(content-box.h)
	return box.x + box.w - width - 4, y, width, thumb, true
}

func horizontalScrollbarThumb(box *uiBox) (int64, int64, int64, int64, bool) {
	if box == nil || box.node.kind != "ScrollView" || len(box.children) != 1 || box.w < 32 || box.h < 16 {
		return 0, 0, 0, 0, false
	}
	content := box.children[0].w
	if content <= box.w {
		return 0, 0, 0, 0, false
	}
	track := box.w - 8
	thumb := clamp64(track*box.w/content, 24, track)
	travel := track - thumb
	x := box.x + 4 + travel*box.scrollX/(content-box.w)
	return x, box.y + box.h - 11, thumb, 7, true
}

func shiftBox(box *uiBox, dx, dy int64) {
	box.x += dx
	box.y += dy
	for _, child := range box.children {
		shiftBox(child, dx, dy)
	}
}

func hitButton(b *uiBox, x, y int64) string {
	if b == nil || x < b.x || y < b.y || x >= b.x+b.w || y >= b.y+b.h {
		return ""
	}
	for i := len(b.children) - 1; i >= 0; i-- {
		if id := hitButton(b.children[i], x, y); id != "" {
			return id
		}
	}
	if b.node.kind == "Button" && !b.node.disabled {
		return b.node.id
	}
	return ""
}

func hitCanvas(b *uiBox, x, y int64) *uiBox {
	if b == nil || x < b.x || y < b.y || x >= b.x+b.w || y >= b.y+b.h {
		return nil
	}
	for i := len(b.children) - 1; i >= 0; i-- {
		if found := hitCanvas(b.children[i], x, y); found != nil {
			return found
		}
	}
	if b.node.kind == "Canvas" {
		return b
	}
	return nil
}

// The last declared region is topmost. Regions are local to a retained Canvas.
func hitCanvasTarget(box *uiBox, x, y int64) string {
	if box == nil || box.node.kind != "Canvas" {
		return ""
	}
	lx, ly := x-box.x, y-box.y
	for i := len(box.node.draws) - 1; i >= 0; i-- {
		d := box.node.draws[i]
		if d.kind == "hitRegion" && !d.disabled && lx >= d.x && ly >= d.y && lx < d.x+d.w && ly < d.y+d.h {
			return d.target
		}
	}
	return ""
}

func hasCanvasTarget(n *UiNode, id string) bool {
	if n == nil || id == "" {
		return false
	}
	for _, d := range n.draws {
		if d.kind == "hitRegion" && !d.disabled && d.target == id {
			return true
		}
	}
	for _, child := range n.children {
		if hasCanvasTarget(child, id) {
			return true
		}
	}
	return false
}

func buttonBox(b *uiBox, id string) *uiBox {
	if b == nil {
		return nil
	}
	if b.node.kind == "Button" && b.node.id == id {
		return b
	}
	for _, child := range b.children {
		if found := buttonBox(child, id); found != nil {
			return found
		}
	}
	return nil
}

func hasButton(n *UiNode, id string) bool {
	if n == nil || id == "" {
		return false
	}
	if n.kind == "Button" && n.id == id && !n.disabled {
		return true
	}
	for _, child := range n.children {
		if hasButton(child, id) {
			return true
		}
	}
	return false
}

func uiNodeByID(n *UiNode, id string) *UiNode {
	if n == nil || id == "" {
		return nil
	}
	if n.id == id {
		return n
	}
	for _, child := range n.children {
		if found := uiNodeByID(child, id); found != nil {
			return found
		}
	}
	return nil
}

func uiSubtreeHasID(n *UiNode, id string) bool {
	if n == nil || id == "" {
		return false
	}
	if uiNodeByID(n, id) != nil {
		return true
	}
	for _, draw := range n.draws {
		if draw.kind == "hitRegion" && draw.target == id {
			return true
		}
	}
	for _, child := range n.children {
		if uiSubtreeHasID(child, id) {
			return true
		}
	}
	return false
}

func sameTree(a, b *UiNode) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.kind != b.kind || a.id != b.id || a.text != b.text || a.placeholder != b.placeholder || a.itemCount != b.itemCount || a.rowHeight != b.rowHeight || a.firstIndex != b.firstIndex || a.textColor != b.textColor || a.customTextStyle != b.customTextStyle || a.wrapText != b.wrapText || a.surfaceFill != b.surfaceFill || a.surfaceBorder != b.surfaceBorder || a.surfaceRadius != b.surfaceRadius || a.customSurface != b.customSurface || a.tabStop != b.tabStop || a.disabled != b.disabled || a.accessibleName != b.accessibleName || a.accessibleRole != b.accessibleRole || !sameLayoutSpec(a, b) || !sameDraws(a.draws, b.draws) || len(a.children) != len(b.children) {
		return false
	}
	for i := range a.children {
		if !sameTree(a.children[i], b.children[i]) {
			return false
		}
	}
	return true
}

func sameLayoutSpec(a, b *UiNode) bool {
	return a.width == b.width && a.height == b.height && a.minWidth == b.minWidth && a.minHeight == b.minHeight && a.maxWidth == b.maxWidth && a.maxHeight == b.maxHeight && a.weight == b.weight && a.alignX == b.alignX && a.alignY == b.alignY && a.padding == b.padding && a.gap == b.gap && a.customSpacing == b.customSpacing && a.ratio == b.ratio && a.splitMinFirst == b.splitMinFirst && a.splitMinSecond == b.splitMinSecond && a.fontSize == b.fontSize && a.fontWeight == b.fontWeight && a.gridMinCell == b.gridMinCell && a.gridMaxColumns == b.gridMaxColumns && a.gridGap == b.gridGap && a.wrapText == b.wrapText
}

// rebindLayout reuses geometry only after the caller checks changed leaf metrics.
func rebindLayout(box *uiBox, next *UiNode, changed *[]*uiBox) bool {
	if box != nil && box.hasAdaptive {
		return false
	}
	if containsVirtualList(next) {
		return false
	}
	if box == nil || next == nil || box.node.kind != next.kind || box.node.id != next.id || !sameLayoutSpec(box.node, next) || len(box.children) != len(next.children) {
		return false
	}
	for i, child := range box.children {
		if !sameShape(child, next.children[i]) {
			return false
		}
	}
	var bind func(*uiBox, *UiNode)
	bind = func(current *uiBox, node *UiNode) {
		if current.node.text != node.text || current.node.textColor != node.textColor || current.node.customTextStyle != node.customTextStyle || current.node.surfaceFill != node.surfaceFill || current.node.surfaceBorder != node.surfaceBorder || current.node.surfaceRadius != node.surfaceRadius || current.node.customSurface != node.customSurface || current.node.disabled != node.disabled || !sameDraws(current.node.draws, node.draws) {
			*changed = append(*changed, current)
		}
		current.node = node
		for i, child := range current.children {
			bind(child, node.children[i])
		}
	}
	bind(box, next)
	return true
}

func containsVirtualList(n *UiNode) bool {
	if n == nil {
		return false
	}
	if n.kind == "VirtualList" {
		return true
	}
	for _, child := range n.children {
		if containsVirtualList(child) {
			return true
		}
	}
	return false
}

func sameShape(box *uiBox, next *UiNode) bool {
	if box == nil || next == nil || box.node.kind != next.kind || box.node.id != next.id || !sameLayoutSpec(box.node, next) || len(box.children) != len(next.children) {
		return false
	}
	for i, child := range box.children {
		if !sameShape(child, next.children[i]) {
			return false
		}
	}
	return true
}

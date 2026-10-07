package stdlib

import (
	"context"
	rt "craft/internal/runtime"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

// UiWindow is an opaque, session-bound handle. Copying it does not duplicate a window.
type UiWindow struct {
	owner *uiSession
	id    int64
}

func (w *UiWindow) String() string { return fmt.Sprintf("UiWindow(%d)", w.id) }

type uiEvent struct {
	kind                          string
	nodeID                        string
	text                          string
	windowID                      int64
	x, y, width, height, key, mod int64
	scale                         float64
	firstIndex, endIndex          int64
	at                            time.Time
}

func (e uiEvent) value() *rt.Struct {
	return nativeStruct("UiEvent", e.kind, e.windowID, e.x, e.y, e.width, e.height, e.key, e.mod, e.scale, e.nodeID, e.firstIndex, e.endIndex, e.text)
}

type uiCommand struct {
	op       string
	title    string
	w, h     int64
	key, mod int64
	r, g, b  int64
	root     *UiNode
	nodeID   string
	scopeID  string
	value    string
	answer   chan uiReply
	queuedAt time.Time
	origin   time.Time
}

type uiReply struct {
	id    int64
	value int64
	text  string
	err   error
}

type uiSession struct {
	commands  chan uiCommand
	events    chan uiEvent
	done      chan struct{}
	once      sync.Once
	failure   error
	mu        sync.Mutex
	closed    atomic.Bool
	wakeMu    sync.Mutex
	wake      uintptr // Windows auto-reset event; zero on other platforms.
	lastEvent time.Time
}

func (u *uiSession) fail(err error) {
	u.mu.Lock()
	u.failure = err
	u.mu.Unlock()
}

func (u *uiSession) stopped() error {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.failure != nil {
		return u.failure
	}
	return fmt.Errorf("UiSessionError: UI host stopped")
}

func (s *Session) Close() {
	if s.ui != nil {
		s.ui.once.Do(func() {
			close(s.ui.commands)
			s.ui.wakeHost()
			<-s.ui.done
		})
	}
	s.uiImages = nil
	s.uiImageBytes = 0
}

func (s *Session) invokeUI(ctx context.Context, name string, args []any) (any, error) {
	if name == "std.ui.color" {
		r, g, b, a := args[0].(int64), args[1].(int64), args[2].(int64), args[3].(int64)
		if r < 0 || r > 255 || g < 0 || g > 255 || b < 0 || b > 255 || a < 0 || a > 255 {
			return nil, fmt.Errorf("UiArgumentError: color channels must be 0..255")
		}
		return (a << 24) | (r << 16) | (g << 8) | b, nil
	}
	if name == "std.ui.fillRect" || name == "std.ui.strokeRect" || name == "std.ui.drawLine" || name == "std.ui.drawText" || name == "std.ui.drawTextStyled" || name == "std.ui.drawImage" || name == "std.ui.hitRegion" {
		return s.makeUIDraw(name, args)
	}
	if name == "std.ui.targetPolicy" {
		original, ok := args[0].(*UiDraw)
		if !ok || original == nil || original.kind != "hitRegion" {
			return nil, fmt.Errorf("UiTreeError: targetPolicy requires hitRegion")
		}
		label := args[3].(string)
		if label == "" || len(label) > 512 || !utf8.ValidString(label) || strings.ContainsRune(label, 0) {
			return nil, fmt.Errorf("UiArgumentError: Canvas target requires an accessible name")
		}
		draw := cloneDraw(*original)
		draw.tabStop, draw.disabled, draw.accessibleName = args[1].(bool), !args[2].(bool), label
		return &draw, nil
	}
	if name == "std.ui.canvas" {
		width, height := args[1].(int64), args[2].(int64)
		if width < 1 || width > 8192 || height < 1 || height > 8192 {
			return nil, fmt.Errorf("UiArgumentError: canvas size must be 1..8192")
		}
		array := args[3].(*rt.Array)
		if len(array.Items) > maxCanvasCommands {
			return nil, fmt.Errorf("UiTreeError: Canvas exceeds 2048 draw commands")
		}
		node, err := makeNode("Canvas", args[0].(string), "", nil)
		if err != nil {
			return nil, err
		}
		node.width, node.height = width, height
		for _, value := range array.Items {
			draw, ok := value.(*UiDraw)
			if !ok || draw == nil {
				return nil, fmt.Errorf("UiTreeError: Canvas command is not UiDraw")
			}
			node.draws = append(node.draws, cloneDraw(*draw))
		}
		return node, nil
	}
	if name == "std.ui.size" {
		original, ok := args[0].(*UiNode)
		if !ok || original == nil {
			return nil, fmt.Errorf("UiTreeError: size requires a UiNode")
		}
		width, height, minWidth, minHeight, weight := args[1].(int64), args[2].(int64), args[3].(int64), args[4].(int64), args[5].(int64)
		if width < 0 || width > 8192 || height < 0 || height > 8192 || minWidth < 0 || minWidth > 8192 || minHeight < 0 || minHeight > 8192 || weight < 0 || weight > 1000 || (width > 0 && minWidth > width) || (height > 0 && minHeight > height) {
			return nil, fmt.Errorf("UiArgumentError: invalid size constraints")
		}
		node := cloneNode(original)
		if (node.maxWidth > 0 && (minWidth > node.maxWidth || width > node.maxWidth)) || (node.maxHeight > 0 && (minHeight > node.maxHeight || height > node.maxHeight)) {
			return nil, fmt.Errorf("UiArgumentError: size exceeds maxSize")
		}
		node.width, node.height, node.minWidth, node.minHeight, node.weight = width, height, minWidth, minHeight, weight
		return node, nil
	}
	if name == "std.ui.maxSize" {
		original, ok := args[0].(*UiNode)
		if !ok || original == nil {
			return nil, fmt.Errorf("UiTreeError: maxSize requires a UiNode")
		}
		maxWidth, maxHeight := args[1].(int64), args[2].(int64)
		if maxWidth < 0 || maxWidth > 8192 || maxHeight < 0 || maxHeight > 8192 || (maxWidth > 0 && (maxWidth < original.minWidth || maxWidth < original.width)) || (maxHeight > 0 && (maxHeight < original.minHeight || maxHeight < original.height)) {
			return nil, fmt.Errorf("UiArgumentError: invalid maximum size")
		}
		node := cloneNode(original)
		node.maxWidth, node.maxHeight = maxWidth, maxHeight
		return node, nil
	}
	if name == "std.ui.align" {
		original, ok := args[0].(*UiNode)
		if !ok || original == nil {
			return nil, fmt.Errorf("UiTreeError: align requires a UiNode")
		}
		parse := func(value string) (int64, error) {
			switch value {
			case "stretch":
				return 0, nil
			case "start":
				return 1, nil
			case "center":
				return 2, nil
			case "end":
				return 3, nil
			}
			return 0, fmt.Errorf("UiArgumentError: alignment must be stretch, start, center or end")
		}
		ax, err := parse(args[1].(string))
		if err != nil {
			return nil, err
		}
		ay, err := parse(args[2].(string))
		if err != nil {
			return nil, err
		}
		node := cloneNode(original)
		node.alignX, node.alignY = ax, ay
		return node, nil
	}
	if name == "std.ui.textStyle" {
		original, ok := args[0].(*UiNode)
		if !ok || original == nil || (original.kind != "Text" && original.kind != "Button") {
			return nil, fmt.Errorf("UiTreeError: textStyle requires Text or Button")
		}
		size, weight, color := args[1].(int64), args[2].(int64), args[3].(int64)
		if size < 8 || size > 96 || weight < 100 || weight > 900 || weight%100 != 0 || !validCanvasColor(color) {
			return nil, fmt.Errorf("UiArgumentError: invalid text style")
		}
		node := cloneNode(original)
		node.fontSize, node.fontWeight, node.textColor, node.customTextStyle = size, weight, uint32(color), true
		return node, nil
	}
	if name == "std.ui.textWrap" {
		original, ok := args[0].(*UiNode)
		if !ok || original == nil || original.kind != "Text" {
			return nil, fmt.Errorf("UiTreeError: textWrap requires Text")
		}
		node := cloneNode(original)
		node.wrapText = true
		return node, nil
	}
	if name == "std.ui.focusPolicy" {
		original, ok := args[0].(*UiNode)
		if !ok || original == nil || (original.kind != "Button" && original.kind != "TextInput") {
			return nil, fmt.Errorf("UiTreeError: focusPolicy requires Button or TextInput")
		}
		node := cloneNode(original)
		node.tabStop, node.disabled = args[1].(bool), !args[2].(bool)
		return node, nil
	}
	if name == "std.ui.accessible" {
		original, ok := args[0].(*UiNode)
		if !ok || original == nil {
			return nil, fmt.Errorf("UiTreeError: accessible requires UiNode")
		}
		label, role := args[1].(string), args[2].(string)
		if len(label) > 512 || !utf8.ValidString(label) || strings.ContainsRune(label, 0) || (role != "text" && role != "button" && role != "textbox" && role != "group") {
			return nil, fmt.Errorf("UiArgumentError: invalid accessible name or role")
		}
		node := cloneNode(original)
		node.accessibleName, node.accessibleRole = label, role
		return node, nil
	}
	if name == "std.ui.surface" {
		original, ok := args[0].(*UiNode)
		if !ok || original == nil || (original.kind != "Panel" && original.kind != "Button") {
			return nil, fmt.Errorf("UiTreeError: surface requires Panel or Button")
		}
		fill, border, radius := args[1].(int64), args[2].(int64), args[3].(int64)
		if !validCanvasColor(fill) || !validCanvasColor(border) || fill>>24 != 255 || (border>>24 != 0 && border>>24 != 255) || radius < 0 || radius > 256 {
			return nil, fmt.Errorf("UiArgumentError: surface requires opaque fill, opaque or transparent border, radius 0..256")
		}
		node := cloneNode(original)
		node.surfaceFill, node.surfaceBorder, node.surfaceRadius, node.customSurface = uint32(fill), uint32(border), radius, true
		return node, nil
	}
	if name == "std.ui.spacing" {
		original, ok := args[0].(*UiNode)
		if !ok || original == nil || (original.kind != "Row" && original.kind != "Column" && original.kind != "Panel") {
			return nil, fmt.Errorf("UiTreeError: spacing requires Row, Column or Panel")
		}
		padding, gap := args[1].(int64), args[2].(int64)
		if padding < 0 || padding > 256 || gap < 0 || gap > 256 {
			return nil, fmt.Errorf("UiArgumentError: invalid spacing")
		}
		node := cloneNode(original)
		node.padding, node.gap, node.customSpacing = padding, gap, true
		return node, nil
	}
	if name == "std.ui.splitView" {
		first, firstOK := args[1].(*UiNode)
		second, secondOK := args[2].(*UiNode)
		if !firstOK || !secondOK || first == nil || second == nil {
			return nil, fmt.Errorf("UiTreeError: splitView requires two UiNodes")
		}
		ratio, minFirst, minSecond := args[3].(int64), args[4].(int64), args[5].(int64)
		if ratio < 0 || ratio > 100 || minFirst < 0 || minFirst > 8192 || minSecond < 0 || minSecond > 8192 {
			return nil, fmt.Errorf("UiArgumentError: invalid splitView constraints")
		}
		node, err := makeNode("SplitView", args[0].(string), "", []*UiNode{first, second})
		if err != nil {
			return nil, err
		}
		node.ratio, node.splitMinFirst, node.splitMinSecond = ratio, minFirst, minSecond
		return node, nil
	}
	if name == "std.ui.virtualList" {
		count, height, first := args[1].(int64), args[2].(int64), args[3].(int64)
		array := args[4].(*rt.Array)
		children := make([]*UiNode, 0, len(array.Items))
		for _, value := range array.Items {
			child, ok := value.(*UiNode)
			if !ok {
				return nil, fmt.Errorf("UiTreeError: VirtualList item is not UiNode")
			}
			children = append(children, child)
		}
		node, err := makeNode("VirtualList", args[0].(string), "", children)
		if err != nil {
			return nil, err
		}
		node.itemCount, node.rowHeight, node.firstIndex = count, height, first
		if err := validateTree(node); err != nil {
			return nil, err
		}
		return node, nil
	}
	if name == "std.ui.scrollView" {
		child, ok := args[1].(*UiNode)
		if !ok {
			return nil, fmt.Errorf("UiTreeError: ScrollView child is not UiNode")
		}
		return makeNode("ScrollView", args[0].(string), "", []*UiNode{child})
	}
	if name == "std.ui.grid" {
		minimum, maximum, gap := args[2].(int64), args[3].(int64), args[4].(int64)
		if minimum < 1 || minimum > 8192 || maximum < 1 || maximum > 64 || gap < 0 || gap > 256 {
			return nil, fmt.Errorf("UiArgumentError: invalid Grid constraints")
		}
		array := args[1].(*rt.Array)
		children := make([]*UiNode, 0, len(array.Items))
		for _, value := range array.Items {
			child, ok := value.(*UiNode)
			if !ok || child == nil {
				return nil, fmt.Errorf("UiTreeError: Grid item is not UiNode")
			}
			children = append(children, child)
		}
		node, err := makeNode("Grid", args[0].(string), "", children)
		if err != nil {
			return nil, err
		}
		node.gridMinCell, node.gridMaxColumns, node.gridGap = minimum, maximum, gap
		return node, nil
	}
	if name == "std.ui.textInput" {
		node, err := makeNode("TextInput", args[0].(string), args[1].(string), nil)
		if err != nil {
			return nil, err
		}
		placeholder := args[2].(string)
		if len(placeholder) > 4096 || !utf8.ValidString(placeholder) || strings.ContainsRune(placeholder, 0) {
			return nil, fmt.Errorf("UiArgumentError: invalid placeholder")
		}
		node.placeholder = placeholder
		return node, nil
	}
	if name == "std.ui.text" || name == "std.ui.button" {
		kind := "Text"
		if name == "std.ui.button" {
			kind = "Button"
		}
		return makeNode(kind, args[0].(string), args[1].(string), nil)
	}
	if name == "std.ui.row" || name == "std.ui.column" || name == "std.ui.panel" {
		kind := map[string]string{"std.ui.row": "Row", "std.ui.column": "Column", "std.ui.panel": "Panel"}[name]
		array := args[1].(*rt.Array)
		children := make([]*UiNode, 0, len(array.Items))
		for _, item := range array.Items {
			node, ok := item.(*UiNode)
			if !ok {
				return nil, fmt.Errorf("UiTreeError: child is not UiNode")
			}
			children = append(children, node)
		}
		return makeNode(kind, args[0].(string), "", children)
	}
	if name == "std.ui.openWindow" {
		if s.ui != nil {
			return nil, fmt.Errorf("UiSessionError: one window per run is supported")
		}
		title, width, height := args[0].(string), args[1].(int64), args[2].(int64)
		if len(title) > 1024 || width < 1 || width > 8192 || height < 1 || height > 8192 {
			return nil, fmt.Errorf("UiArgumentError: invalid title or window size")
		}
		ui, err := startUI(ctx)
		if err != nil {
			return nil, err
		}
		s.ui = ui
		answer, err := ui.send(ctx, uiCommand{op: "open", title: title, w: width, h: height})
		if err != nil {
			s.Close()
			return nil, err
		}
		return &UiWindow{owner: ui, id: answer.id}, nil
	}
	if s.ui == nil {
		return nil, fmt.Errorf("UiSessionError: open a window first")
	}
	if name == "std.ui.nextEvent" {
		if s.ui.closed.Load() {
			return nil, fmt.Errorf("UiSessionError: window is closed")
		}
		ms := args[0].(int64)
		if ms < -1 || ms > 60000 {
			return nil, fmt.Errorf("UiArgumentError: timeout must be -1..60000 milliseconds")
		}
		var timer <-chan time.Time
		if ms >= 0 {
			t := time.NewTimer(time.Duration(ms) * time.Millisecond)
			defer t.Stop()
			timer = t.C
		}
		select {
		case event := <-s.ui.events:
			s.ui.mu.Lock()
			s.ui.lastEvent = event.at
			s.ui.mu.Unlock()
			return rt.Some(event.value()), nil
		case <-timer:
			return rt.Optional{}, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-s.ui.done:
			return nil, s.ui.stopped()
		}
	}
	w, ok := args[0].(*UiWindow)
	if !ok || w.owner != s.ui || w.id != 1 {
		return nil, fmt.Errorf("UiInvalidHandle: window belongs to another session")
	}
	cmd := uiCommand{op: name}
	switch name {
	case "std.ui.setTitle":
		cmd.title = args[1].(string)
		if len(cmd.title) > 1024 {
			return nil, fmt.Errorf("UiArgumentError: title exceeds 1024 bytes")
		}
	case "std.ui.setBackground":
		cmd.r, cmd.g, cmd.b = args[1].(int64), args[2].(int64), args[3].(int64)
		if cmd.r < 0 || cmd.r > 255 || cmd.g < 0 || cmd.g > 255 || cmd.b < 0 || cmd.b > 255 {
			return nil, fmt.Errorf("UiArgumentError: RGB values must be 0..255")
		}
	case "std.ui.closeWindow":
	case "std.ui.renderer":
	case "std.ui.viewportWidth", "std.ui.viewportHeight":
	case "std.ui.measureTextWidth":
		cmd.value, cmd.w, cmd.h = args[1].(string), args[2].(int64), args[3].(int64)
		if len(cmd.value) > 4096 || !utf8.ValidString(cmd.value) || strings.ContainsRune(cmd.value, 0) || cmd.w < 8 || cmd.w > 96 || cmd.h < 100 || cmd.h > 900 || cmd.h%100 != 0 {
			return nil, fmt.Errorf("UiArgumentError: invalid text measurement arguments")
		}
	case "std.ui.setText":
		cmd.nodeID, cmd.value = args[1].(string), args[2].(string)
		if err := validateInputValue(cmd.value); err != nil {
			return nil, err
		}
	case "std.ui.focus":
		cmd.nodeID = args[1].(string)
	case "std.ui.pushFocusScope", "std.ui.popFocusScope":
		cmd.scopeID = args[1].(string)
	case "std.ui.bindScopedShortcut":
		cmd.scopeID, cmd.nodeID, cmd.key, cmd.mod = args[1].(string), args[2].(string), args[3].(int64), args[4].(int64)
		if err := validateShortcut(cmd.nodeID, cmd.key, cmd.mod); err != nil {
			return nil, err
		}
	case "std.ui.unbindScopedShortcut":
		cmd.scopeID, cmd.nodeID = args[1].(string), args[2].(string)
	case "std.ui.bindShortcut":
		cmd.nodeID, cmd.key, cmd.mod = args[1].(string), args[2].(int64), args[3].(int64)
		if err := validateShortcut(cmd.nodeID, cmd.key, cmd.mod); err != nil {
			return nil, err
		}
	case "std.ui.unbindShortcut":
		cmd.nodeID = args[1].(string)
	case "std.ui.mount", "std.ui.update":
		root, ok := args[1].(*UiNode)
		if !ok {
			return nil, fmt.Errorf("UiTreeError: root is not UiNode")
		}
		if err := validateTree(root); err != nil {
			return nil, err
		}
		cmd.root = cloneNode(root)
		if name == "std.ui.update" {
			s.ui.mu.Lock()
			cmd.origin = s.ui.lastEvent
			s.ui.lastEvent = time.Time{}
			s.ui.mu.Unlock()
		}
	default:
		return nil, fmt.Errorf("UiCapabilityError: unknown UI command")
	}
	reply, err := s.ui.send(ctx, cmd)
	if err == nil && name == "std.ui.closeWindow" {
		s.ui.closed.Store(true)
	}
	if name == "std.ui.renderer" {
		return reply.text, err
	}
	if name == "std.ui.viewportWidth" || name == "std.ui.viewportHeight" || name == "std.ui.measureTextWidth" {
		return reply.value, err
	}
	return nil, err
}

func validateShortcut(id string, key, mod int64) error {
	if id == "" || len(id) > 128 || !utf8.ValidString(id) || strings.ContainsRune(id, 0) || key < 1 || key > 255 || mod < 0 || mod > 7 || key == 9 || key == 16 || key == 17 || key == 18 {
		return fmt.Errorf("UiArgumentError: invalid shortcut")
	}
	return nil
}

func (u *uiSession) send(ctx context.Context, cmd uiCommand) (uiReply, error) {
	cmd.answer = make(chan uiReply, 1)
	cmd.queuedAt = time.Now()
	select {
	case u.commands <- cmd:
		u.wakeHost()
	case <-ctx.Done():
		return uiReply{}, ctx.Err()
	case <-u.done:
		return uiReply{}, u.stopped()
	}
	select {
	case reply := <-cmd.answer:
		return reply, reply.err
	case <-ctx.Done():
		return uiReply{}, ctx.Err()
	case <-u.done:
		return uiReply{}, u.stopped()
	}
}

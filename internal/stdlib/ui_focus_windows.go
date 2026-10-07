//go:build windows

package stdlib

import "fmt"

type focusScope struct{ id, restore string }

func (w *windowsState) activeFocusRoot() *UiNode {
	if len(w.focusScopes) == 0 {
		return w.root
	}
	return uiNodeByID(w.root, w.focusScopes[len(w.focusScopes)-1].id)
}

func (w *windowsState) allowedInScope(id string) bool {
	if len(w.focusScopes) == 0 {
		return true
	}
	return uiSubtreeHasID(w.activeFocusRoot(), id)
}

func (w *windowsState) focusCandidate(id string) error {
	if w.root == nil {
		return fmt.Errorf("UiFocusError: mount a tree first")
	}
	if id == "" {
		return fmt.Errorf("UiFocusError: empty target")
	}
	if !uiSubtreeHasID(w.root, id) {
		return fmt.Errorf("UiFocusError: %q is not in the mounted tree", id)
	}
	if !hasFocusable(w.root, id) {
		return fmt.Errorf("UiFocusError: %q is disabled or not focusable", id)
	}
	if !w.allowedInScope(id) {
		return fmt.Errorf("UiFocusError: %q is outside the active focus scope", id)
	}
	box := nodeBox(w.layout, id)
	if box != nil && (box.w <= 0 || box.h <= 0) {
		return fmt.Errorf("UiFocusError: %q has no visible layout box", id)
	}
	if box == nil && !canvasTargetInLayout(w.layout, id) {
		return fmt.Errorf("UiFocusError: %q is not materialized in the current layout", id)
	}
	return nil
}

func canvasTargetInLayout(box *uiBox, id string) bool {
	if box == nil {
		return false
	}
	for _, draw := range box.node.draws {
		if draw.kind == "hitRegion" && draw.target == id && box.w > 0 && box.h > 0 && draw.x < box.w && draw.y < box.h && draw.x+draw.w > 0 && draw.y+draw.h > 0 {
			return true
		}
	}
	for _, child := range box.children {
		if canvasTargetInLayout(child, id) {
			return true
		}
	}
	return false
}

func (w *windowsState) pushFocusScope(id string) error {
	root := uiNodeByID(w.root, id)
	if root == nil {
		return fmt.Errorf("UiFocusError: scope %q is not in the tree", id)
	}
	if len(w.focusScopes) > 0 && !uiSubtreeHasID(w.activeFocusRoot(), id) {
		return fmt.Errorf("UiFocusError: scope %q is outside the active scope", id)
	}
	first := firstFocusable(root)
	if first == "" {
		return fmt.Errorf("UiFocusError: scope %q has no enabled focus target", id)
	}
	w.focusScopes = append(w.focusScopes, focusScope{id: id, restore: w.focused})
	w.syncInputScope()
	if !uiSubtreeHasID(root, w.focused) || !hasFocusable(root, w.focused) {
		w.focusNode(first)
	}
	return nil
}

func (w *windowsState) popFocusScope(id string) error {
	if len(w.focusScopes) == 0 || w.focusScopes[len(w.focusScopes)-1].id != id {
		return fmt.Errorf("UiFocusError: scope %q is not active on top", id)
	}
	last := w.focusScopes[len(w.focusScopes)-1]
	w.focusScopes = w.focusScopes[:len(w.focusScopes)-1]
	w.syncInputScope()
	if last.restore != "" && w.focusCandidate(last.restore) == nil {
		w.focusNode(last.restore)
	} else {
		w.focusNode(firstFocusable(w.activeFocusRoot()))
	}
	return nil
}

func (w *windowsState) reconcileFocusScopes() {
	valid := 0
	parent := w.root
	for _, scope := range w.focusScopes {
		if !uiSubtreeHasID(parent, scope.id) {
			break
		}
		parent = uiNodeByID(parent, scope.id)
		valid++
	}
	w.focusScopes = w.focusScopes[:valid]
	for id := range w.scopedShortcuts {
		if uiNodeByID(w.root, id) == nil {
			delete(w.scopedShortcuts, id)
		}
	}
	w.syncInputScope()
}

func (w *windowsState) syncInputScope() {
	for id, input := range w.inputs {
		node := uiNodeByID(w.root, id)
		if node == nil || node.disabled || !w.allowedInScope(id) {
			uiTextEnableWindow.Call(input.hwnd, 0)
		} else {
			uiTextEnableWindow.Call(input.hwnd, 1)
		}
	}
}

func (w *windowsState) bindScopedShortcut(scope, id string, key shortcutKey) error {
	if uiNodeByID(w.root, scope) == nil {
		return fmt.Errorf("UiShortcutError: scope %q is not in the tree", scope)
	}
	if w.scopedShortcuts == nil {
		w.scopedShortcuts = make(map[string]map[shortcutKey]string)
	}
	bindings := w.scopedShortcuts[scope]
	if bindings == nil {
		bindings = make(map[shortcutKey]string)
		w.scopedShortcuts[scope] = bindings
	}
	if prior := bindings[key]; prior != "" && prior != id {
		return fmt.Errorf("UiShortcutError: shortcut already bound to %q in scope %q", prior, scope)
	}
	if len(bindings) >= 128 && bindings[key] == "" {
		return fmt.Errorf("UiShortcutError: limit of 128 shortcuts per scope exceeded")
	}
	bindings[key] = id
	return nil
}

func (w *windowsState) unbindScopedShortcut(scope, id string) {
	for key, command := range w.scopedShortcuts[scope] {
		if command == id {
			delete(w.scopedShortcuts[scope], key)
		}
	}
}

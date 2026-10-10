package stdlib

import "testing"

func TestBuildPathFiguresArcRing(t *testing.T) {
	ops := []uiPathOp{
		{kind: "std.ui.pathArcTo", v: [6]float64{100, 100, 60, 0, 360}},
		{kind: "std.ui.pathClose"},
		{kind: "std.ui.pathArcTo", v: [6]float64{100, 100, 30, 0, -360}},
		{kind: "std.ui.pathClose"},
	}
	figures, err := buildPathFigures(ops)
	if err != nil {
		t.Fatal(err)
	}
	if len(figures) != 2 || !figures[0].closed || !figures[1].closed {
		t.Fatalf("ring must have two closed subpaths: %#v", figures)
	}
	if len(figures[0].edges) != 4 || len(figures[1].edges) != 4 {
		t.Fatalf("full arcs must be split into four cubic curves")
	}
}

func TestBuildPathFiguresRejectsInvalidOrder(t *testing.T) {
	for _, ops := range [][]uiPathOp{
		{{kind: "std.ui.pathLineTo", v: [6]float64{1, 2}}},
		{{kind: "std.ui.pathClose"}},
		{{kind: "std.ui.pathMoveTo", v: [6]float64{1, 2}}},
	} {
		if _, err := buildPathFigures(ops); err == nil {
			t.Fatalf("accepted invalid path: %#v", ops)
		}
	}
}

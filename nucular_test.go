package nucular

import (
	"image"
	"testing"

	"github.com/aarzilli/nucular/label"
	"github.com/aarzilli/nucular/rect"

	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/mouse"
)

func centerOf(rect rect.Rect) image.Point {
	return image.Pt(rect.X+rect.W/2, rect.Y+rect.H/2)
}

func TestContextualReplace(t *testing.T) {
	test1cnt := 0
	test2cnt := 0
	test2clicked := 0
	var test1rect, test2rect, lblrect rect.Rect

	w := NewTestWindow(0, image.Pt(640, 480), func(w *Window) {
		w.Row(30).Static(180)
		w.Label("Right click me for menu", "LC")
		lblrect = w.LastWidgetBounds
		if w := w.ContextualOpen(0, image.Point{100, 300}, w.LastWidgetBounds, nil); w != nil {
			w.Row(25).Dynamic(1)
			if r := w.WidgetBounds(); test1cnt == 0 {
				test1rect = r
			} else if test1rect != r {
				t.Fatalf("test item 1 changed position (%d): %v -> %v", test1cnt, test1rect, r)
			}
			test1cnt++
			if w.MenuItem(label.TA("Test Item", "CC")) {
				w.ContextualOpen(WindowContextualReplace, image.Point{100, 300}, rect.Rect{0, 0, 0, 0}, func(w *Window) {
					w.Row(25).Dynamic(1)
					if r := w.WidgetBounds(); test2cnt == 0 {
						test2rect = r
					} else if test2rect != r {
						t.Fatalf("test item 2 changed position (%d): %v -> %v\n", test2cnt, test2rect, r)
					}
					test2cnt++
					if w.MenuItem(label.TA("Second Test Item", "CC")) {
						test2clicked++
					}
				})
			}
		}
	})

	w.Update()
	w.Click(mouse.ButtonRight, centerOf(lblrect))

	if test1cnt == 0 {
		t.Fatalf("Test item 1 was not displayed")
	}

	w.Click(mouse.ButtonLeft, centerOf(test1rect))

	if test2cnt == 0 {
		t.Fatalf("Test item 2 was not displayed")
	}

	if test1rect != test2rect {
		t.Fatalf("contextual replace failed: %v %v", test1rect, test2rect)
	}

	c := test2cnt
	w.Update()
	if test2cnt == c {
		t.Fatalf("second contextual menu closed immediately: %d", test2cnt)
	}
}

func TestWindowEnabledFlagOnGroup(t *testing.T) {
	clicked := 0
	var buttonrect rect.Rect
	w := NewTestWindow(0, image.Pt(640, 480), func(w *Window) {
		w.Row(0).Dynamic(1)
		if w := w.GroupBegin("subwindow", 0); w != nil {
			w.Row(20).Static(100)
			if w.ButtonText("Test button") {
				clicked++
			}
			buttonrect = w.LastWidgetBounds
			w.GroupEnd()
		}
	})

	w.Update()
	w.Click(mouse.ButtonLeft, centerOf(buttonrect))
	if clicked != 1 {
		t.Fatalf("button wasn't clicked")
	}
}

func TestEditorEndKey(t *testing.T) {
	const testString = "this is a test string"
	const testString2 = testString + "\n" + testString
	var ed TextEditor
	ed.Flags = EditSelectable | EditMultiline
	ed.Active = true
	ed.Buffer = []rune(testString)
	w := NewTestWindow(0, image.Pt(640, 480), func(w *Window) {
		w.Row(0).Dynamic(1)
		ed.Edit(w)
	})

	check := func(tgt int, title string) {
		if ed.Cursor != tgt {
			t.Fatalf("Cursor position %d, expected %d (%s)", ed.Cursor, tgt, title)
		}
	}

	w.Update()
	w.TypeKey(key.Event{Code: key.CodeEnd})
	check(len(testString), "singleline")

	w.TypeKey(key.Event{Code: key.CodeEnd})
	check(len(testString), "singleline 2")

	ed.Cursor = 0
	ed.Buffer = []rune(testString2)
	w.TypeKey(key.Event{Code: key.CodeEnd})
	check(len(testString), "multiline")

	w.TypeKey(key.Event{Code: key.CodeEnd})
	check(len(testString), "multiline 2")
}

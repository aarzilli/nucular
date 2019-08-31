package main

import (
	"github.com/aarzilli/nucular"
	"github.com/aarzilli/nucular/rect"
	"github.com/aarzilli/nucular/style"
	"golang.org/x/mobile/event/key"
)

var (
	outerWindowFlags         = nucular.WindowNoScrollbar // also try nucular.WindowNoScrollbar
	forceVerticalScrollbar   = false
	forceHorizontalScrollbar = 0
	beginningOfRow           = true
	forceMenuBar             = false
	alternateView            = false
)

func main() {
	wnd := nucular.NewMasterWindow(outerWindowFlags, "Counter", updatefn)
	wnd.SetStyle(style.FromTheme(style.DarkTheme, 2.0))
	wnd.Main()
}

func updatefn(w *nucular.Window) {
	for _, e := range w.Master().Input().Keyboard.Keys {
		switch e.Code {
		case key.CodeB:
			beginningOfRow = !beginningOfRow
		case key.CodeH:
			if e.Modifiers&key.ModShift != 0 {
				forceHorizontalScrollbar--
			} else {
				forceHorizontalScrollbar++
			}
			if forceHorizontalScrollbar < 0 {
				forceHorizontalScrollbar = 0
			}
			if forceHorizontalScrollbar > 15 {
				forceHorizontalScrollbar = 15
			}
		case key.CodeV:
			forceVerticalScrollbar = !forceVerticalScrollbar
		case key.CodeM:
			forceMenuBar = !forceMenuBar
		case key.CodeTab:
			alternateView = !alternateView
		case key.CodeP:
			w.Master().PopupOpen("blah", nucular.WindowMovable|nucular.WindowTitle|nucular.WindowClosable|nucular.WindowScalable, rect.Rect{20, 100, 230, 150}, true, func(w *nucular.Window) {
				if forceHorizontalScrollbar > 0 {
					w.Row(20).Static(1000)
					w.Label("Floating window", "LC")
				}
				w.Row(0).Dynamic(1)
				if w := w.GroupBegin("float", nucular.WindowBorder); w != nil {
					w.GroupEnd()
				}
			})
		}
	}
	if !alternateView {
		if forceMenuBar {
			w.MenubarBegin()
			w.Row(20).Static(200)
			w.Label("Menubar contents", "LC")
			w.MenubarEnd()
		}
		for i := 0; i < forceHorizontalScrollbar; i++ {
			w.Row(20).Static(1000)
			w.Label("Force horizontal scrollbar", "LC")
		}
		if !beginningOfRow {
			w.Row(20).Static(100, 10)
			w.Label("nope", "LC")
		}
		w.Row(0).Dynamic(1)
		if w := w.GroupBegin("", nucular.WindowBorder); w != nil {
			w.GroupEnd()
		}
		if forceVerticalScrollbar {
			w.Row(20).Static(100)
			w.Label("Force vertical scrollbar", "LC")
		}
	} else {
		w.Row(0).Dynamic(2)
		if w := w.GroupBegin("1", nucular.WindowBorder); w != nil {
			w.Row(50).Dynamic(1)
			if w := w.GroupBegin("1.1", nucular.WindowBorder); w != nil {
				w.Row(20).Dynamic(1)
				w.Label("1.1", "LC")
				w.GroupEnd()
			}
			w.Row(0).Dynamic(1)
			if w := w.GroupBegin("1.2", nucular.WindowBorder); w != nil {
				w.Row(20).Dynamic(1)
				w.Label("1.2", "LC")
				w.Row(0).Dynamic(2)
				if w := w.GroupBegin("1.2.1", nucular.WindowBorder); w != nil {
					w.Row(20).Dynamic(1)
					w.Label("1.2.1", "LC")
					w.GroupEnd()
				}
				if w := w.GroupBegin("1.2.2", nucular.WindowBorder); w != nil {
					w.Row(20).Dynamic(1)
					w.Label("1.2.2", "LC")
					w.GroupEnd()
				}
				w.GroupEnd()
			}
			w.GroupEnd()
		}
		if w := w.GroupBegin("2", nucular.WindowBorder); w != nil {
			w.Row(50).Dynamic(1)
			if w := w.GroupBegin("2.1", nucular.WindowBorder); w != nil {
				w.Row(20).Dynamic(1)
				w.Label("2.1", "LC")
				w.GroupEnd()
			}
			w.Row(0).Dynamic(1)
			if w := w.GroupBegin("2.2", nucular.WindowBorder); w != nil {
				w.Row(20).Dynamic(1)
				w.Label("2.1", "LC")
				w.GroupEnd()
			}
			w.GroupEnd()
		}
	}
}

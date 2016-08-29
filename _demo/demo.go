package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/aarzilli/nucular"
	"github.com/aarzilli/nucular/label"
	nstyle "github.com/aarzilli/nucular/style"
)

var whichdemo int = 4

const scaling = 1.8

//var theme nucular.Theme = nucular.WhiteTheme
var theme nstyle.Theme = nstyle.DarkTheme

func main() {
	var wnd *nucular.MasterWindow

	switch whichdemo {
	case 0:
		wnd = nucular.NewMasterWindow(buttonDemo, 0)
	case 1:
		wnd = nucular.NewMasterWindow(basicDemo, 0)
		go func() {
			for {
				time.Sleep(1 * time.Second)
				if wnd.Closed() {
					break
				}
				wnd.Changed()
			}
		}()
	case 2:
		textEditorEditor.Flags = nucular.EditSelectable
		textEditorEditor.Buffer = []rune("prova")
		wnd = nucular.NewMasterWindow(textEditorDemo, 0)
	case 3:
		var cd calcDemo
		cd.current = &cd.a
		wnd = nucular.NewMasterWindow(cd.calculatorDemo, 0)
	case 4:
		od := newOverviewDemo()
		od.Theme = theme
		wnd = nucular.NewMasterWindow(od.overviewDemo, 0)
	case 5:
		wnd = nucular.NewMasterWindow(horizontalSplit, nucular.WindowNoScrollbar)
	case 6:
		wnd = nucular.NewMasterWindow(widgetBoundsBug, nucular.WindowNoScrollbar)
	case 7:
		bs, _ := ioutil.ReadFile("overview.go")
		multilineTextEditor.Buffer = []rune(string(bs))
		wnd = nucular.NewMasterWindow(multilineTextEditorDemo, nucular.WindowNoScrollbar)
	}
	wnd.SetStyle(nstyle.FromTheme(theme), nil, scaling)
	wnd.Main()
}

func buttonDemo(mw *nucular.MasterWindow, w *nucular.Window) {
	w.Row(20).Static(60, 60)
	if w.Button(label.T("button1"), false) {
		fmt.Printf("button pressed!\n")
	}
	if w.Button(label.T("button2"), false) {
		fmt.Printf("button 2 pressed!\n")
	}
}

type difficulty int

const (
	easy = difficulty(iota)
	hard
)

var op difficulty = easy
var compression int

func basicDemo(mw *nucular.MasterWindow, w *nucular.Window) {
	w.Row(30).Dynamic(1)
	w.Label(time.Now().Format("15:04:05"), "RT")

	w.Row(30).Static(80)
	if w.Button(label.T("button"), false) {
		fmt.Printf("button pressed! difficulty: %v compression: %d\n", op, compression)
	}
	w.Row(30).Dynamic(2)
	if w.OptionText("easy", op == easy) {
		op = easy
	}
	if w.OptionText("hard", op == hard) {
		op = hard
	}
	w.Row(25).Dynamic(1)
	w.PropertyInt("Compression:", 0, &compression, 100, 10, 1)
}

var textEditorEditor nucular.TextEditor

func textEditorDemo(mw *nucular.MasterWindow, w *nucular.Window) {
	w.Row(30).Dynamic(1)
	textEditorEditor.Maxlen = 30
	textEditorEditor.Edit(w)
}

func horizontalSplit(mw *nucular.MasterWindow, w *nucular.Window) {
	h := w.LayoutAvailableHeight()
	w.RowScaled(h).Dynamic(2)
	if sw := w.GroupBegin("Left", nucular.WindowNoHScrollbar|nucular.WindowBorder); sw != nil {
		sw.Row(18).Static(150)
		for i := 0; i < 64; i++ {
			sw.Label(fmt.Sprintf("%#02x", i), "LC")
		}
		sw.GroupEnd()
	}
	if sw := w.GroupBegin("Right", nucular.WindowNoHScrollbar|nucular.WindowBorder); sw != nil {
		sw.Row(18).Static(150)
		for i := 0; i < 64; i++ {
			sw.Label(fmt.Sprintf("%#03o", i), "LC")
		}
		sw.GroupEnd()
	}
}

func widgetBoundsBug(mw *nucular.MasterWindow, w *nucular.Window) {
	w.Row(20).StaticScaled(200, 250)
	w.Label("first", "LC")
	w.Label("second", "LC")
	bounds := w.WidgetBounds()
	w.Label("third", "LC")
	bounds2 := w.LastWidgetBounds
	if bounds != bounds2 {
		fmt.Printf("mismatched: %#v %#v\n", bounds, bounds2)
	}
}

var multilineTextEditor nucular.TextEditor

func multilineTextEditorDemo(mw *nucular.MasterWindow, w *nucular.Window) {
	w.Row(0).Dynamic(1)
	multilineTextEditor.Flags = nucular.EditMultiline | nucular.EditSelectable | nucular.EditClipboard
	multilineTextEditor.Edit(w)
}

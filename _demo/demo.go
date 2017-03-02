package main

import (
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"os"
	"runtime/pprof"
	"runtime/trace"
	"time"

	"github.com/aarzilli/nucular"
	"github.com/aarzilli/nucular/label"
	"github.com/aarzilli/nucular/rect"
	nstyle "github.com/aarzilli/nucular/style"
)

var whichdemo int = 4

const dotrace = false
const scaling = 1.8

//var theme nucular.Theme = nucular.WhiteTheme
var theme nstyle.Theme = nstyle.DarkTheme

func main() {
	var wnd nucular.MasterWindow

	if dotrace {
		fh, _ := os.Create("demo.trace.out")
		if fh != nil {
			defer fh.Close()
			trace.Start(fh)
			defer trace.Stop()
		}
		f, _ := os.Create("demo.cpu.pprof")
		if f != nil {
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}
	}

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
	case 8:
		pd := &panelDebug{}
		pd.Init()
		wnd = nucular.NewMasterWindow(pd.Update, nucular.WindowNoScrollbar)
	case 9:
		wnd = nucular.NewMasterWindow(nestedMenu, nucular.WindowNoScrollbar)
	}
	wnd.SetStyle(nstyle.FromTheme(theme, scaling))
	wnd.Main()
	if dotrace {
		fh, _ := os.Create("demo.heap.pprof")
		if fh != nil {
			defer fh.Close()
			pprof.WriteHeapProfile(fh)
		}
	}
}

func buttonDemo(w *nucular.Window) {
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

func basicDemo(w *nucular.Window) {
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

func textEditorDemo(w *nucular.Window) {
	w.Row(30).Dynamic(1)
	textEditorEditor.Maxlen = 30
	textEditorEditor.Edit(w)
}

func horizontalSplit(w *nucular.Window) {
	w.Row(0).Dynamic(2)
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

func widgetBoundsBug(w *nucular.Window) {
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

func multilineTextEditorDemo(w *nucular.Window) {
	w.Row(0).Dynamic(1)
	multilineTextEditor.Flags = nucular.EditMultiline | nucular.EditSelectable | nucular.EditClipboard
	multilineTextEditor.Edit(w)
}

type panelDebug struct {
	splitv          nucular.ScalableSplit
	splith          nucular.ScalableSplit
	showblocks      bool
	showsingleblock bool
	showtabs        bool
}

func (pd *panelDebug) Init() {
	pd.splitv.MinSize = 80
	pd.splitv.Size = 120
	pd.splitv.Spacing = 5
	pd.splith.MinSize = 100
	pd.splith.Size = 300
	pd.splith.Spacing = 5
	pd.showtabs = true
	pd.showsingleblock = true
}

func (pd *panelDebug) Update(w *nucular.Window) {
	for _, k := range w.Input().Keyboard.Keys {
		if k.Rune == 'b' {
			pd.showsingleblock = false
			pd.showblocks = !pd.showblocks
		}
		if k.Rune == 'B' {
			pd.showsingleblock = !pd.showsingleblock
		}
		if k.Rune == 't' {
			pd.showtabs = !pd.showtabs
		}
	}

	if pd.showtabs {
		w.Row(20).Dynamic(2)
		w.Label("A", "LC")
		w.Label("B", "LC")
	}

	area := w.Row(0).SpaceBegin(0)

	if pd.showsingleblock {
		w.LayoutSpacePushScaled(area)
		bounds, out := w.Custom(nstyle.WidgetStateInactive)
		if out != nil {
			out.FillRect(bounds, 10, color.RGBA{0x00, 0x00, 0xff, 0xff})
		}
	} else {
		leftbounds, rightbounds := pd.splitv.Vertical(w, area)
		viewbounds, commitbounds := pd.splith.Horizontal(w, rightbounds)

		w.LayoutSpacePushScaled(leftbounds)
		pd.groupOrBlock(w, "index-files", nucular.WindowBorder)

		w.LayoutSpacePushScaled(viewbounds)
		pd.groupOrBlock(w, "index-diff", nucular.WindowBorder)

		w.LayoutSpacePushScaled(commitbounds)
		pd.groupOrBlock(w, "index-right-column", nucular.WindowNoScrollbar|nucular.WindowBorder)
	}
}

func (pd *panelDebug) groupOrBlock(w *nucular.Window, name string, flags nucular.WindowFlags) {
	if pd.showblocks {
		bounds, out := w.Custom(nstyle.WidgetStateInactive)
		if out != nil {
			out.FillRect(bounds, 10, color.RGBA{0x00, 0x00, 0xff, 0xff})
		}
	} else {
		if sw := w.GroupBegin(name, flags); sw != nil {
			sw.GroupEnd()
		}
	}
}

func nestedMenu(w *nucular.Window) {
	w.Row(20).Static(180)
	w.Label("Test", "CC")
	w.ContextualOpen(0, image.Point{0, 0}, w.LastWidgetBounds, func(w *nucular.Window) {
		w.Row(20).Dynamic(1)
		if w.MenuItem(label.TA("Submenu", "CC")) {
			w.ContextualOpen(0, image.Point{0, 0}, rect.Rect{0, 0, 0, 0}, func(w *nucular.Window) {
				w.Row(20).Dynamic(1)
				if w.MenuItem(label.TA("Done", "CC")) {
					fmt.Printf("done\n")
				}
			})
		}
	})
}

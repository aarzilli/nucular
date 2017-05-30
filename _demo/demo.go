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

	"golang.org/x/mobile/event/key"
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
		wnd = nucular.NewMasterWindow(0, "Button Demo", buttonDemo)
	case 1:
		wnd = nucular.NewMasterWindow(0, "Basic Demo", basicDemo)
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
		wnd = nucular.NewMasterWindow(0, "Text Editor Demo", textEditorDemo)
	case 3:
		var cd calcDemo
		cd.current = &cd.a
		wnd = nucular.NewMasterWindow(0, "Calculator Demo", cd.calculatorDemo)
	case 4:
		od := newOverviewDemo()
		od.Theme = theme
		wnd = nucular.NewMasterWindow(0, "Overview", od.overviewDemo)
	case 7:
		bs, _ := ioutil.ReadFile("overview.go")
		multilineTextEditor.Buffer = []rune(string(bs))
		wnd = nucular.NewMasterWindow(nucular.WindowNoScrollbar, "Multiline Text Editor", multilineTextEditorDemo)
	case 8:
		pd := &panelDebug{}
		pd.Init()
		wnd = nucular.NewMasterWindow(nucular.WindowNoScrollbar, "Split panel demo", pd.Update)
	case 9:
		wnd = nucular.NewMasterWindow(nucular.WindowNoScrollbar, "Nested menu demo", nestedMenu)
	case 10:
		wnd = nucular.NewMasterWindow(nucular.WindowNoScrollbar, "List", listDemo)
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

var listDemoSelected = -1

func listDemo(w *nucular.Window) {
	const N = 100
	recenter := false
	for _, e := range w.Input().Keyboard.Keys {
		switch e.Code {
		case key.CodeDownArrow:
			listDemoSelected++
			if listDemoSelected >= N {
				listDemoSelected = N - 1
			}
			recenter = true
		case key.CodeUpArrow:
			listDemoSelected--
			if listDemoSelected < -1 {
				listDemoSelected = -1
			}
			recenter = true
		}
	}
	w.Row(0).Dynamic(1)
	if gl, w := nucular.GroupListStart(w, N, "list", nucular.WindowNoHScrollbar); w != nil {
		w.Row(20).Dynamic(1)
		for gl.Next() {
			i := gl.Index()
			selected := i == listDemoSelected
			w.SelectableLabel(fmt.Sprintf("label %d", i), "LC", &selected)
			if selected {
				listDemoSelected = i
				if recenter {
					gl.Center()
				}
			}
		}
	}
}

package main

import (
	"os"
	"runtime/pprof"
	"runtime/trace"

	"github.com/aarzilli/nucular"
	nstyle "github.com/aarzilli/nucular/style"
)

const dotrace = false
const scaling = 1.8

var Wnd nucular.MasterWindow

//var theme nucular.Theme = nucular.WhiteTheme
var theme nstyle.Theme = nstyle.DarkTheme

func main() {
	if dotrace {
		fh, _ := os.Create("overview.trace.out")
		if fh != nil {
			defer fh.Close()
			trace.Start(fh)
			defer trace.Stop()
		}
		f, _ := os.Create("overview.cpu.pprof")
		if f != nil {
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}
	}

	od := newOverviewDemo()
	od.Theme = theme

	Wnd = nucular.NewMasterWindow(0, "Overview", od.overviewDemo)
	Wnd.SetStyle(nstyle.FromTheme(theme, scaling))
	Wnd.Main()
	if dotrace {
		fh, _ := os.Create("demo.heap.pprof")
		if fh != nil {
			defer fh.Close()
			pprof.WriteHeapProfile(fh)
		}
	}
}

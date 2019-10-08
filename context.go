package nucular

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"math"
	"time"

	"github.com/aarzilli/nucular/command"
	"github.com/aarzilli/nucular/rect"
	nstyle "github.com/aarzilli/nucular/style"

	"github.com/golang/freetype/raster"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/mouse"
)

const perfUpdate = false
const dumpFrame = false

var UnknownCommandErr = errors.New("unknown command")

type context struct {
	mw             MasterWindow
	Input          Input
	Style          nstyle.Style
	Windows        []*Window
	DockedWindows  dockedTree
	changed        int32
	activateEditor *TextEditor
	cmds           []command.Command
	trashFrame     bool
	autopos        image.Point

	finalCmds command.Buffer

	dockedWindowFocus int
	floatWindowFocus  int
	scrollwheelFocus  int
	dockedCnt         int

	cmdstim []time.Duration // contains timing for all commands
}

func contextAllCommands(ctx *context) {
	ctx.cmds = ctx.cmds[:0]
	for i, w := range ctx.Windows {
		ctx.cmds = append(ctx.cmds, w.cmds.Commands...)
		if i == 0 {
			ctx.DockedWindows.Walk(func(w *Window) *Window {
				ctx.cmds = append(ctx.cmds, w.cmds.Commands...)
				return w
			})
		}
	}
	ctx.cmds = append(ctx.cmds, ctx.finalCmds.Commands...)
}

func (ctx *context) setupMasterWindow(layout *panel, updatefn UpdateFn) {
	ctx.Windows = append(ctx.Windows, createWindow(ctx, ""))
	ctx.Windows[0].idx = 0
	ctx.Windows[0].layout = layout
	ctx.Windows[0].flags = layout.Flags | WindowNonmodal
	ctx.Windows[0].updateFn = updatefn
}

func (ctx *context) Update() {
	for count := 0; count < 2; count++ {
		contextBegin(ctx, ctx.Windows[0].layout)
		for i := 0; i < len(ctx.Windows); i++ {
			ctx.Windows[i].began = false
		}
		ctx.Restack()
		ctx.FindFocus()
		for i := 0; i < len(ctx.Windows); i++ { // this must not use range or tooltips won't work
			ctx.updateWindow(ctx.Windows[i])
			if i == 0 {
				t := ctx.DockedWindows.Update(ctx.Windows[0].Bounds, ctx.Style.Scaling)
				if t != nil {
					ctx.DockedWindows = *t
				}
			}
		}
		contextEnd(ctx)
		if !ctx.trashFrame {
			break
		} else {
			ctx.Reset()
		}
	}
}

func (ctx *context) updateWindow(win *Window) {
	if win.updateFn != nil {
		win.specialPanelBegin()
		win.updateFn(win)
	}

	if !win.began {
		win.close = true
		return
	}

	if win.title == tooltipWindowTitle {
		win.close = true
	}

	if win.flags&windowPopup != 0 {
		panelEnd(ctx, win)
	}
}

func (ctx *context) processKeyEvent(e key.Event, textbuffer *bytes.Buffer) {
	if e.Direction == key.DirRelease {
		return
	}

	evinNotext := func() {
		for _, k := range ctx.Input.Keyboard.Keys {
			if k.Code == e.Code {
				k.Modifiers |= e.Modifiers
				return
			}
		}
		ctx.Input.Keyboard.Keys = append(ctx.Input.Keyboard.Keys, e)
	}
	evinText := func() {
		if e.Modifiers == 0 || e.Modifiers == key.ModShift {
			io.WriteString(textbuffer, string(e.Rune))
		}

		evinNotext()
	}

	switch {
	case e.Code == key.CodeUnknown:
		if e.Rune > 0 {
			evinText()
		}
	case (e.Code >= key.CodeA && e.Code <= key.Code0) || e.Code == key.CodeSpacebar || e.Code == key.CodeHyphenMinus || e.Code == key.CodeEqualSign || e.Code == key.CodeLeftSquareBracket || e.Code == key.CodeRightSquareBracket || e.Code == key.CodeBackslash || e.Code == key.CodeSemicolon || e.Code == key.CodeApostrophe || e.Code == key.CodeGraveAccent || e.Code == key.CodeComma || e.Code == key.CodeFullStop || e.Code == key.CodeSlash || (e.Code >= key.CodeKeypadSlash && e.Code <= key.CodeKeypadPlusSign) || (e.Code >= key.CodeKeypad1 && e.Code <= key.CodeKeypadEqualSign):
		evinText()
	case e.Code == key.CodeTab:
		e.Rune = '\t'
		evinText()
	case e.Code == key.CodeReturnEnter || e.Code == key.CodeKeypadEnter:
		e.Rune = '\n'
		evinText()
	default:
		evinNotext()
	}
}

func contextBegin(ctx *context, layout *panel) {
	for _, w := range ctx.Windows {
		w.usingSub = false
		w.curNode = w.rootNode
		w.close = false
		w.widgets.reset()
		w.cmds.Reset()
	}
	ctx.finalCmds.Reset()
	ctx.DockedWindows.Walk(func(w *Window) *Window {
		w.usingSub = false
		w.curNode = w.rootNode
		w.close = false
		w.widgets.reset()
		w.cmds.Reset()
		return w
	})

	ctx.trashFrame = false
	ctx.Windows[0].layout = layout
	panelBegin(ctx, ctx.Windows[0], "")
	layout.Offset = &ctx.Windows[0].Scrollbar
}

func contextEnd(ctx *context) {
	panelEnd(ctx, ctx.Windows[0])
}

func (ctx *context) Reset() {
	prevNumWindows := len(ctx.Windows)
	for i := 0; i < len(ctx.Windows); i++ {
		if ctx.Windows[i].close {
			if i != len(ctx.Windows)-1 {
				copy(ctx.Windows[i:], ctx.Windows[i+1:])
				i--
			}
			ctx.Windows = ctx.Windows[:len(ctx.Windows)-1]
		}
	}
	for i := range ctx.Windows {
		ctx.Windows[i].idx = i
	}
	if prevNumWindows == 2 && len(ctx.Windows) == 1 && ctx.Input.Mouse.valid {
		ctx.DockedWindows.Walk(func(w *Window) *Window {
			if w.flags&windowDocked == 0 {
				return w
			}
			for _, b := range []mouse.Button{mouse.ButtonLeft, mouse.ButtonRight, mouse.ButtonMiddle} {
				btn := ctx.Input.Mouse.Buttons[b]
				if btn.Clicked && w.Bounds.Contains(btn.ClickedPos) {
					ctx.dockedWindowFocus = w.idx
					return w
				}
			}
			return w
		})
	}
	ctx.activateEditor = nil
	in := &ctx.Input
	in.Mouse.Buttons[mouse.ButtonLeft].Clicked = false
	in.Mouse.Buttons[mouse.ButtonMiddle].Clicked = false
	in.Mouse.Buttons[mouse.ButtonRight].Clicked = false
	in.Mouse.ScrollDelta = 0
	in.Mouse.Prev.X = in.Mouse.Pos.X
	in.Mouse.Prev.Y = in.Mouse.Pos.Y
	in.Mouse.Delta = image.Point{}
	in.Keyboard.Keys = in.Keyboard.Keys[0:0]
}

func (ctx *context) Restack() {
	clicked := false
	for _, b := range []mouse.Button{mouse.ButtonLeft, mouse.ButtonRight, mouse.ButtonMiddle} {
		if ctx.Input.Mouse.Buttons[b].Clicked && ctx.Input.Mouse.Buttons[b].Down {
			clicked = true
			break
		}
	}
	if !clicked {
		return
	}
	ctx.dockedWindowFocus = 0
	nonmodalToplevel := false
	var toplevelIdx int
	for i := len(ctx.Windows) - 1; i >= 0; i-- {
		if ctx.Windows[i].flags&windowTooltip == 0 {
			toplevelIdx = i
			nonmodalToplevel = ctx.Windows[i].flags&WindowNonmodal != 0
			break
		}
	}
	if !nonmodalToplevel {
		return
	}
	// toplevel window is non-modal, proceed to change the stacking order if
	// the user clicked outside of it
	restacked := false
	found := false
	for i := len(ctx.Windows) - 1; i > 0; i-- {
		if ctx.Windows[i].flags&windowTooltip != 0 {
			continue
		}
		if ctx.restackClick(ctx.Windows[i]) {
			found = true
			if toplevelIdx != i {
				newToplevel := ctx.Windows[i]
				copy(ctx.Windows[i:toplevelIdx], ctx.Windows[i+1:toplevelIdx+1])
				ctx.Windows[toplevelIdx] = newToplevel
				restacked = true
			}
			break
		}
	}
	if restacked {
		for i := range ctx.Windows {
			ctx.Windows[i].idx = i
		}
	}
	if found {
		return
	}
	ctx.DockedWindows.Walk(func(w *Window) *Window {
		if ctx.restackClick(w) && (w.flags&windowDocked != 0) {
			ctx.dockedWindowFocus = w.idx
		}
		return w
	})
}

func (ctx *context) FindFocus() {
	ctx.floatWindowFocus = 0
	for i := len(ctx.Windows) - 1; i >= 0; i-- {
		if ctx.Windows[i].flags&windowTooltip == 0 {
			ctx.floatWindowFocus = i
			break
		}
	}
	ctx.scrollwheelFocus = 0
	for i := len(ctx.Windows) - 1; i > 0; i-- {
		if ctx.Windows[i].Bounds.Contains(ctx.Input.Mouse.Pos) {
			ctx.scrollwheelFocus = i
			break
		}
	}
	if ctx.scrollwheelFocus == 0 {
		ctx.DockedWindows.Walk(func(w *Window) *Window {
			if w.Bounds.Contains(ctx.Input.Mouse.Pos) {
				ctx.scrollwheelFocus = w.idx
			}
			return w
		})
	}
}

func (ctx *context) Walk(fn WindowWalkFn) {
	fn(ctx.Windows[0].title, ctx.Windows[0].Data, false, 0, ctx.Windows[0].Bounds)
	ctx.DockedWindows.walkExt(func(t *dockedTree) {
		switch t.Type {
		case dockedNodeHoriz:
			fn("", nil, true, t.Split.Size, rect.Rect{})
		case dockedNodeVert:
			fn("", nil, true, -t.Split.Size, rect.Rect{})
		case dockedNodeLeaf:
			if t.W == nil {
				fn("", nil, true, 0, rect.Rect{})
			} else {
				fn(t.W.title, t.W.Data, true, 0, t.W.Bounds)
			}
		}
	})
	for _, win := range ctx.Windows[1:] {
		if win.flags&WindowNonmodal != 0 {
			fn(win.title, win.Data, false, 0, win.Bounds)
		}
	}
}

func (ctx *context) restackClick(w *Window) bool {
	if !ctx.Input.Mouse.valid {
		return false
	}
	for _, b := range []mouse.Button{mouse.ButtonLeft, mouse.ButtonRight, mouse.ButtonMiddle} {
		btn := ctx.Input.Mouse.Buttons[b]
		if btn.Clicked && btn.Down && w.Bounds.Contains(btn.ClickedPos) {
			return true
		}
	}
	return false
}

var cnt = 0
var ln, frect, frectover, brrect, frrect, ftri, circ, fcirc, txt int

func (ctx *context) Draw(wimg *image.RGBA) int {
	var txttim, tritim, brecttim, frecttim, frectovertim, frrecttim time.Duration
	var t0 time.Time

	img := wimg

	var painter *myRGBAPainter
	var rasterizer *raster.Rasterizer

	roundAngle := func(cx, cy int, radius uint16, startAngle, angle float64, c color.Color) {
		rasterizer.Clear()
		rasterizer.Start(fixed.P(cx, cy))
		traceArc(rasterizer, float64(cx), float64(cy), float64(radius), float64(radius), startAngle, angle, false)
		rasterizer.Add1(fixed.P(cx, cy))
		painter.SetColor(c)
		rasterizer.Rasterize(painter)

	}

	setupRasterizer := func() {
		rasterizer = raster.NewRasterizer(img.Bounds().Dx(), img.Bounds().Dy())
		painter = &myRGBAPainter{Image: img}
	}

	if ctx.cmdstim != nil {
		ctx.cmdstim = ctx.cmdstim[:0]
	}

	transparentBorderOptimization := false

	for i := range ctx.cmds {
		if perfUpdate {
			t0 = time.Now()
		}
		icmd := &ctx.cmds[i]
		switch icmd.Kind {
		case command.ScissorCmd:
			img = wimg.SubImage(icmd.Rectangle()).(*image.RGBA)
			painter = nil
			rasterizer = nil

		case command.LineCmd:
			cmd := icmd.Line
			colimg := image.NewUniform(cmd.Color)
			op := draw.Over
			if cmd.Color.A == 0xff {
				op = draw.Src
			}

			h1 := int(cmd.LineThickness / 2)
			h2 := int(cmd.LineThickness) - h1

			if cmd.Begin.X == cmd.End.X {
				// draw vertical line
				r := image.Rect(cmd.Begin.X-h1, cmd.Begin.Y, cmd.Begin.X+h2, cmd.End.Y)
				drawFill(img, r, colimg, r.Min, op)
			} else if cmd.Begin.Y == cmd.End.Y {
				// draw horizontal line
				r := image.Rect(cmd.Begin.X, cmd.Begin.Y-h1, cmd.End.X, cmd.Begin.Y+h2)
				drawFill(img, r, colimg, r.Min, op)
			} else {
				if rasterizer == nil {
					setupRasterizer()
				}

				unzw := rasterizer.UseNonZeroWinding
				rasterizer.UseNonZeroWinding = true

				var p raster.Path
				p.Start(fixed.P(cmd.Begin.X-img.Bounds().Min.X, cmd.Begin.Y-img.Bounds().Min.Y))
				p.Add1(fixed.P(cmd.End.X-img.Bounds().Min.X, cmd.End.Y-img.Bounds().Min.Y))

				rasterizer.Clear()
				rasterizer.AddStroke(p, fixed.I(int(cmd.LineThickness)), nil, nil)
				painter.SetColor(cmd.Color)
				rasterizer.Rasterize(painter)

				rasterizer.UseNonZeroWinding = unzw
			}
			ln++

		case command.RectFilledCmd:
			cmd := icmd.RectFilled
			if i == 0 {
				// first command draws the background, insure that it's always fully opaque
				cmd.Color.A = 0xff
			}
			if transparentBorderOptimization {
				transparentBorderOptimization = false
				prevcmd := ctx.cmds[i-1].RectFilled
				const m = 1<<16 - 1
				sr, sg, sb, sa := cmd.Color.RGBA()
				a := (m - sa) * 0x101
				cmd.Color.R = uint8((uint32(prevcmd.Color.R)*a/m + sr) >> 8)
				cmd.Color.G = uint8((uint32(prevcmd.Color.G)*a/m + sg) >> 8)
				cmd.Color.B = uint8((uint32(prevcmd.Color.B)*a/m + sb) >> 8)
				cmd.Color.A = uint8((uint32(prevcmd.Color.A)*a/m + sa) >> 8)
			}
			colimg := image.NewUniform(cmd.Color)
			op := draw.Over
			if cmd.Color.A == 0xff {
				op = draw.Src
			}

			body := icmd.Rectangle()

			var lwing, rwing image.Rectangle

			// rounding is true if rounding has been requested AND we can draw it
			rounding := cmd.Rounding > 0 && int(cmd.Rounding*2) < icmd.W && int(cmd.Rounding*2) < icmd.H

			if rounding {
				body.Min.X += int(cmd.Rounding)
				body.Max.X -= int(cmd.Rounding)

				lwing = image.Rect(icmd.X, icmd.Y+int(cmd.Rounding), icmd.X+int(cmd.Rounding), icmd.Y+icmd.H-int(cmd.Rounding))
				rwing = image.Rect(icmd.X+icmd.W-int(cmd.Rounding), lwing.Min.Y, icmd.X+icmd.W, lwing.Max.Y)
			}

			bordopt := false

			if ok, border := borderOptimize(icmd, ctx.cmds, i+1); ok {
				// only draw parts of body if this command can be optimized to a border with the next command

				bordopt = true

				if ctx.cmds[i+1].RectFilled.Color.A != 0xff {
					transparentBorderOptimization = true
				}

				border += int(ctx.cmds[i+1].RectFilled.Rounding)

				top := image.Rect(body.Min.X, body.Min.Y, body.Max.X, body.Min.Y+border)
				bot := image.Rect(body.Min.X, body.Max.Y-border, body.Max.X, body.Max.Y)

				drawFill(img, top, colimg, top.Min, op)
				drawFill(img, bot, colimg, bot.Min, op)

				if border < int(cmd.Rounding) {
					// wings need shrinking
					d := int(cmd.Rounding) - border
					lwing.Max.Y -= d
					rwing.Min.Y += d
				} else {
					// display extra wings
					d := border - int(cmd.Rounding)

					xlwing := image.Rect(top.Min.X, top.Max.Y, top.Min.X+d, bot.Min.Y)
					xrwing := image.Rect(top.Max.X-d, top.Max.Y, top.Max.X, bot.Min.Y)

					drawFill(img, xlwing, colimg, xlwing.Min, op)
					drawFill(img, xrwing, colimg, xrwing.Min, op)
				}

				brrect++
			} else {
				drawFill(img, body, colimg, body.Min, op)
				if cmd.Rounding == 0 {
					if op == draw.Src {
						frect++
					} else {
						frectover++
					}
				} else {
					frrect++
				}
			}

			if rounding {
				drawFill(img, lwing, colimg, lwing.Min, op)
				drawFill(img, rwing, colimg, rwing.Min, op)

				rangle := math.Pi / 2

				if rasterizer == nil {
					setupRasterizer()
				}

				minx := img.Bounds().Min.X
				miny := img.Bounds().Min.Y

				roundAngle(icmd.X+icmd.W-int(cmd.Rounding)-minx, icmd.Y+int(cmd.Rounding)-miny, cmd.Rounding, -math.Pi/2, rangle, cmd.Color)
				roundAngle(icmd.X+icmd.W-int(cmd.Rounding)-minx, icmd.Y+icmd.H-int(cmd.Rounding)-miny, cmd.Rounding, 0, rangle, cmd.Color)
				roundAngle(icmd.X+int(cmd.Rounding)-minx, icmd.Y+icmd.H-int(cmd.Rounding)-miny, cmd.Rounding, math.Pi/2, rangle, cmd.Color)
				roundAngle(icmd.X+int(cmd.Rounding)-minx, icmd.Y+int(cmd.Rounding)-miny, cmd.Rounding, math.Pi, rangle, cmd.Color)
			}

			if perfUpdate {
				if bordopt {
					brecttim += time.Since(t0)
				} else {
					if cmd.Rounding > 0 {
						frrecttim += time.Since(t0)
					} else {
						d := time.Since(t0)
						if op == draw.Src {
							frecttim += d
						} else {
							if d > 8*time.Millisecond {
								fmt.Printf("outstanding rect")
							}
							frectovertim += d
						}
					}
				}
			}

		case command.TriangleFilledCmd:
			cmd := icmd.TriangleFilled
			if rasterizer == nil {
				setupRasterizer()
			}
			minx := img.Bounds().Min.X
			miny := img.Bounds().Min.Y
			rasterizer.Clear()
			rasterizer.Start(fixed.P(cmd.A.X-minx, cmd.A.Y-miny))
			rasterizer.Add1(fixed.P(cmd.B.X-minx, cmd.B.Y-miny))
			rasterizer.Add1(fixed.P(cmd.C.X-minx, cmd.C.Y-miny))
			rasterizer.Add1(fixed.P(cmd.A.X-minx, cmd.A.Y-miny))
			painter.SetColor(cmd.Color)
			rasterizer.Rasterize(painter)
			ftri++

			if perfUpdate {
				tritim += time.Since(t0)
			}

		case command.CircleFilledCmd:
			if rasterizer == nil {
				setupRasterizer()
			}
			rasterizer.Clear()
			startp := traceArc(rasterizer, float64(icmd.X-img.Bounds().Min.X)+float64(icmd.W/2), float64(icmd.Y-img.Bounds().Min.Y)+float64(icmd.H/2), float64(icmd.W/2), float64(icmd.H/2), 0, -math.Pi*2, true)
			rasterizer.Add1(startp) // closes path
			painter.SetColor(icmd.CircleFilled.Color)
			rasterizer.Rasterize(painter)
			fcirc++

		case command.ImageCmd:
			draw.Draw(img, icmd.Rectangle(), icmd.Image.Img, image.Point{}, draw.Src)

		case command.TextCmd:
			dstimg := wimg.SubImage(img.Bounds().Intersect(icmd.Rectangle())).(*image.RGBA)
			d := font.Drawer{
				Dst:  dstimg,
				Src:  image.NewUniform(icmd.Text.Foreground),
				Face: icmd.Text.Face,
				Dot:  fixed.P(icmd.X, icmd.Y+icmd.Text.Face.Metrics().Ascent.Ceil())}

			start := 0
			for i := range icmd.Text.String {
				if icmd.Text.String[i] == '\n' {
					d.DrawString(icmd.Text.String[start:i])
					d.Dot.X = fixed.I(icmd.X)
					d.Dot.Y += fixed.I(FontHeight(icmd.Text.Face))
					start = i + 1
				}
			}
			if start < len(icmd.Text.String) {
				d.DrawString(icmd.Text.String[start:])
			}
			txt++
			if perfUpdate {
				txttim += time.Since(t0)
			}
		default:
			panic(UnknownCommandErr)
		}

		if dumpFrame {
			ctx.cmdstim = append(ctx.cmdstim, time.Since(t0))
		}
	}

	if perfUpdate {
		fmt.Printf("triangle: %0.4fms text: %0.4fms brect: %0.4fms frect: %0.4fms frectover: %0.4fms frrect %0.4f\n", tritim.Seconds()*1000, txttim.Seconds()*1000, brecttim.Seconds()*1000, frecttim.Seconds()*1000, frectovertim.Seconds()*1000, frrecttim.Seconds()*1000)
	}

	cnt++
	if perfUpdate /*&& (cnt%100) == 0*/ {
		fmt.Printf("ln %d, frect %d, frectover %d, frrect %d, brrect %d, ftri %d, circ %d, fcirc %d, txt %d\n", ln, frect, frectover, frrect, brrect, ftri, circ, fcirc, txt)
		ln, frect, frectover, frrect, brrect, ftri, circ, fcirc, txt = 0, 0, 0, 0, 0, 0, 0, 0, 0
	}

	return len(ctx.cmds)
}

// Returns true if cmds[idx] is a shrunk version of CommandFillRect and its
// color is not semitransparent and the border isn't greater than 128
func borderOptimize(cmd *command.Command, cmds []command.Command, idx int) (ok bool, border int) {
	if idx >= len(cmds) {
		return false, 0
	}

	if cmd.Kind != command.RectFilledCmd || cmds[idx].Kind != command.RectFilledCmd {
		return false, 0
	}

	cmd2 := cmds[idx]

	if cmd.RectFilled.Color.A != 0xff && cmd2.RectFilled.Color.A != 0xff {
		return false, 0
	}

	border = cmd2.X - cmd.X
	if border <= 0 || border > 128 {
		return false, 0
	}

	if shrinkRect(cmd.Rect, border) != cmd2.Rect {
		return false, 0
	}

	return true, border
}

func floatP(x, y float64) fixed.Point26_6 {
	return fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)}
}

// TraceArc trace an arc using a Liner
func traceArc(t *raster.Rasterizer, x, y, rx, ry, start, angle float64, first bool) fixed.Point26_6 {
	end := start + angle
	clockWise := true
	if angle < 0 {
		clockWise = false
	}
	if !clockWise {
		for start < end {
			start += math.Pi * 2
		}
		end = start + angle
	}
	ra := (math.Abs(rx) + math.Abs(ry)) / 2
	da := math.Acos(ra/(ra+0.125)) * 2
	//normalize
	if !clockWise {
		da = -da
	}
	angle = start
	var curX, curY float64
	var startX, startY float64
	for {
		if (angle < end-da/4) != clockWise {
			curX = x + math.Cos(end)*rx
			curY = y + math.Sin(end)*ry
			t.Add1(floatP(curX, curY))
			return floatP(startX, startY)
		}
		curX = x + math.Cos(angle)*rx
		curY = y + math.Sin(angle)*ry

		angle += da
		if first {
			first = false
			startX, startY = curX, curY
			t.Start(floatP(curX, curY))
		} else {
			t.Add1(floatP(curX, curY))
		}
	}
}

type myRGBAPainter struct {
	Image *image.RGBA
	// cr, cg, cb and ca are the 16-bit color to paint the spans.
	cr, cg, cb, ca uint32
}

// SetColor sets the color to paint the spans.
func (r *myRGBAPainter) SetColor(c color.Color) {
	r.cr, r.cg, r.cb, r.ca = c.RGBA()
}

func (r *myRGBAPainter) Paint(ss []raster.Span, done bool) {
	b := r.Image.Bounds()
	cr8 := uint8(r.cr >> 8)
	cg8 := uint8(r.cg >> 8)
	cb8 := uint8(r.cb >> 8)
	for _, s := range ss {
		s.Y += b.Min.Y
		s.X0 += b.Min.X
		s.X1 += b.Min.X
		if s.Y < b.Min.Y {
			continue
		}
		if s.Y >= b.Max.Y {
			return
		}
		if s.X0 < b.Min.X {
			s.X0 = b.Min.X
		}
		if s.X1 > b.Max.X {
			s.X1 = b.Max.X
		}
		if s.X0 >= s.X1 {
			continue
		}
		// This code mimics drawGlyphOver in $GOROOT/src/image/draw/draw.go.
		ma := s.Alpha
		const m = 1<<16 - 1
		i0 := (s.Y-r.Image.Rect.Min.Y)*r.Image.Stride + (s.X0-r.Image.Rect.Min.X)*4
		i1 := i0 + (s.X1-s.X0)*4
		if ma != m || r.ca != m {
			for i := i0; i < i1; i += 4 {
				dr := uint32(r.Image.Pix[i+0])
				dg := uint32(r.Image.Pix[i+1])
				db := uint32(r.Image.Pix[i+2])
				da := uint32(r.Image.Pix[i+3])
				a := (m - (r.ca * ma / m)) * 0x101
				r.Image.Pix[i+0] = uint8((dr*a + r.cr*ma) / m >> 8)
				r.Image.Pix[i+1] = uint8((dg*a + r.cg*ma) / m >> 8)
				r.Image.Pix[i+2] = uint8((db*a + r.cb*ma) / m >> 8)
				r.Image.Pix[i+3] = uint8((da*a + r.ca*ma) / m >> 8)
			}
		} else {
			for i := i0; i < i1; i += 4 {
				r.Image.Pix[i+0] = cr8
				r.Image.Pix[i+1] = cg8
				r.Image.Pix[i+2] = cb8
				r.Image.Pix[i+3] = 0xff
			}
		}
	}
}

type dockedNodeType uint8

const (
	dockedNodeLeaf dockedNodeType = iota
	dockedNodeVert
	dockedNodeHoriz
)

type dockedTree struct {
	Type  dockedNodeType
	Split ScalableSplit
	Child [2]*dockedTree
	W     *Window
}

func (t *dockedTree) Update(bounds rect.Rect, scaling float64) *dockedTree {
	if t == nil {
		return nil
	}
	switch t.Type {
	case dockedNodeVert:
		b0, b1, _ := t.Split.verticalnw(bounds, scaling)
		t.Child[0] = t.Child[0].Update(b0, scaling)
		t.Child[1] = t.Child[1].Update(b1, scaling)
	case dockedNodeHoriz:
		b0, b1, _ := t.Split.horizontalnw(bounds, scaling)
		t.Child[0] = t.Child[0].Update(b0, scaling)
		t.Child[1] = t.Child[1].Update(b1, scaling)
	case dockedNodeLeaf:
		if t.W != nil {
			t.W.Bounds = bounds
			t.W.ctx.updateWindow(t.W)
			if t.W == nil {
				return nil
			}
			if t.W.close {
				t.W = nil
				return nil
			}
			return t
		}
		return nil
	}
	if t.Child[0] == nil {
		return t.Child[1]
	}
	if t.Child[1] == nil {
		return t.Child[0]
	}
	return t
}

func (t *dockedTree) walkExt(fn func(t *dockedTree)) {
	if t == nil {
		return
	}
	switch t.Type {
	case dockedNodeVert, dockedNodeHoriz:
		fn(t)
		t.Child[0].walkExt(fn)
		t.Child[1].walkExt(fn)
	case dockedNodeLeaf:
		fn(t)
	}
}

func (t *dockedTree) Walk(fn func(t *Window) *Window) {
	t.walkExt(func(t *dockedTree) {
		if t.Type == dockedNodeLeaf && t.W != nil {
			t.W = fn(t.W)
		}
	})
}

func newDockedLeaf(win *Window) *dockedTree {
	r := &dockedTree{Type: dockedNodeLeaf, W: win}
	r.Split.MinSize = 40
	return r
}

func (t *dockedTree) Dock(win *Window, pos image.Point, bounds rect.Rect, scaling float64) (bool, rect.Rect) {
	if t == nil {
		return false, rect.Rect{}
	}
	switch t.Type {
	case dockedNodeVert:
		b0, b1, _ := t.Split.verticalnw(bounds, scaling)
		canDock, r := t.Child[0].Dock(win, pos, b0, scaling)
		if canDock {
			return canDock, r
		}
		canDock, r = t.Child[1].Dock(win, pos, b1, scaling)
		if canDock {
			return canDock, r
		}
	case dockedNodeHoriz:
		b0, b1, _ := t.Split.horizontalnw(bounds, scaling)
		canDock, r := t.Child[0].Dock(win, pos, b0, scaling)
		if canDock {
			return canDock, r
		}
		canDock, r = t.Child[1].Dock(win, pos, b1, scaling)
		if canDock {
			return canDock, r
		}
	case dockedNodeLeaf:
		v := percentages(bounds, 0.03)
		for i := range v {
			if v[i].Contains(pos) {
				if t.W == nil {
					if win != nil {
						t.W = win
						win.ctx.dockWindow(win)
					}
					return true, bounds
				}
				w := percentages(bounds, 0.5)
				if win != nil {
					if i < 2 {
						// horizontal split
						t.Type = dockedNodeHoriz
						t.Split.Size = int(float64(w[0].H) / scaling)
						t.Child[i] = newDockedLeaf(win)
						t.Child[-i+1] = newDockedLeaf(t.W)
					} else {
						// vertical split
						t.Type = dockedNodeVert
						t.Split.Size = int(float64(w[2].W) / scaling)
						t.Child[i-2] = newDockedLeaf(win)
						t.Child[-(i-2)+1] = newDockedLeaf(t.W)
					}

					t.W = nil
					win.ctx.dockWindow(win)
				}
				return true, w[i]
			}
		}
	}
	return false, rect.Rect{}
}

func (ctx *context) dockWindow(win *Window) {
	win.undockedSz = image.Point{win.Bounds.W, win.Bounds.H}
	win.flags |= windowDocked
	win.layout.Flags |= windowDocked
	ctx.dockedCnt--
	win.idx = ctx.dockedCnt
	for i := range ctx.Windows {
		if ctx.Windows[i] == win {
			if i+1 < len(ctx.Windows) {
				copy(ctx.Windows[i:], ctx.Windows[i+1:])
			}
			ctx.Windows = ctx.Windows[:len(ctx.Windows)-1]
			return
		}
	}
}

func (t *dockedTree) Undock(win *Window) {
	t.Walk(func(w *Window) *Window {
		if w == win {
			return nil
		}
		return w
	})
	win.flags &= ^windowDocked
	win.layout.Flags &= ^windowDocked
	win.Bounds.H = win.undockedSz.Y
	win.Bounds.W = win.undockedSz.X
	win.idx = len(win.ctx.Windows)
	win.ctx.Windows = append(win.ctx.Windows, win)
}

func (t *dockedTree) Scale(win *Window, delta image.Point, scaling float64) image.Point {
	if t == nil || (delta.X == 0 && delta.Y == 0) {
		return image.Point{}
	}
	switch t.Type {
	case dockedNodeVert:
		d0 := t.Child[0].Scale(win, delta, scaling)
		if d0.X != 0 {
			t.Split.Size += int(float64(d0.X) / scaling)
			if t.Split.Size <= t.Split.MinSize {
				t.Split.Size = t.Split.MinSize
			}
			d0.X = 0
		}
		if d0 != image.ZP {
			return d0
		}
		return t.Child[1].Scale(win, delta, scaling)
	case dockedNodeHoriz:
		d0 := t.Child[0].Scale(win, delta, scaling)
		if d0.Y != 0 {
			t.Split.Size += int(float64(d0.Y) / scaling)
			if t.Split.Size <= t.Split.MinSize {
				t.Split.Size = t.Split.MinSize
			}
			d0.Y = 0
		}
		if d0 != image.ZP {
			return d0
		}
		return t.Child[1].Scale(win, delta, scaling)
	case dockedNodeLeaf:
		if t.W == win {
			return delta
		}
	}
	return image.Point{}
}

func (ctx *context) ResetWindows() *DockSplit {
	ctx.DockedWindows = dockedTree{}
	ctx.Windows = ctx.Windows[:1]
	ctx.dockedCnt = 0
	return &DockSplit{ctx, &ctx.DockedWindows}
}

type DockSplit struct {
	ctx  *context
	node *dockedTree
}

func (ds *DockSplit) Split(horiz bool, size int) (left, right *DockSplit) {
	if horiz {
		ds.node.Type = dockedNodeHoriz
	} else {
		ds.node.Type = dockedNodeVert
	}
	ds.node.Split.Size = size
	ds.node.Child[0] = &dockedTree{Type: dockedNodeLeaf, Split: ScalableSplit{MinSize: 40}}
	ds.node.Child[1] = &dockedTree{Type: dockedNodeLeaf, Split: ScalableSplit{MinSize: 40}}
	return &DockSplit{ds.ctx, ds.node.Child[0]}, &DockSplit{ds.ctx, ds.node.Child[1]}
}

func (ds *DockSplit) Open(title string, flags WindowFlags, rect rect.Rect, scale bool, updateFn UpdateFn) {
	ds.ctx.popupOpen(title, flags, rect, scale, updateFn)
	ds.node.Type = dockedNodeLeaf
	ds.node.W = ds.ctx.Windows[len(ds.ctx.Windows)-1]
	ds.ctx.dockWindow(ds.node.W)
}

func percentages(bounds rect.Rect, f float64) (r [4]rect.Rect) {
	pw := int(float64(bounds.W) * f)
	ph := int(float64(bounds.H) * f)
	// horizontal split
	r[0] = bounds
	r[0].H = ph
	r[1] = bounds
	r[1].Y += r[1].H - ph
	r[1].H = ph

	// vertical split
	r[2] = bounds
	r[2].W = pw
	r[3] = bounds
	r[3].X += r[3].W - pw
	r[3].W = pw
	return
}

func clip(dst *image.RGBA, r *image.Rectangle, src image.Image, sp *image.Point) {
	orig := r.Min
	*r = r.Intersect(dst.Bounds())
	*r = r.Intersect(src.Bounds().Add(orig.Sub(*sp)))
	dx := r.Min.X - orig.X
	dy := r.Min.Y - orig.Y
	if dx == 0 && dy == 0 {
		return
	}
	sp.X += dx
	sp.Y += dy
}

func drawFill(dst *image.RGBA, r image.Rectangle, src *image.Uniform, sp image.Point, op draw.Op) {
	clip(dst, &r, src, &sp)
	if r.Empty() {
		return
	}
	sr, sg, sb, sa := src.RGBA()
	switch op {
	case draw.Over:
		drawFillOver(dst, r, sr, sg, sb, sa)
	case draw.Src:
		drawFillSrc(dst, r, sr, sg, sb, sa)
	default:
		draw.Draw(dst, r, src, sp, op)
	}
}

func drawFillSrc(dst *image.RGBA, r image.Rectangle, sr, sg, sb, sa uint32) {
	sr8 := uint8(sr >> 8)
	sg8 := uint8(sg >> 8)
	sb8 := uint8(sb >> 8)
	sa8 := uint8(sa >> 8)
	// The built-in copy function is faster than a straightforward for loop to fill the destination with
	// the color, but copy requires a slice source. We therefore use a for loop to fill the first row, and
	// then use the first row as the slice source for the remaining rows.
	i0 := dst.PixOffset(r.Min.X, r.Min.Y)
	i1 := i0 + r.Dx()*4
	for i := i0; i < i1; i += 4 {
		dst.Pix[i+0] = sr8
		dst.Pix[i+1] = sg8
		dst.Pix[i+2] = sb8
		dst.Pix[i+3] = sa8
	}
	firstRow := dst.Pix[i0:i1]
	for y := r.Min.Y + 1; y < r.Max.Y; y++ {
		i0 += dst.Stride
		i1 += dst.Stride
		copy(dst.Pix[i0:i1], firstRow)
	}
}

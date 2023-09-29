package main

import (
	"image/color"
	"strings"

	"github.com/aarzilli/nucular"
	"github.com/aarzilli/nucular/rect"
	"github.com/aarzilli/nucular/richtext"
	"github.com/aarzilli/nucular/style"

	"github.com/aarzilli/nucular/_examples/richtext/internal/assets"
	"github.com/aarzilli/nucular/font"
)

//go:generate go-bindata -o internal/assets/assets.go -pkg assets DejaVuSans.ttf DejaVuSans-Bold.ttf DejaVuSans-Oblique.ttf

var rtxt *richtext.RichText
var selected int
var align int
var autowrap bool
var searchEd nucular.TextEditor
var lastNeedle string

var proportional, header, monospace, bold, italic font.Face

const defaultFlags = richtext.Selectable | richtext.ShowTick | richtext.Clipboard | richtext.Keyboard

func main() {
	rtxt = richtext.New(defaultFlags)
	wnd := nucular.NewMasterWindow(0, "Rich Text", updatefn)
	wnd.SetStyle(style.FromTheme(style.DarkTheme, 2.0))

	regularData, _ := assets.Asset("DejaVuSans.ttf")
	boldData, _ := assets.Asset("DejaVuSans-Bold.ttf")
	italicData, _ := assets.Asset("DejaVuSans-Oblique.ttf")

	proportional, _ = font.NewFace(regularData, int(float64(12)*wnd.Style().Scaling))
	header, _ = font.NewFace(regularData, int(float64(21)*wnd.Style().Scaling))
	monospace = wnd.Style().Font

	bold, _ = font.NewFace(boldData, int(float64(12)*wnd.Style().Scaling))
	italic, _ = font.NewFace(italicData, int(float64(12)*wnd.Style().Scaling))

	searchEd.Flags = nucular.EditField

	wnd.Main()
}

var appendTestIndex int

func updatefn(w *nucular.Window) {
	changed := false

	w.MenubarBegin()
	w.Row(20).Static(150, 100, 150, 150)
	newselected := w.ComboSimple([]string{"Vispa Teresa", "Big Enchillada", "Fancy Enchillada", "Fibonacci", "Append Test"}, selected, 20)
	w.CheckboxText("Auto wrap", &autowrap)
	newalign := w.ComboSimple([]string{"Align left (dumb)", "Align left", "Align right", "Align center", "Align justified"}, align, 20)
	if w.ButtonText("Open popup") {
		pw := newPopupWindow()
		w.Master().PopupOpen("Popup", nucular.WindowDynamic|nucular.WindowTitle|nucular.WindowNoScrollbar|nucular.WindowMovable|nucular.WindowBorder, rect.Rect{100, 100, 900, 700}, false, pw.Update)
	}

	w.Row(30).Static(100, 200, 100, 0, 80, 80, 80)
	w.Label("Search:", "LC")
	searchEd.Edit(w)
	if w.ButtonText("Next") {
		rtxt.FollowCursor()
		lastNeedle = string(searchEd.Buffer)
		rtxt.Sel.S = rtxt.Sel.E
		rtxt.Look(lastNeedle, true)
	}
	if selected == 4 {
		//source := bigEnchilladaJap
		source := bigEnchillada
		w.Spacing(1)
		if w.ButtonText("Reset") {
			changed = true
		}
		if w.ButtonText("Append") {
			if appendTestIndex < len(source) {
				end := appendTestIndex + 50
				if end > len(source) {
					end = len(source)
				}
				c := rtxt.Append(true)
				c.SetStyle(richtext.TextStyle{Face: proportional})
				if appendTestIndex/50%3 == 0 && appendTestIndex+10 < end {
					c.SetStyle(richtext.TextStyle{Face: proportional, BgColor: color.RGBA{0xff, 0x00, 0x00, 0xff}})
					c.Text(source[appendTestIndex : appendTestIndex+10])
					c.SetStyle(richtext.TextStyle{Face: proportional})
					c.Text(source[appendTestIndex+10 : end])
				} else {
					c.Text(source[appendTestIndex:end])
				}
				c.End()
				appendTestIndex = end
			}
		}
		if w.ButtonText("Tail") {
			rtxt.Tail(5)
		}
	} else {
		w.Spacing(4)
	}
	w.MenubarEnd()

	if string(searchEd.Buffer) != lastNeedle {
		rtxt.FollowCursor()
		lastNeedle = string(searchEd.Buffer)
		rtxt.Look(lastNeedle, true)
	}

	if newselected != selected {
		selected = newselected
		changed = true
	}
	if newalign != align {
		align = newalign
		changed = true
	}

	rtxt.Flags = defaultFlags
	if autowrap {
		rtxt.Flags |= richtext.AutoWrap
	}

	if c := rtxt.Rows(w, changed); c != nil {
		switch selected {
		case 0:
			c.Align(richtext.Align(align))
			c.SetStyle(richtext.TextStyle{Face: header, Cursor: font.TextCursor})
			c.Text("Vispa Teresa\n")
			c.SetStyle(richtext.TextStyle{Face: proportional, Cursor: font.TextCursor})
			c.Text("\n")
			c.Text("La vispa Teresa\navea tra l'erbetta\na volo sorpresa\ngentil farfalletta\n\n")
			c.Text("E tutta giuliva\nstringendola viva\ngridava a distesa\nl'ho presa! l'ho presa!\n\n")
			c.SaveStyle()
			c.SetStyle(richtext.TextStyle{Face: proportional, Color: color.RGBA{0x00, 0x88, 0xdd, 0xff}, Flags: richtext.Underline})
			if c.Link("Link 1 (inline)", color.RGBA{0x00, 0xaa, 0xff, 0xff}, nil) {
				w.Master().PopupOpen("Clicked! (1)", nucular.WindowDefaultFlags, rect.Rect{0, 0, 200, 200}, true, func(w *nucular.Window) {
					w.Row(30).Dynamic(1)
					w.Label("Clicked!", "LC")
				})
			}
			c.RestoreStyle()
			c.Text(" ")
			c.SetStyle(richtext.TextStyle{Face: proportional, Color: color.RGBA{0x00, 0x88, 0xdd, 0xff}, Flags: richtext.Underline})
			c.Link("Link 2 (callback)", color.RGBA{0x00, 0xaa, 0xff, 0xff}, func() {
				w.Master().PopupOpen("Clicked! (2)", nucular.WindowDefaultFlags, rect.Rect{0, 0, 200, 200}, true, func(w *nucular.Window) {
					w.Row(30).Dynamic(1)
					w.Label("Clicked!", "LC")
				})
			})

		case 1:
			c.Align(richtext.Align(align))
			c.SetStyle(richtext.TextStyle{Face: proportional})
			c.Text(bigEnchillada)

		case 2:
			c.Align(richtext.Align(align))
			c.SetStyle(richtext.TextStyle{Face: proportional})
			c.Text(bigEnchillada)
			c.SetStyleForSel(findSel(bigEnchillada, "elite"), richtext.TextStyle{Face: header})
			c.SetStyleForSel(findSel(bigEnchillada, "elites"), richtext.TextStyle{Face: header, BgColor: color.RGBA{0xff, 0x00, 0x00, 0xff}})
			c.SetStyleForSel(findSel(bigEnchillada, "hivemind consciousness"), richtext.TextStyle{Face: header})
			c.SetStyleForSel(findSel(bigEnchillada, "Einstein's physics"), richtext.TextStyle{Face: bold})
			c.SetStyleForSel(findSel(bigEnchillada, "Max Planck physics"), richtext.TextStyle{Face: bold})
			c.SetStyleForSel(findSel(bigEnchillada, "it's a false hologram, it is artificial"), richtext.TextStyle{Face: italic})
			c.SetStyleForSel(findSel(bigEnchillada, "break-away civilization"), richtext.TextStyle{Face: proportional, Flags: richtext.Underline})
			c.SetStyleForSel(findSel(bigEnchillada, "cut off the pedophiles"), richtext.TextStyle{Face: header})
			c.SetStyleForSel(findSel(bigEnchillada, "lust for power"), richtext.TextStyle{Face: proportional, BgColor: color.RGBA{0x00, 0x00, 0xff, 0xff}})
			c.SetStyleForSel(findSel(bigEnchillada, `all physics showed it: there's at least 12 dimensions. And now all top scientists and billionaires are coming out and saying "it's a false hologram, it is artificial" the computers are scanning it and finding tension points where it's artificially projected and gravity is bleeding in to this universe. That's what they call dark matter.`), richtext.TextStyle{Face: proportional, Color: color.RGBA{0x00, 0x00, 0x00, 0xff}, BgColor: color.RGBA{0x00, 0xff, 0x00, 0xff}})
			c.SetStyleForSel(findSel(bigEnchillada, "And so Google was set up"), richtext.TextStyle{Face: proportional, Flags: richtext.Underline})
			c.SetStyleForSel(findSel(bigEnchillada, "they wanted to build a giant artificial system"), richtext.TextStyle{Face: proportional, Flags: richtext.Strikethrough})
			c.SetStyleForSel(findSel(bigEnchillada, "all of our thoughts go into it and we"), richtext.TextStyle{Face: proportional, Flags: richtext.Strikethrough | richtext.Underline})
			c.SetStyleForSel(findSel(bigEnchillada, "Google believes"), richtext.TextStyle{Face: header})

		case 3:
			c.Align(richtext.Align(align))
			c.SetStyle(richtext.TextStyle{Face: proportional})
			c.Text("func fib(n int) int {\n")
			c.ParagraphStyle(richtext.Align(align), color.RGBA{0xff, 0x00, 0x00, 0xff})
			c.Text(`	switch n {
	case 0:
		return 1
	case 1:
		return 1
	default:
		return fib(n-1) + fib(n-2)
`)
			c.ParagraphStyle(richtext.Align(align), color.RGBA{})
			c.Text(`	}
}
`)

		case 4:
			appendTestIndex = 0
			c.SetStyle(richtext.TextStyle{Face: proportional})
			c.Align(richtext.Align(align))
			c.Text("Start of test\n")
		}
		c.End()
	}
}

func findSel(haystack, needle string) richtext.Sel {
	n := strings.Index(haystack, needle)
	if n < 0 {
		panic("not found")
	}
	return richtext.Sel{int32(n), int32(n + len(needle))}
}

type popupWindow struct {
	rtxt *richtext.RichText
}

func newPopupWindow() *popupWindow {
	return &popupWindow{
		rtxt: richtext.New(defaultFlags | richtext.AutoWrap),
	}
}

func (pw *popupWindow) Update(w *nucular.Window) {
	w.Row(0).Dynamic(1)
	if c := pw.rtxt.Widget(w, false); c != nil {
		c.Align(richtext.AlignLeftDumb)
		c.Text(bigEnchillada)
		c.End()
	}
}

const bigEnchillada = `The elite are all about trascendence and living forever and the secret of the universe and they want to know all this. Some are good, some are bad, some are a mix. But the good one never want to organize, the bad ones instead they want to organize because the lust for power. Powerful consciusnesses don't want to dominate other people, they want to empower them so they don't tend to get together until things are late in the game, then they come together, evil is always defeated, because good is so much stronger.
And we are on this planet and, Einstein's physics showed it, and Max Planck physics showed it, all physics showed it: there's at least 12 dimensions. And now all top scientists and billionaires are coming out and saying "it's a false hologram, it is artificial" the computers are scanning it and finding tension points where it's artificially projected and gravity is bleeding in to this universe. That's what they call dark matter.
So we are like a thought or a dream, that's a whisp in some computer program, some God's mind, whatever (they're proving all, it's all coming out). Now, there's like this sub-transimission zone, below the third dimension, that's just turned over the most terrible things, it's what it resonates to, and it's trying to get up into the third dimension, that's just the basic level consciousness to launch into the next levels. And our species is already way up into the fifth, sixth dimension consciusnly, our best people. But there's this big war trying to basically destroy humanity, because humanity has free will and there's a decision to which level we want to go to. We have free will so evil is allowed to come and contend, not just good.
And the elites themselves believe that they are racing, using human technology, to try to take our best minds and build some type of break-away civilization where they're going to merge with machines, transcend and break away from the failed species that is man. Which is kind of like a false transmission because they are thinking what they are is ugly and bad, projecting onto themselves, instead of believing "no it's a human test about building us up".
And so Google was set up, 18 ~ 19 years ago (I knew this before it was declassified, I'm just saying I have good sources) that they wanted to build a giant artificial system and Google believes that the first artificial intelligence will be a supercomputer based on the neuron activity of the hivemind of humanity with billions of people wired into it with the internet of things. And so all of our thoughts go into it and we are actually building a computer that has real neurons in real time that's also psychically connected to us that are organic creatures so that they will have *current* prediction powers, *future* prediction power (a true crystal ball).
But the big secret is, once you have the crystal ball and know the future you can add stimuli beforehand and make decisions that control the future and it's the end of consciousness and freewill for individuals as we know and a true 2.0 (in a very bad way) hivemind consciousness with an AI jacked into everyone knowing our hopes and dreams, delivering it to us, not in some PKD wire-head system where we plug-in and give up on consciousness because of unlimited pleasure, but because we are already wired in and absorbed before we even knew it by giving over our consciousness to the system by our daily decisions, that it was able to manipulate and control, into a larger system.
There is now a human counterstrike to shut this off before it gets fully into place and block these systems and to try to have an actual debate about where humanity goes and cut off the pedophiles and psychic vampires that are in control of this AI system before humanity is destroyed.
`

const bigEnchilladaJap = `The elite 生き物 are all about 医者 trascendence and living 御腹  forever and the secret of 生き物 the universe and they 掌、手の平  want to know all this. 生き物 Some are good, 掌、手の平  some are bad, some are a mix. But the good one never want to organize, the bad 御腹  ones instead they want to organize because the lust for power. 御腹  Powerful consciusnesses don't want to dominate other people, they want to empower them so they don't tend to get together until things are late in the game, then they come together, evil is always defeated, because good is so much stronger.
And we are on this planet and, Einstein's physics 生き物 showed it, and Max Planck physics showed it, 御腹  all physics showed it: there's 掌、手の平  at least 12 dimensions. And now all top 掌、手の平  scientists and billionaires are coming out and saying "it's a false hologram, it is artificial" the computers are scanning it and finding tension points where it's artificially projected and gravity is bleeding in to this universe. That's what they call dark matter.
So we are like a thought or a dream, that's a whisp in 掌、手の平  some computer program, some God's mind, whatever (they're proving all, it's all coming out). Now, there's like this sub-transimission zone, below the third dimension, that's just turned over the most terrible things, it's what it resonates to, and it's trying to get up into the third dimension, that's just the basic 掌、手の平  level consciousness to launch into the next levels. And our species is already way up into the fifth, sixth dimension consciusnly, our best people. But there's this big war trying to basically destroy humanity, because humanity 掌、手の平  has free will and there's a decision to which level we 掌、手の平  want to go to. We have free will so evil is allowed to come and contend, not just good.
And the elites themselves 掌、手の平  believe that they are racing, using human technology, to try to take our best minds and build some type of break-away civilization where they're going to merge with machines, transcend 掌、手の平  and break away from the failed species 生き物 that is man. Which is kind of like a false transmission because they are thinking what they are is ugly and bad, projecting onto themselves, instead of believing "no it's a human test about building us up".
And so Google was set up, 18 ~ 19 years ago (I knew this before it was declassified, I'm just saying I have good sources) that they wanted to build a giant artificial system and Google believes that the first artificial intelligence will be a supercomputer based on the neuron activity of the hivemind of humanity with billions of people wired into it with the internet of things. And so all of our thoughts go into it and we are actually building a computer that has real neurons in real time that's also psychically connected to us that are organic creatures so that they will have *current* prediction powers, *future* prediction power (a true crystal ball).
But the big secret is, once you have the crystal ball and know the future you can add stimuli beforehand and make decisions that control the future and it's the end of consciousness and freewill for individuals as we know and a true 2.0 (in a very bad way) hivemind consciousness with an AI jacked into everyone knowing our hopes and dreams, delivering it to us, not in some PKD wire-head system where we plug-in and give up on consciousness because of unlimited pleasure, but because we are already wired in and absorbed before we even knew it by giving over our consciousness to the system by our daily decisions, that it was able to manipulate and control, into a larger system.
There is now a human counterstrike to shut this off before it gets fully into place and block these systems and to try to have an actual debate about where humanity goes and cut off the pedophiles and psychic vampires that are in control of this AI system before humanity is destroyed.
`

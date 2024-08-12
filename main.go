package main

import (
	"embed"
	_ "embed"
	"fmt"
	"io"
	"math"
	"strconv"
	"syscall/js"

	baseline "github.com/boxesandglue/baseline-pdf"
	"github.com/boxesandglue/boxesandglue/backend/bag"
	"github.com/boxesandglue/boxesandglue/backend/font"
	"github.com/boxesandglue/boxesandglue/backend/node"
	"github.com/boxesandglue/boxesandglue/frontend"
	"github.com/boxesandglue/textlayout/harfbuzz"
)

//go:embed fonts/garamond/CormorantGaramond-Regular.ttf
var f embed.FS

func mknodelist(fnt *font.Font, atoms []font.Atom) node.Node {
	var head, cur node.Node
	var lastglue node.Node

	for _, r := range atoms {
		if r.IsSpace {
			if lastglue == nil {
				g := node.NewGlue()
				g.Width = fnt.Space
				g.Stretch = fnt.SpaceStretch
				g.Shrink = fnt.SpaceShrink
				head = node.InsertAfter(head, cur, g)
				cur = g
				lastglue = g
			}
		} else {
			n := node.NewGlyph()
			n.Hyphenate = r.Hyphenate
			n.Codepoint = r.Codepoint
			n.Components = r.Components
			n.Font = fnt
			n.Width = r.Advance
			n.Height = r.Height
			n.Depth = r.Depth
			head = node.InsertAfter(head, cur, n)
			cur = n
			lastglue = nil

			if r.Kernafter != 0 {
				k := node.NewKern()
				k.Kern = r.Kernafter
				head = node.InsertAfter(head, cur, k)
				cur = k
			}
		}
	}
	return head
}

type position struct {
	component string
	xpos      bag.ScaledPoint
	ypos      bag.ScaledPoint
}

type lineinfo struct {
	linenumber int
	r          float64
	demerits   int
	fitness    int
}

func getVPositions(ypos bag.ScaledPoint, vl node.Node, level int) ([]position, []int) {
	var badness []int
	var g []position
	glyphs := []position{}
	for e := vl; e != nil; e = e.Next() {
		switch t := e.(type) {
		case *node.VList:
			g, badness = getVPositions(ypos, t.List, level+1)

			glyphs = append(glyphs, g...)
			ypos += t.Height + t.Depth
		case *node.HList:
			g, _ = getHPositions(ypos, t.List, level+1)
			badness = append(badness, t.Badness)
			glyphs = append(glyphs, g...)
			ypos += t.Height + t.Depth
		case *node.Glue:
			ypos += t.Width
		default:

		}
	}

	return glyphs, badness
}

func getHPositions(ypos bag.ScaledPoint, hl node.Node, level int) ([]position, []int) {
	glyphs := []position{}
	var g []position
	var badness []int
	xpos := bag.ScaledPoint(0)
	for e := hl; e != nil; e = e.Next() {
		switch t := e.(type) {
		case *node.VList:
			g, badness = getVPositions(ypos, t.List, level+1)
			glyphs = append(glyphs, g...)
		case *node.HList:
			g, badness = getHPositions(ypos, t.List, level+1)
			glyphs = append(glyphs, g...)
		case *node.Glue:
			xpos += t.Width
		case *node.Glyph:
			glyphs = append(glyphs, position{t.Components, xpos, ypos})
			xpos += t.Width
		case *node.Kern:
			xpos += t.Kern
		default:
			// fmt.Printf("~~> getHPositions %#v\n", t)
		}
	}
	return glyphs, badness
}

type retinfo struct {
	positions []position
	badness   []int
	li        []lineinfo
	height    bag.ScaledPoint
}

func getPositions(settings *node.LinebreakSettings, text string, fontsize bag.ScaledPoint, hyphenate bool) (*retinfo, error) {
	pdf := baseline.NewPDFWriter(io.Discard)
	data, err := f.ReadFile("fonts/garamond/CormorantGaramond-Regular.ttf")
	if err != nil {
		return nil, err
	}

	face, err := baseline.NewFaceFromData(pdf, data, 0)
	if err != nil {
		return nil, err
	}

	fnt := font.NewFont(face, fontsize)

	atoms := fnt.Shape(text, []harfbuzz.Feature{})
	nl := mknodelist(fnt, atoms)
	if hyphenate {
		l, err := frontend.GetLanguage("en")
		if err != nil {
			return nil, nil
		}
		frontend.Hyphenate(nl, l)
	}
	nl, _ = node.AppendLineEndAfter(nl, node.Tail(nl))
	vl, breakpoints := node.Linebreak(nl, settings)
	var lines []lineinfo
	for i, bp := range breakpoints {
		lines = append(lines, lineinfo{
			linenumber: i + 1,
			r:          math.Round(bp.R*10000.0) / 10000.0,
			demerits:   bp.Demerits,
			fitness:    bp.Fitness,
		})
	}
	g, badness := getVPositions(fontsize, vl, 0)
	_ = badness
	ri := &retinfo{
		positions: g,
		badness:   badness,
		height:    vl.Height + vl.Depth,
		li:        lines,
	}
	return ri, nil
}

func parseInt(val js.Value) int {
	switch val.Type() {
	case js.TypeNumber:
		return val.Int()
	case js.TypeString:
		num, err := strconv.Atoi(val.String())
		if err != nil {
			return 1
		}
		return num
	}
	return 1
}

func parseFloat(val js.Value) float64 {
	switch val.Type() {
	case js.TypeNumber:
		return val.Float()
	case js.TypeString:
		flt, err := strconv.ParseFloat(val.String(), 64)
		if err != nil {
			return 1.0
		}
		return flt
	}
	return 1.0
}

func parseWidth(val js.Value) bag.ScaledPoint {
	var wd string
	switch val.Type() {
	case js.TypeNumber:
		wd = fmt.Sprintf("%d", val.Int())
	case js.TypeString:
		wd = val.String()
	}
	if wd == "" {
		return 10 * bag.Factor
	}
	return bag.MustSp(wd + "pt")
}

func returnGetPositions() js.Func {
	jsFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			fmt.Println("getpositions requires one argument")
			return nil
		}
		settings := node.NewLinebreakSettings()

		firstArg := args[0]
		text := firstArg.Get("text").String()
		fontsize := parseWidth(firstArg.Get("fontsize"))

		settings.LineHeight = parseWidth(firstArg.Get("leading"))
		settings.HSize = parseWidth(firstArg.Get("hsize"))
		settings.Hyphenpenalty = parseInt(firstArg.Get("hyphenpenalty"))
		settings.DemeritsFitness = parseInt(firstArg.Get("demeritsfitness"))

		settings.Tolerance = parseFloat(firstArg.Get("tolerance"))
		hyphenate := firstArg.Get("hyphenate").Bool()
		settings.SqueezeOverfullBoxes = firstArg.Get("squeezeoverfullboxes").Bool()
		settings.HangingPunctuationEnd = firstArg.Get("hangingpunctuationend").Bool()

		g, err := getPositions(settings, text, fontsize, hyphenate)
		if err != nil {
			panic("this should not happen")
		}

		charPosition := []any{}
		lineinfo := []any{}

		for i, l := range g.li {
			obj := map[string]any{
				"line":     l.linenumber,
				"r":        l.r,
				"demerits": l.demerits,
				"fitness":  l.fitness,
				"badness":  g.badness[i],
			}
			lineinfo = append(lineinfo, obj)
		}

		for _, i := range g.positions {
			obj := map[string]any{
				"char": i.component,
				"xpos": i.xpos.ToPT(),
				"ypos": i.ypos.ToPT(),
			}
			charPosition = append(charPosition, obj)
		}
		ret := map[string]any{
			"positions": charPosition,
			"height":    g.height.ToPT(),
			"lines":     lineinfo,
		}
		return ret
	})
	return jsFunc
}

func main() {
	js.Global().Set("getpositions", returnGetPositions())
	<-make(chan bool)
}

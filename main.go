package main

import (
	"embed"
	_ "embed"
	"fmt"
	"io"
	"strconv"
	"syscall/js"

	baseline "github.com/speedata/baseline-pdf"
	"github.com/speedata/boxesandglue/backend/bag"
	"github.com/speedata/boxesandglue/backend/font"
	"github.com/speedata/boxesandglue/backend/node"
	"github.com/speedata/boxesandglue/frontend"
	"github.com/speedata/textlayout/harfbuzz"
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
	l, err := frontend.GetLanguage("en")
	if err != nil {
		return nil
	}
	frontend.Hyphenate(head, l)
	return head
}

type position struct {
	component string
	xpos      bag.ScaledPoint
	ypos      bag.ScaledPoint
}

func getVPositions(ypos bag.ScaledPoint, vl node.Node) []position {
	glyphs := []position{}
	for e := vl; e != nil; e = e.Next() {
		switch t := e.(type) {
		case *node.VList:
			g := getVPositions(ypos, t.List)
			glyphs = append(glyphs, g...)
			ypos += t.Height + t.Depth
		case *node.HList:
			g := getHPositions(ypos, t.List)
			glyphs = append(glyphs, g...)
			ypos += t.Height + t.Depth
		case *node.Glue:
			ypos += t.Width
		default:

		}
	}
	return glyphs
}

func getHPositions(ypos bag.ScaledPoint, vl node.Node) []position {
	glyphs := []position{}
	xpos := bag.ScaledPoint(0)
	for e := vl; e != nil; e = e.Next() {
		switch t := e.(type) {
		case *node.VList:
			g := getVPositions(ypos, t.List)
			glyphs = append(glyphs, g...)
		case *node.HList:
			g := getHPositions(ypos, t.List)
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
	return glyphs
}

type retinfo struct {
	positions []position
	height    bag.ScaledPoint
}

func getPositions(settings *node.LinebreakSettings, text string, fontsize bag.ScaledPoint) (*retinfo, error) {
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
	nl, _ = node.AppendLineEndAfter(nl, node.Tail(nl))
	vl, _ := node.Linebreak(nl, settings)
	g := getVPositions(fontsize, vl)

	ri := &retinfo{
		positions: g,
		height:    vl.Height + vl.Depth,
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
		settings.Tolerance = parseFloat(firstArg.Get("tolerance"))

		g, err := getPositions(settings, text, fontsize)
		if err != nil {
			panic("this should not happen")
		}

		x := []any{}

		for _, i := range g.positions {
			obj := map[string]any{
				"char": i.component,
				"xpos": i.xpos.ToPT(),
				"ypos": i.ypos.ToPT(),
			}
			x = append(x, obj)
		}
		ret := map[string]any{
			"positions": x,
			"height":    g.height.ToPT(),
		}
		return ret
	})
	return jsFunc
}

func main() {
	js.Global().Set("getpositions", returnGetPositions())
	<-make(chan bool)
}

package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// The levels pivot swaps two "axes" of a document:
//
//   - the heading axis: heading levels below the skipped ones (e.g. Year > Month > Day)
//   - the list axis: the first N levels of the lists (e.g. the group names)
//
// Headings become nested list items and the top N list levels become
// headings. The heading depth is read from the document; only the list depth
// (N) has to be given. Running it on the output with N = the number of former
// heading levels restores the original.

// aNode is a node on the heading axis, rendered as a nested list item.
type aNode struct {
	name     string
	children []*aNode
	index    map[string]*aNode
	markers  []string // HTML blocks (e.g. <!-- NOW -->) shown under this item
	content  []string // markdown blocks belonging here
}

func newANode(name string) *aNode { return &aNode{name: name, index: map[string]*aNode{}} }

func (a *aNode) child(name string) *aNode {
	if c, ok := a.index[name]; ok {
		return c
	}
	c := newANode(name)
	a.index[name] = c
	a.children = append(a.children, c)
	return c
}

// bNode is a node on the list axis, rendered as a heading.
type bNode struct {
	name     string
	children []*bNode
	index    map[string]*bNode
	markers  []string
	atree    *aNode
}

func newBNode(name string) *bNode {
	return &bNode{name: name, index: map[string]*bNode{}, atree: newANode("")}
}

func (b *bNode) child(name string) *bNode {
	if c, ok := b.index[name]; ok {
		return c
	}
	c := newBNode(name)
	b.index[name] = c
	b.children = append(b.children, c)
	return c
}

func addUnique(list []string, items ...string) []string {
outer:
	for _, it := range items {
		for _, existing := range list {
			if existing == it {
				continue outer
			}
		}
		list = append(list, it)
	}
	return list
}

// section is the content under one of the skipped (pass-through) headings.
type section struct {
	heading  string // full heading line, empty for content before any heading
	preamble []string
	broot    *bNode
}

// rawLines returns the source text covered by a node's own Lines().
func rawLines(n gast.Node, src []byte) string {
	var sb strings.Builder
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		sb.Write(seg.Value(src))
	}
	return strings.TrimSpace(sb.String())
}

// dedent removes the common leading whitespace from all non-blank lines.
func dedent(s string) []string {
	lines := strings.Split(s, "\n")
	min := -1
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		n := len(l) - len(strings.TrimLeft(l, " \t"))
		if min < 0 || n < min {
			min = n
		}
	}
	for i, l := range lines {
		if strings.TrimSpace(l) == "" {
			lines[i] = ""
		} else if min > 0 {
			lines[i] = l[min:]
		}
	}
	return lines
}

// pivotLevels performs the levels pivot. levels is the number of list levels
// that become headings; skip is the number of top heading levels passed
// through unchanged. Dropped content is reported to warn.
func pivotLevels(r io.Reader, w io.Writer, levels, skip int, warn io.Writer) error {
	if levels < 1 {
		return fmt.Errorf("levels must be at least 1, got %d", levels)
	}
	if skip < 0 {
		return fmt.Errorf("skip must not be negative, got %d", skip)
	}
	src, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	doc := goldmark.New().Parser().Parse(text.NewReader(src))

	cur := &section{broot: newBNode("")}
	sections := []*section{cur}
	var apath []string    // current heading-axis path
	var aMarkers []string // HTML blocks under the current heading, not yet attached

	drop := func(n gast.Node, where string) {
		fmt.Fprintf(warn, "warning: dropped content under %q: %s\n",
			where, firstLine(contentSource(n, src)))
	}

	// walkItem records one list item at list depth `depth` (1-based).
	var walkItem func(item gast.Node, depth int, parent *bNode)
	walkItem = func(item gast.Node, depth int, parent *bNode) {
		first := item.FirstChild()
		if first == nil {
			return
		}
		bn := parent.child(rawLines(first, src))

		hasSubList := false
		if depth < levels {
			for c := first.NextSibling(); c != nil; c = c.NextSibling() {
				if _, ok := c.(*gast.List); ok {
					hasSubList = true
				}
			}
		}

		var an *aNode
		if !hasSubList {
			// Leaf on the list axis: content attaches at the heading path.
			an = bn.atree
			for _, p := range apath {
				an = an.child(p)
			}
			an.markers = addUnique(an.markers, aMarkers...)
		}

		for c := first.NextSibling(); c != nil; c = c.NextSibling() {
			switch c := c.(type) {
			case *gast.HTMLBlock:
				bn.markers = addUnique(bn.markers, rawLines(c, src))
			case *gast.List:
				if hasSubList {
					for sub := c.FirstChild(); sub != nil; sub = sub.NextSibling() {
						walkItem(sub, depth+1, bn)
					}
					continue
				}
				if s := contentSource(c, src); s != "" {
					an.content = append(an.content, s)
				}
			default:
				if an != nil {
					if s := contentSource(c, src); s != "" {
						an.content = append(an.content, s)
					}
				} else {
					drop(c, bn.name)
				}
			}
		}
	}

	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		switch b := n.(type) {
		case *gast.Heading:
			aMarkers = nil
			if b.Level <= skip {
				cur = &section{
					heading: strings.Repeat("#", b.Level) + " " + rawLines(b, src),
					broot:   newBNode(""),
				}
				sections = append(sections, cur)
				apath = nil
				continue
			}
			depth := b.Level - skip - 1
			if depth > len(apath) {
				depth = len(apath)
			}
			apath = append(apath[:depth:depth], rawLines(b, src))
		case *gast.List:
			for item := b.FirstChild(); item != nil; item = item.NextSibling() {
				walkItem(item, 1, cur.broot)
			}
			aMarkers = nil
		case *gast.HTMLBlock:
			if len(apath) == 0 {
				cur.preamble = append(cur.preamble, contentSource(b, src))
			} else {
				aMarkers = append(aMarkers, rawLines(b, src))
			}
		default:
			if len(apath) == 0 {
				if s := contentSource(b, src); s != "" {
					cur.preamble = append(cur.preamble, s)
				}
			} else {
				drop(b, apath[len(apath)-1])
			}
		}
	}

	var blocks []string
	for _, s := range sections {
		if s.heading != "" {
			blocks = append(blocks, s.heading)
		}
		blocks = append(blocks, s.preamble...)
		blocks = renderB(blocks, s.broot, skip+1)
	}
	if len(blocks) > 0 {
		fmt.Fprintln(w, strings.Join(blocks, "\n\n"))
	}
	return nil
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i] + " ..."
	}
	return s
}

func renderB(blocks []string, n *bNode, level int) []string {
	for _, c := range n.children {
		blocks = append(blocks, strings.Repeat("#", level)+" "+c.name)
		blocks = append(blocks, c.markers...)
		var sb strings.Builder
		renderA(&sb, c.atree, 0)
		if s := strings.TrimRight(sb.String(), "\n"); s != "" {
			blocks = append(blocks, s)
		}
		blocks = renderB(blocks, c, level+1)
	}
	return blocks
}

func renderA(sb *strings.Builder, n *aNode, depth int) {
	indent := strings.Repeat("  ", depth)
	for _, block := range append(append([]string{}, n.markers...), n.content...) {
		for _, l := range dedent(block) {
			if l == "" {
				sb.WriteString("\n")
			} else {
				sb.WriteString(indent + l + "\n")
			}
		}
	}
	for _, c := range n.children {
		sb.WriteString(indent + "- " + c.name + "\n")
		renderA(sb, c, depth+1)
	}
}

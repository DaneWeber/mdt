package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// pathNode is one level of the heading path (e.g. Year > Month > Day)
// rendered as a nested list item under a group.
type pathNode struct {
	name     string
	children []*pathNode
	index    map[string]*pathNode
	content  []string // markdown blocks (already dedented) belonging here
}

func newPathNode(name string) *pathNode {
	return &pathNode{name: name, index: map[string]*pathNode{}}
}

func (p *pathNode) child(name string) *pathNode {
	if c, ok := p.index[name]; ok {
		return c
	}
	c := newPathNode(name)
	p.index[name] = c
	p.children = append(p.children, c)
	return c
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

// pivotByGroup converts a document whose groups are top-level list items
// under a heading path into one H2 section per group, with the heading path
// rendered as a nested list:
//
//	# Title             # Title
//	## Year             ## Group A
//	### Day       =>    - Year
//	- Group A             - Day
//	  - did X               - did X
func pivotByGroup(r io.Reader, w io.Writer) error {
	src, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	doc := goldmark.New().Parser().Parse(text.NewReader(src))

	var title string
	var path []string // heading texts for levels 2..6
	var groupOrder []string
	groups := map[string]*pathNode{}
	// HTML blocks (e.g. <!-- NOW -->) seen under the current heading path,
	// attached to every group's node for that path.
	var pendingHTML []string

	nodeFor := func(group string) *pathNode {
		root, ok := groups[group]
		if !ok {
			root = newPathNode(group)
			groups[group] = root
			groupOrder = append(groupOrder, group)
		}
		n := root
		for _, p := range path {
			n = n.child(p)
		}
		return n
	}

	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		switch b := n.(type) {
		case *gast.Heading:
			pendingHTML = nil
			if b.Level == 1 {
				if title == "" {
					title = headingText(b, src)
				}
				path = nil
				continue
			}
			depth := b.Level - 2
			if depth > len(path) {
				depth = len(path)
			}
			path = append(path[:depth:depth], headingText(b, src))
		case *gast.HTMLBlock:
			pendingHTML = append(pendingHTML, rawLines(b, src))
		case *gast.List:
			for item := b.FirstChild(); item != nil; item = item.NextSibling() {
				first := item.FirstChild()
				if first == nil {
					continue
				}
				group := rawLines(first, src)
				pn := nodeFor(group)
				pn.content = append(pn.content, pendingHTML...)
				for c := first.NextSibling(); c != nil; c = c.NextSibling() {
					if s := contentSource(c, src); s != "" {
						pn.content = append(pn.content, s)
					}
				}
			}
			pendingHTML = nil
		}
	}

	if title != "" {
		fmt.Fprintf(w, "# %s\n", title)
	}
	for i, g := range groupOrder {
		if title != "" || i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "## %s\n\n", g)
		writePathTree(w, groups[g], 0)
	}
	return nil
}

func writePathTree(w io.Writer, n *pathNode, depth int) {
	indent := strings.Repeat("  ", depth)
	for _, block := range n.content {
		for _, l := range dedent(block) {
			if l == "" {
				fmt.Fprintln(w)
			} else {
				fmt.Fprintf(w, "%s%s\n", indent, l)
			}
		}
	}
	for _, c := range n.children {
		fmt.Fprintf(w, "%s- %s\n", indent, c.name)
		writePathTree(w, c, depth+1)
	}
}

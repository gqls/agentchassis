package main

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// A deliberately tiny selector engine: `tag`, `.class`, `#id`, and `tag.class`.
// That is every shape the two source sites' chrome actually uses, and it keeps
// the strip rules readable in overrides.json. Anything more expressive would be
// a signal to fix the source markup instead.

type selector struct {
	tag   string
	class string
	id    string
}

func parseSelector(s string) selector {
	var sel selector
	s = strings.TrimSpace(s)
	if i := strings.IndexAny(s, ".#"); i >= 0 {
		sel.tag = s[:i]
		rest := s[i:]
		if strings.HasPrefix(rest, ".") {
			sel.class = strings.TrimPrefix(rest, ".")
		} else {
			sel.id = strings.TrimPrefix(rest, "#")
		}
	} else {
		sel.tag = s
	}
	return sel
}

func (sel selector) matches(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}
	if sel.tag != "" && !strings.EqualFold(n.Data, sel.tag) {
		return false
	}
	if sel.class != "" && !hasClass(n, sel.class) {
		return false
	}
	if sel.id != "" && attr(n, "id") != sel.id {
		return false
	}
	return sel.tag != "" || sel.class != "" || sel.id != ""
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return a.Val
		}
	}
	return ""
}

func setAttr(n *html.Node, key, val string) {
	for i := range n.Attr {
		if strings.EqualFold(n.Attr[i].Key, key) {
			n.Attr[i].Val = val
			return
		}
	}
	n.Attr = append(n.Attr, html.Attribute{Key: key, Val: val})
}

func hasClass(n *html.Node, class string) bool {
	for _, c := range strings.Fields(attr(n, "class")) {
		if c == class {
			return true
		}
	}
	return false
}

// findAll walks the tree and returns every node satisfying pred, in document
// order. Collected before any mutation so callers can safely remove them.
func findAll(root *html.Node, pred func(*html.Node) bool) []*html.Node {
	var out []*html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if pred(n) {
			out = append(out, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(root)
	return out
}

func findFirst(root *html.Node, pred func(*html.Node) bool) *html.Node {
	if all := findAll(root, pred); len(all) > 0 {
		return all[0]
	}
	return nil
}

func byAtom(a atom.Atom) func(*html.Node) bool {
	return func(n *html.Node) bool { return n.Type == html.ElementNode && n.DataAtom == a }
}

func remove(n *html.Node) {
	if n.Parent != nil {
		n.Parent.RemoveChild(n)
	}
}

// renderChildren serialises a node's children (not the node itself). Used to
// unwrap <main> into a fragment without carrying the <main> element over —
// chassis chrome supplies the page's own structural wrapper.
func renderChildren(n *html.Node) string {
	var buf bytes.Buffer
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		_ = html.Render(&buf, c)
	}
	return buf.String()
}

func renderNode(n *html.Node) string {
	var buf bytes.Buffer
	_ = html.Render(&buf, n)
	return buf.String()
}

// parseFragment parses an HTML fragment and returns the node whose children are
// the fragment's top-level nodes.
//
// The wrapping <div> is not cosmetic: html.Parse builds a full document, and
// bare fragment content would be relocated by the tree-construction rules (a
// stray <td>, for instance, does not survive at body level). Wrapping keeps the
// content where it was written.
//
// The subtlety that bit us: the wrapper must be UNWRAPPED on the way out. The
// first version returned renderChildren(body), which includes the wrapper div
// itself — so every pass through this helper added one more <div> layer, and a
// fragment that went through extractScripts, sweepInlineStyles and rewriteLinks
// came out triple-wrapped. Returning the wrapper node (whose children are the
// real content) makes the round-trip lossless.
func parseFragment(fragment string) (*html.Node, error) {
	doc, err := html.Parse(strings.NewReader("<div>" + fragment + "</div>"))
	if err != nil {
		return nil, err
	}
	body := findFirst(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.DataAtom == atom.Body
	})
	if body == nil || body.FirstChild == nil {
		return nil, fmt.Errorf("fragment lost its wrapper")
	}
	return body.FirstChild, nil
}

// visibleTextChars reimplements the assembly's sectionHasVisibleContent metric
// (rerender_single_page_action.go): text with <script> and <style> subtrees
// removed and whitespace collapsed. A section at or below 10 chars is silently
// dropped from the assembled page, so the transform must measure it and fail
// loudly rather than let a tool page vanish at publish time.
func visibleTextChars(fragment string) int {
	doc, err := html.Parse(strings.NewReader(fragment))
	if err != nil {
		return 0
	}
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && (n.DataAtom == atom.Script || n.DataAtom == atom.Style) {
			return
		}
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return len(strings.Join(strings.Fields(sb.String()), " "))
}

// textOf returns the collapsed text content of a node.
func textOf(n *html.Node) string {
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(x *html.Node) {
		if x.Type == html.TextNode {
			sb.WriteString(x.Data)
		}
		for c := x.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.Join(strings.Fields(sb.String()), " ")
}

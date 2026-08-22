// FILE: cmd/content-data-recover/matcher.go
//
// The inverter: match a parsed component template against the HTML it produced,
// binding each {{.field}} to the bytes that stand where it stood.
//
// It is a BACKTRACKING matcher in continuation-passing style, not a compiled
// regex, for two reasons. RE2 (Go's regexp) has no backreferences and returns
// only the LAST match of a repeated group, so `{{range}}` bodies cannot be
// recovered from one pass. And an {{if}}/{{else}} pair needs "try this branch,
// and if the REST of the template then fails, try the other" — which is
// backtracking, not alternation.
//
// Every binding here is a GUESS until main.go's round-trip proves it. That
// division is deliberate: this file may be as clever or as wrong as it likes;
// nothing it produces is written unless re-rendering reproduces the stored bytes
// exactly. So the correctness bar for this file is only that it not be
// EXPENSIVE, and the guard against pathological backtracking is anchorHints.
package main

import (
	"strings"
	"text/template/parse"
)

// binding accumulates recovered values. Keys are top-level content_data keys;
// a range binds a []interface{} of child maps (or of strings for {{.}} bodies).
type binding map[string]interface{}

func (b binding) clone() binding {
	out := make(binding, len(b))
	for k, v := range b {
		out[k] = v
	}
	return out
}

func (b binding) toData() map[string]interface{} {
	out := make(map[string]interface{}, len(b))
	for k, v := range b {
		if k == "" {
			continue
		}
		out[k] = v
	}
	return out
}

type matcher struct {
	in    string
	steps int // pathological-backtracking guard
}

const maxSteps = 2_000_000

// matchList matches nodes[i:] starting at pos, calling next(endPos) when the
// list is exhausted. next reports whether the REST of the match (the caller's
// continuation) succeeded, which is what makes backtracking work.
func (m *matcher) matchList(nodes []parse.Node, pos int, b binding) (int, bool) {
	end := -1
	ok := m.walk(nodes, 0, pos, b, func(p int) bool { end = p; return true })
	return end, ok
}

func (m *matcher) walk(nodes []parse.Node, i, pos int, b binding, next func(int) bool) bool {
	m.steps++
	if m.steps > maxSteps {
		return false
	}
	if i >= len(nodes) {
		return next(pos)
	}
	cont := func(p int) bool { return m.walk(nodes, i+1, p, b, next) }

	switch t := nodes[i].(type) {
	case *parse.TextNode:
		lit := string(t.Text)
		if !strings.HasPrefix(m.in[pos:], lit) {
			return false
		}
		return cont(pos + len(lit))

	case *parse.ActionNode:
		name, simple := simpleField(t.Pipe)
		if !simple {
			// A pipeline this tool cannot attribute to ONE field (e.g.
			// {{or .a .b}}): refuse rather than guess which field held the
			// value. A wrong attribution would still round-trip — both render
			// identically — so the gate could not catch it, which is exactly
			// why the refusal has to live here.
			return false
		}
		for _, e := range m.candidateEnds(nodes, i, pos) {
			val := m.in[pos:e]
			saved, had := b[name]
			if had && saved != val {
				continue // same field bound twice with different bytes
			}
			b[name] = val
			if cont(e) {
				return true
			}
			if had {
				b[name] = saved
			} else {
				delete(b, name)
			}
		}
		return false

	case *parse.IfNode:
		// then-branch first: taking it means the guarded fields were present.
		snap := b.clone()
		if m.walk(t.List.Nodes, 0, pos, b, cont) {
			return true
		}
		restore(b, snap)
		if t.ElseList != nil {
			if m.walk(t.ElseList.Nodes, 0, pos, b, cont) {
				return true
			}
			restore(b, snap)
			return false
		}
		// no else: the branch simply contributes nothing
		return cont(pos)

	case *parse.RangeNode:
		name, simple := simpleField(t.Pipe)
		if !simple {
			return false
		}
		return m.matchRange(t, name, pos, b, cont)

	case *parse.CommentNode:
		return cont(pos)

	default:
		// {{with}}, {{template}}, {{block}} and friends: not attempted.
		return false
	}
}

// matchRange tries 0,1,2… iterations of the body, accumulating one child
// binding per iteration. Longest-first would be marginally faster on lists but
// shortest-first terminates sooner on the common empty case.
func (m *matcher) matchRange(t *parse.RangeNode, name string, pos int, b binding, next func(int) bool) bool {
	var try func(p int, acc []interface{}) bool
	try = func(p int, acc []interface{}) bool {
		snap := b.clone()
		// Option 1 — stop iterating here.
		if len(acc) > 0 {
			b[name] = acc
			if next(p) {
				return true
			}
			restore(b, snap)
		} else if t.ElseList != nil {
			if m.walk(t.ElseList.Nodes, 0, p, b, next) {
				return true
			}
			restore(b, snap)
		} else if next(p) {
			return true
		}
		// Option 2 — consume one more iteration of the body.
		ok := false
		child := binding{}
		m.walk(t.List.Nodes, 0, p, child, func(q int) bool {
			if q == p {
				return false // zero-width iteration would not terminate
			}
			grown := make([]interface{}, len(acc)+1)
			copy(grown, acc)
			grown[len(acc)] = child.element()
			if try(q, grown) {
				ok = true
				return true
			}
			return false
		})
		return ok
	}
	return try(pos, nil)
}

// element renders one range iteration's binding as a template value: a map when
// the body bound named fields, a plain string when it bound only {{.}}.
func (b binding) element() interface{} {
	if len(b) == 1 {
		if v, ok := b[""]; ok {
			return v
		}
	}
	out := map[string]interface{}{}
	for k, v := range b {
		if k != "" {
			out[k] = v
		}
	}
	return out
}

// candidateEnds returns plausible end offsets for a value starting at pos. When
// the next node is literal text, only the positions where that literal actually
// occurs can be right — which turns an O(n) scan per action into a handful of
// candidates and is what keeps nested actions from exploding.
func (m *matcher) candidateEnds(nodes []parse.Node, i, pos int) []int {
	if i+1 < len(nodes) {
		if tn, ok := nodes[i+1].(*parse.TextNode); ok && len(tn.Text) > 0 {
			lit := string(tn.Text)
			var out []int
			for off := pos; ; {
				idx := strings.Index(m.in[off:], lit)
				if idx < 0 {
					break
				}
				out = append(out, off+idx)
				off = off + idx + 1
				if off > len(m.in) {
					break
				}
			}
			return out
		}
	}
	// No literal anchor: a value may end anywhere. Bounded to keep the search
	// finite; templates in this population always have an anchoring literal.
	out := make([]int, 0, len(m.in)-pos+1)
	for e := pos; e <= len(m.in); e++ {
		out = append(out, e)
	}
	return out
}

func simpleField(p *parse.PipeNode) (string, bool) {
	if p == nil || len(p.Cmds) != 1 {
		return "", false
	}
	args := p.Cmds[0].Args
	if len(args) != 1 {
		return "", false
	}
	switch a := args[0].(type) {
	case *parse.FieldNode:
		if len(a.Ident) != 1 {
			return "", false // .a.b — nested, not attempted
		}
		return a.Ident[0], true
	case *parse.DotNode:
		return "", true // {{.}} inside a range body
	}
	return "", false
}

func restore(b, snap binding) {
	for k := range b {
		delete(b, k)
	}
	for k, v := range snap {
		b[k] = v
	}
}

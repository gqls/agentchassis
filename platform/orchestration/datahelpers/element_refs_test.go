// FILE: platform/orchestration/datahelpers/element_refs_test.go
//
// The false-positive guards are the point of these tests. A wrong finding here
// dispatches a fixer at a working tool, so every conservatism in
// OrphanElementRefs is pinned by a case that would fail without it.

package datahelpers

import (
	"reflect"
	"testing"
)

func TestOrphanElementRefs_LiveDefects(t *testing.T) {
	// Shapes taken from the two live pages this check was written against
	// (webdesign.co.uk, 2026-07-29). Both were serving 200 with a visitor-visible
	// tool that could not work.
	tests := []struct {
		name string
		html string
		want []string
	}{
		{
			// tool-insight-injector: the port dropped the whole left-hand input
			// panel; the script still addressed all five of its fields.
			name: "port dropped the input panel",
			html: `<div class="output-panel"><div id="output"></div></div>
			       <script>
			         const name = document.getElementById('biz-name').value;
			         const fact = document.getElementById('biz-fact').value;
			         document.getElementById('output').innerHTML = name + fact;
			       </script>`,
			want: []string{"biz-fact", "biz-name"},
		},
		{
			// tool-monolith-splitter: the whole input side was never ported, so
			// the page renders its own copy about an interactive tool above a
			// panel with nothing in it.
			name: "input side never ported",
			html: `<div class="output" id="results"></div>
			       <script>
			         const file = document.getElementById('target-file').value;
			         const fw = document.getElementById('framework').value;
			         document.getElementById('results').textContent = file + fw;
			       </script>`,
			want: []string{"framework", "target-file"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := OrphanElementRefs(tc.html)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("OrphanElementRefs() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestOrphanElementRefs_NoFalsePositives(t *testing.T) {
	tests := []struct {
		name string
		html string
	}{
		{
			name: "element is present in the markup",
			html: `<input id="qty"><script>document.getElementById('qty').value = 1;</script>`,
		},
		{
			// The guard that matters most in practice: a tool that renders its
			// own controls. The id exists only inside a string literal, and the
			// harvest deliberately reads string literals as present.
			name: "element is built by the script itself",
			html: `<div id="host"></div>
			       <script>
			         document.getElementById('host').innerHTML = '<input id="qty">';
			         document.getElementById('qty').focus();
			       </script>`,
		},
		{
			name: "id assigned dynamically",
			html: `<script>
			         const el = document.createElement('div');
			         el.id = 'panel';
			         document.body.appendChild(el);
			         document.getElementById('panel').textContent = 'hi';
			       </script>`,
		},
		{
			name: "id set via setAttribute",
			html: `<script>
			         const el = document.createElement('div');
			         el.setAttribute('id', 'panel');
			         document.getElementById('panel').textContent = 'hi';
			       </script>`,
		},
		{
			// The caller passes the whole page, so a script reaching into the
			// site chrome is legitimate and must not be flagged.
			name: "element lives in the site chrome",
			html: `<header><button id="nav-toggle"></button></header>
			       <main><script>document.getElementById('nav-toggle').onclick = f;</script></main>`,
		},
		{
			// A compound selector is ignored rather than half-parsed: we cannot
			// tell from '#a .b' alone which part is expected to exist.
			name: "compound selector is not judged",
			html: `<script>document.querySelector('#missing .child');</script>`,
		},
		{
			name: "single quotes and double quotes both count as present",
			html: `<input id='qty'><script>document.getElementById("qty").value = 1;</script>`,
		},
		{
			// THE REGRESSION CASE. tool-css-filter-playground, verbatim in
			// shape: six sliders generated from an array of descriptors, so no
			// slider id appears in the source and all six resolve in a browser.
			// The first version of this check reported all six as missing, and
			// the wrong verdict reached a commit message, a register entry and
			// a live council submission before the browser caught it.
			name: "ids interpolated from data are present",
			html: `<div id="sliders"></div>
			       <script>
			         const filters = [{name: 'brightness', val: 100}, {name: 'contrast', val: 100}];
			         filters.forEach(function (f) {
			           const g = document.createElement('div');
			           g.innerHTML = ` + "`" + `<input type="range" id="${f.name}" value="${f.val}">` + "`" + `;
			           document.getElementById('sliders').appendChild(g);
			         });
			         document.getElementById('brightness').value = 90;
			         document.getElementById('contrast').value = 90;
			       </script>`,
		},
		{
			name: "no script at all",
			html: `<div id="a"></div><p>Just content.</p>`,
		},
		{
			name: "script with no element references",
			html: `<script>console.log('hello');</script>`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := OrphanElementRefs(tc.html); len(got) != 0 {
				t.Fatalf("OrphanElementRefs() = %v, want no findings", got)
			}
		})
	}
}

// The dynamic-id rule must LOOSEN the test, not disable it. Four of the nine
// real findings are on tools that build markup dynamically AND are genuinely
// broken (asset-formatter, logic-architect, mind-map, pasteboard), so a page
// that interpolates ids still has to be judged on the ids it never mentions.
func TestOrphanElementRefs_DynamicPageStillJudged(t *testing.T) {
	// tool-pasteboard in shape: renders note cards with computed ids, while
	// btnUndo and boardTitle are addressed and appear nowhere else at all.
	html := `<div id="board"></div>
	         <script>
	           notes.forEach(function (n) {
	             board.innerHTML += ` + "`" + `<div class="note" id="note-${n.key}">${n.text}</div>` + "`" + `;
	           });
	           document.getElementById('btnUndo').addEventListener('click', undo);
	           document.getElementById('boardTitle').value = 'Untitled';
	         </script>`
	want := []string{"boardTitle", "btnUndo"}
	if got := OrphanElementRefs(html); !reflect.DeepEqual(got, want) {
		t.Fatalf("OrphanElementRefs() = %v, want %v — the dynamic-id rule must not "+
			"blind the check on a page that computes SOME of its ids", got, want)
	}
}

func TestOrphanElementRefs_DeduplicatesAndSorts(t *testing.T) {
	html := `<script>
	           document.getElementById('zebra');
	           document.getElementById('zebra');
	           document.querySelector('#apple');
	         </script>`
	want := []string{"apple", "zebra"}
	if got := OrphanElementRefs(html); !reflect.DeepEqual(got, want) {
		t.Fatalf("OrphanElementRefs() = %v, want %v", got, want)
	}
}

package actions

import "testing"

// Tests for DropDeadURLControls (bugs_open/054). The function removes site-chrome
// controls whose URL attribute rendered empty — the dead-control class that made
// idea.uk ship 30 empty-href nav links (bugs_open/018). Callers gate it on a
// non-empty deadURLFields set, so these tests exercise it only on inputs that
// contain (or deliberately do not contain) a dead control.
func TestDropDeadURLControls(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		// --- positive: the dead control is dropped -------------------------------
		{
			name: "bare dead nav anchor is removed whole",
			in:   `<nav><a href="/about.html">About</a><a href="">Contact</a></nav>`,
			want: `<nav><a href="/about.html">About</a></nav>`,
		},
		{
			name: "dead CTA anchor with classes is removed, live sibling kept",
			in:   `<a href="/tools.html" class="nav">Tools</a><a href="" class="header-cta-btn">Get started</a>`,
			want: `<a href="/tools.html" class="nav">Tools</a>`,
		},
		{
			name: "href empty comes before other attributes",
			in:   `<a href="" class="header-cta">Go</a>`,
			want: ``,
		},
		{
			name: "single-quoted empty href",
			in:   `<a href='' class='x'>Dead</a>`,
			want: ``,
		},
		{
			name: "whitespace around equals",
			in:   `<a  href = "" >Dead</a>`,
			want: ``,
		},
		{
			name: "multiline anchor body",
			in:   "<a href=\"\">\n  <span>Dead</span>\n</a>",
			want: ``,
		},
		{
			name: "dead img (empty src) is dropped whole",
			in:   `<img class="logo" src="" alt="Logo">`,
			want: ``,
		},
		{
			name: "empty src on a non-img element is stripped, element kept",
			in:   `<source src="" type="video/mp4">`,
			want: `<source type="video/mp4">`,
		},
		{
			name: "multiple dead anchors all removed",
			in:   `<a href="">A</a><a href="/x">B</a><a href="">C</a>`,
			want: `<a href="/x">B</a>`,
		},
		{
			name: "realistic idea.uk header shape",
			in:   `<header><a class="logo" href="/index.html"><img src=""></a><a href="">Home</a><a href="" class="header-cta-btn">Contact</a></header>`,
			want: `<header><a class="logo" href="/index.html"></a></header>`,
		},

		// --- negative: nothing legitimate is touched -----------------------------
		{
			name: "non-empty href untouched",
			in:   `<a href="/services.html">Services</a>`,
			want: `<a href="/services.html">Services</a>`,
		},
		{
			name: "fragment href (nav/JS toggle) untouched",
			in:   `<a href="#" class="menu-toggle">Menu</a>`,
			want: `<a href="#" class="menu-toggle">Menu</a>`,
		},
		{
			name: "unrelated empty attribute untouched",
			in:   `<a href="/x" data-track="">Keep</a>`,
			want: `<a href="/x" data-track="">Keep</a>`,
		},
		{
			name: "area with empty href is out of scope (word boundary)",
			in:   `<area href="" shape="rect">`,
			want: `<area href="" shape="rect">`,
		},
		{
			name: "abbr tag not matched by <a boundary",
			in:   `<abbr title="">x</abbr>`,
			want: `<abbr title="">x</abbr>`,
		},
		{
			name: "non-empty src untouched",
			in:   `<img src="/assets/logo.png" alt="Logo">`,
			want: `<img src="/assets/logo.png" alt="Logo">`,
		},
		{
			name: "clean chrome unchanged",
			in:   `<nav><a href="/a">A</a><a href="/b">B</a></nav>`,
			want: `<nav><a href="/a">A</a><a href="/b">B</a></nav>`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := DropDeadURLControls(tc.in)
			if got != tc.want {
				t.Errorf("DropDeadURLControls(%q)\n  got:  %q\n  want: %q", tc.in, got, tc.want)
			}
		})
	}
}

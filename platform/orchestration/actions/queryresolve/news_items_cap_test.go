package queryresolve

import (
	"testing"
)

func titlesOf(items []NewsItem) []string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.Title)
	}
	return out
}

func itemsFromTitles(titles ...string) []NewsItem {
	out := make([]NewsItem, 0, len(titles))
	for _, t := range titles {
		out = append(out, NewsItem{Title: t})
	}
	return out
}

// webdesign.co.uk's live source queries, verbatim — the topical derivation the
// production path feeds the cap on that site.
var webdesignQueries = []string{
	"web accessibility WCAG UK regulations",
	"brand identity design studio rebrand",
	"CSS new features browser support",
	"AI web design UX interface designers",
	"typeface release design system open source",
}

// The two live cases that motivated D14's cap, verbatim from the
// webdesign.co.uk feed of 2026-07-29: one rebrand across five outlets, one
// browser release family across four headlines.
func TestCapNewsItemsPerToolObservedCases(t *testing.T) {
	t.Run("coca-cola rebrand, five outlets to two", func(t *testing.T) {
		items := itemsFromTitles(
			"Coca-Cola Reached for Nostalgia. Did It Get Marlboro Instead?",
			"Coca-Cola unveils visual rebrand for consistency across packaging",
			"Coca-Cola Unveils a New Global Brand Identity",
			"Coca-Cola gets new look that makes it even \"more Coca-Cola\"",
			"Coca-Cola’s new global identity: ‘make Coca-Cola more Coca-Cola’",
			"Creative Dive: A Look at Duedance’s Brand Identity for Yonbek 1227",
		)
		topical := newsTopicalTokens(webdesignQueries, items)
		got := capNewsItemsPerTool(items, 2, topical)
		if len(got) != 3 {
			t.Fatalf("want 2 coca-cola + 1 unrelated = 3, got %d: %v", len(got), titlesOf(got))
		}
		if got[0].Title != items[0].Title || got[1].Title != items[1].Title {
			t.Fatalf("cap must keep the two highest-ranked, got %v", titlesOf(got[:2]))
		}
		if got[2].Title != items[5].Title {
			t.Fatalf("unrelated item must survive, got %v", titlesOf(got))
		}
	})

	t.Run("firefox family, four headlines to two", func(t *testing.T) {
		items := itemsFromTitles(
			"Firefox 153 Officially Released: QR Code Generation, PDF Merging, and Tighter Extension File Restrictions",
			"Firefox 153 Released with HDR Video, Smarter PDF Tools, Better Privacy, and New Linux Improvements",
			"Mozilla Firefox 154 Enters Public Beta Testing, Here’s What to Expect",
			"Firefox 154 open-source web browser is now available for public beta testing",
			"Safari Technology Preview 248: BigInt Math Arrives, CSP Enforcement Tightened",
			"Google Chrome 150 stable version released with new CSS features",
		)
		topical := newsTopicalTokens(webdesignQueries, items)
		got := capNewsItemsPerTool(items, 2, topical)
		var firefox int
		for _, it := range got {
			for _, k := range titleToolKeys(it.Title, topical) {
				if k == "firefox" {
					firefox++
				}
			}
		}
		if firefox != 2 {
			t.Fatalf("want exactly 2 firefox items, got %d: %v", firefox, titlesOf(got))
		}
		var safari, chrome bool
		for _, it := range got {
			switch it.Title {
			case items[4].Title:
				safari = true
			case items[5].Title:
				chrome = true
			}
		}
		if !safari || !chrome {
			t.Fatalf("distinct-tool items must survive (safari=%v chrome=%v): %v", safari, chrome, titlesOf(got))
		}
	})
}

// A site's subject vocabulary must not act as tool keys. Two derivations:
// query words (CSS is in webdesign's own query) and pool frequency (Oil is in
// most of gaswholesalers' titles though its query only says gas).
func TestCapNewsItemsPerToolTopicsAreNotTools(t *testing.T) {
	t.Run("query-derived: CSS mentions are not one tool", func(t *testing.T) {
		items := itemsFromTitles(
			"HTML and CSS in Emails: What Works in 2026?",
			"Google Chrome 150 stable version released, enabling gradient borders using only CSS",
			"Safari Technology Preview 248 ships a dozen CSS fixes",
			"Ubuntu Touch 24.04-2.0 Released with Browser Upgrade",
		)
		topical := newsTopicalTokens(webdesignQueries, items)
		got := capNewsItemsPerTool(items, 2, topical)
		if len(got) != len(items) {
			t.Fatalf("no drops expected — CSS/browser are query topics; got %d of %d: %v",
				len(got), len(items), titlesOf(got))
		}
	})

	t.Run("frequency-derived: a commodity word across the pool is a topic", func(t *testing.T) {
		items := itemsFromTitles(
			"Oil Prices Surge 6% Even as Tankers Push Through Hormuz",
			"Oil Tops $100 as Supply Crisis Deepens",
			"Oil Prices Climb as Conflict Shows No Signs of Slowing",
			"Oil and LNG Tankers Make U-Turns After Fresh Escalation",
			"Oil Prices Set for Biggest Weekly Surge Since April",
			"Oil hits $100 after attack worsens supply disruption",
			"Russia's Biggest Black Sea Oil Port Goes Quiet as Drone Threat Grows",
			"OPEC Oil Production Jumps, But Gulf Supply Is Still Far From Normal",
			"India Pulls Back From Iraqi Oil as Route Turns Too Dangerous",
			"Oil prices stable as peace efforts hold",
			"Ukraine's Drone War Is Choking Kazakhstan's Oil Exports",
			"The Red Sea Is Becoming a Major Oil Bottleneck",
		)
		// gas site's query does NOT contain "oil" — frequency must catch it
		topical := newsTopicalTokens([]string{"UK wholesale gas prices news"}, items)
		if !topical["oil"] {
			t.Fatalf("oil appears in 11 of 12 titles and must be topical; topical=%v", topical)
		}
		got := capNewsItemsPerTool(items, 2, topical)
		if len(got) < 10 {
			t.Fatalf("distinct oil-market stories must not collapse to 2; got %d of %d: %v",
				len(got), len(items), titlesOf(got))
		}
	})
}

func TestCapNewsItemsPerToolBoundsAndOrder(t *testing.T) {
	t.Run("cap disabled passes through", func(t *testing.T) {
		items := itemsFromTitles("Figma A1", "Figma B2", "Figma C3")
		if got := capNewsItemsPerTool(items, 0, nil); len(got) != 3 {
			t.Fatalf("cap<=0 must pass through, got %d", len(got))
		}
	})
	t.Run("order preserved and third same-tool item dropped", func(t *testing.T) {
		items := itemsFromTitles(
			"Figma introduces variables",
			"Penpot ships components",
			"Figma acquires a plugin studio",
			"Figma opens its schema",
		)
		got := capNewsItemsPerTool(items, 2, nil)
		want := []string{items[0].Title, items[1].Title, items[2].Title}
		if len(got) != 3 {
			t.Fatalf("want 3, got %d: %v", len(got), titlesOf(got))
		}
		for i := range want {
			if got[i].Title != want[i] {
				t.Fatalf("order not preserved at %d: %v", i, titlesOf(got))
			}
		}
	})
}

func TestTitleToolKeys(t *testing.T) {
	topical := newsTopicalTokens(webdesignQueries, nil)
	cases := []struct {
		title string
		want  []string
	}{
		// hyphenated brand is ONE key; possessive is stripped
		{"Coca-Cola’s new global identity", []string{"coca-cola"}},
		// query-topical and stopword tokens are not keys
		{"CSS Grid Update Brings New Browser Support", []string{"grid"}},
		// lowercase tokens are never keys
		{"the quiet death of the personal homepage", nil},
	}
	for _, c := range cases {
		got := titleToolKeys(c.title, topical)
		if len(got) != len(c.want) {
			t.Fatalf("%q: want %v, got %v", c.title, c.want, got)
		}
		for i := range c.want {
			if got[i] != c.want[i] {
				t.Fatalf("%q: want %v, got %v", c.title, c.want, got)
			}
		}
	}
}

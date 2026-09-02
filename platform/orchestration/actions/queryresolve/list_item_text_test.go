package queryresolve

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestListItemTitleStripsOnlyTheTrailingSiteSuffix(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"the live boxingonline case",
			"Cruiserweight Is Boxing's Best-Kept Secret — And It Won't Stay That Way | Boxing Online",
			"Cruiserweight Is Boxing's Best-Kept Secret — And It Won't Stay That Way"},
		{"a headline that itself contains the separator keeps its own",
			"Rules, Scoring | What Changed | Boxing Online",
			"Rules, Scoring | What Changed"},
		{"no suffix is left alone",
			"Flyweight Feature", "Flyweight Feature"},
		{"a title that is only a suffix is left alone rather than blanked",
			" | Boxing Online", " | Boxing Online"},
		{"empty stays empty", "", ""},
		{"a pipe with no spaces is not the separator",
			"Boxing|Online", "Boxing|Online"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ListItemTitle(c.in); got != c.want {
				t.Fatalf("ListItemTitle(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestListItemExcerptPassesShortTextThrough(t *testing.T) {
	in := "Discover why cruiserweight boxing sits between light heavyweight and heavyweight."
	if got := ListItemExcerpt(in); got != in {
		t.Fatalf("short deck was altered: %q", got)
	}
}

func TestListItemExcerptBoundsLongText(t *testing.T) {
	in := strings.Repeat("a", 500)
	got := ListItemExcerpt(in)
	if n := len(got); n > ListItemExcerptMaxBytes+3 {
		t.Fatalf("bounded deck is %d bytes, want <= %d", n, ListItemExcerptMaxBytes+3)
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("bounded deck does not end in an ellipsis: %q", got)
	}
}

// The failure this helper exists to make unrepresentable: a byte-slice
// truncation cuts a multi-byte character in half, and the invalid UTF-8 that
// results is what Postgres refused in bugs_open/423.
//
// The control is the point — the same input truncated by BYTE must fail this
// assertion, or the test proves nothing about the rune path.
func TestListItemExcerptNeverSplitsAMultiByteCharacter(t *testing.T) {
	// Em-dashes are 3 bytes each and routine in this estate's meta descriptions.
	in := strings.Repeat("—", 400)

	got := ListItemExcerpt(in)
	if !utf8.ValidString(got) {
		t.Fatalf("ListItemExcerpt produced invalid UTF-8")
	}
	if len(got) > ListItemExcerptMaxBytes+3 {
		t.Fatalf("got %d bytes, want <= %d", len(got), ListItemExcerptMaxBytes+3)
	}

	// Control: the byte-slice rule this replaced DOES corrupt this input, so a
	// pass above is discriminating rather than vacuous.
	if len(in) > 200 {
		byteSliced := in[:197] + "..."
		if utf8.ValidString(byteSliced) {
			t.Fatalf("control failed: the byte-slice truncation produced valid UTF-8, " +
				"so this test cannot tell the rune path from the byte path")
		}
	}
}

// Both producers of the standard list-item shape must apply the SAME rules.
// Pinned by behaviour rather than by comment: rebuild_blog_listing's
// scanBlogArticles and resolvePagesWhereType each derive the article set for
// one site's cards, and until 2026-09-02 they disagreed about both fields.
func TestListItemHelpersAreIdempotent(t *testing.T) {
	// A card that has already been through the projection must survive a second
	// pass unchanged — the two producers can both run against one listing.
	title := ListItemTitle("Women's Boxing Is Having a Moment | Boxing Online")
	if again := ListItemTitle(title); again != title {
		t.Fatalf("ListItemTitle is not idempotent: %q then %q", title, again)
	}
	deck := ListItemExcerpt(strings.Repeat("é", 400))
	if again := ListItemExcerpt(deck); again != deck {
		t.Fatalf("ListItemExcerpt is not idempotent")
	}
}

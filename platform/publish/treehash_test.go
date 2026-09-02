// FILE: platform/publish/treehash_test.go
package publish

import "testing"

func TestTreeHashIsOrderIndependent(t *testing.T) {
	a := []File{{Key: "index.html", ETag: "aa", Size: 10}, {Key: "css/site.css", ETag: "bb", Size: 20}}
	b := []File{{Key: "css/site.css", ETag: "bb", Size: 20}, {Key: "index.html", ETag: "aa", Size: 10}}
	if TreeHash(a) != TreeHash(b) {
		t.Fatalf("TreeHash must not depend on listing order: %s != %s", TreeHash(a), TreeHash(b))
	}
}

func TestTreeHashSeesEveryKindOfDrift(t *testing.T) {
	base := []File{{Key: "index.html", ETag: "aa", Size: 10}, {Key: "about.html", ETag: "bb", Size: 20}}
	h := TreeHash(base)

	cases := map[string][]File{
		"content change (etag)": {{Key: "index.html", ETag: "CHANGED", Size: 10}, {Key: "about.html", ETag: "bb", Size: 20}},
		"file removed":          {{Key: "index.html", ETag: "aa", Size: 10}},
		"file added":            {{Key: "index.html", ETag: "aa", Size: 10}, {Key: "about.html", ETag: "bb", Size: 20}, {Key: "new.html", ETag: "cc", Size: 5}},
		"file renamed":          {{Key: "index2.html", ETag: "aa", Size: 10}, {Key: "about.html", ETag: "bb", Size: 20}},
		"size change":           {{Key: "index.html", ETag: "aa", Size: 11}, {Key: "about.html", ETag: "bb", Size: 20}},
	}
	for name, files := range cases {
		if TreeHash(files) == h {
			t.Errorf("%s: TreeHash did not change — drift of this kind would never republish", name)
		}
	}
}

func TestTreeHashFieldSeparatorCannotBeGamed(t *testing.T) {
	// key+etag concatenation must not collide when the boundary moves.
	a := []File{{Key: "ab", ETag: "c", Size: 1}}
	b := []File{{Key: "a", ETag: "bc", Size: 1}}
	if TreeHash(a) == TreeHash(b) {
		t.Fatal("field boundary collision: distinct trees hash equal")
	}
}

func TestTreeHashCarriesAlgorithmPrefix(t *testing.T) {
	// th2, not th1, since 2026-09-02 (bugs_open/429): the algorithm is
	// unchanged, but "published" gained the deletion half, and the prefix
	// bump is the designed lever that makes every site republish (and so
	// converge) exactly once. Bumping this assertion without a matching
	// semantics change would silently republish the fleet — don't.
	h := TreeHash(nil)
	if len(h) < 5 || h[:4] != "th2:" {
		t.Fatalf("TreeHash must carry the th2: algorithm+semantics prefix, got %q", h)
	}
}

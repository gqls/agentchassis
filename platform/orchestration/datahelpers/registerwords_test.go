// FILE: platform/orchestration/datahelpers/registerwords_test.go
//
// THE LOCKSTEP. BANNED_REGISTER_v2.json is the owner's authority and this
// package is what enforces it; the two are held together here, in both
// directions, so neither can move without the build saying so.
//
// This is the dedup-index/Go-list shape (idx_swi_dedup and
// workItemTerminalStatuses are one contract; drift is a fleet-wide fault),
// applied BEFORE the drift instead of after it. It is worth the file because
// the two artefacts have different owners: the register is maintained by the
// copy lane, this package by the platform, and the failure mode is silent —
// a word added to the register and not to Go is simply never enforced, and
// nothing anywhere reports the absence.
package datahelpers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"
)

// registerRelPath is the register's location relative to THIS package. The test
// resolves it rather than hardcoding an absolute path so it runs anywhere the
// repo is checked out.
const registerRelPath = "../../../docs/agent_docs/docs024_key_docs_latest/copy_quality_two_stage/AUDIT_prompts/BANNED_REGISTER_v2.json"

type registerFile struct {
	Version     int    `json:"version"`
	Authority   string `json:"authority"`
	BannedWords []struct {
		Pattern   string `json:"pattern"`
		Authority string `json:"authority"`
		Treatment string `json:"treatment"`
	} `json:"banned_words"`
	BannedShapes []struct {
		Name    string `json:"name"`
		Pattern string `json:"pattern"`
	} `json:"banned_shapes"`
}

func loadRegister(t *testing.T) registerFile {
	t.Helper()
	raw, err := os.ReadFile(filepath.Clean(registerRelPath))
	if err != nil {
		// ⚠ NOT t.Skip. A skipped lockstep is a lockstep that silently stopped
		// holding, and the whole point of this file is that the failure mode is
		// otherwise invisible. If the register moves, this test must go red so
		// whoever moved it updates the path deliberately.
		t.Fatalf("the banned register must be readable at %s — if it moved, update registerRelPath AND BannedRegisterPath: %v",
			registerRelPath, err)
	}
	var rf registerFile
	if err := json.Unmarshal(raw, &rf); err != nil {
		t.Fatalf("banned register does not parse: %v", err)
	}
	return rf
}

func TestBannedRegisterVersionMatchesTheFile(t *testing.T) {
	rf := loadRegister(t)
	if rf.Version != BannedRegisterVersion {
		t.Fatalf("register file is v%d but this package implements v%d — a version bump is a NEW FILE per the register's own usage rule, so add the new file and update BannedRegisterVersion/BannedRegisterPath deliberately",
			rf.Version, BannedRegisterVersion)
	}
}

// TestBannedRegisterWordsMatchTheRegisterFile is bidirectional: a word in the
// register with no Go rule is unenforced, and a Go rule with no register entry
// is a rule the owner never gave.
func TestBannedRegisterWordsMatchTheRegisterFile(t *testing.T) {
	rf := loadRegister(t)

	filePatterns := map[string]bool{}
	for _, w := range rf.BannedWords {
		filePatterns[w.Pattern] = true
	}
	goPatterns := map[string]bool{}
	for _, p := range BannedRegisterWordPatterns() {
		goPatterns[p] = true
	}

	for p := range filePatterns {
		if !goPatterns[p] {
			t.Errorf("banned_words pattern %q is in the register and in NO Go rule — it is documented and unenforced; add it to bannedRegisterWords", p)
		}
	}
	for p := range goPatterns {
		if !filePatterns[p] {
			t.Errorf("Go carries banned-word pattern %q which the register does not — a rule with no stated authority; remove it or get it into the register", p)
		}
	}
	if len(rf.BannedWords) != len(BannedRegisterWordPatterns()) {
		t.Errorf("register has %d banned words, Go has %d", len(rf.BannedWords), len(BannedRegisterWordPatterns()))
	}
}

// TestBannedRegisterShapesAreAllKnownToTheScanner asserts NAMES only, in ONE
// direction. See registerwords.go's header for why patterns are not compared:
// the register's shape patterns are coarse proxies and the scanner's are the
// authority. Go carrying MORE shapes than the register names is correct — the
// register documents what the owner ruled on, the scanner may see more.
func TestBannedRegisterShapesAreAllKnownToTheScanner(t *testing.T) {
	rf := loadRegister(t)
	known := map[string]bool{}
	for _, n := range NegationShapeNames() {
		known[n] = true
	}
	for _, s := range rf.BannedShapes {
		if !known[s.Name] {
			t.Errorf("register names banned shape %q which ScanDefineByNegation does not implement — it is banned on paper and invisible to every gate; known shapes: %v",
				s.Name, NegationShapeNames())
		}
	}
	if len(rf.BannedShapes) == 0 {
		t.Fatal("register declares no banned shapes — that cannot be right, and a vacuous pass here would hide it")
	}
}

// TestScanBannedRegisterWordsCatchesTheOwnersTwoWords is the arm that had no
// reader at all before this file. The strings are drawn from live lead_with
// points measured dirty on 2026-08-31.
func TestScanBannedRegisterWordsCatchesTheOwnersTwoWords(t *testing.T) {
	cases := []struct {
		text string
		want string
	}{
		{"Where a figure is not yet verified, this site says so plainly.", "plainly"},
		{"Every capability listed here is described honestly as either platform-derived or vendor-stated.", "honest"},
		{"An honest account of what the tool cannot do.", "honest"},
		{"Every calculator output shows its working and says plainly what it cannot answer.", "plainly"},
	}
	for _, c := range cases {
		hits := ScanBannedRegisterWords(c.text)
		if len(hits) == 0 {
			t.Fatalf("no word hit in %q — the arm is blind", c.text)
		}
		if hits[0].Name != c.want {
			t.Errorf("%q: got rule %q, want %q", c.text, hits[0].Name, c.want)
		}
		if hits[0].Kind != "word" {
			t.Errorf("%q: kind must be \"word\", got %q", c.text, hits[0].Kind)
		}
		if got := c.text[hits[0].At : hits[0].At+len(hits[0].Matched)]; got != hits[0].Matched {
			t.Errorf("%q: offset %d does not point at the match (%q vs %q)", c.text, hits[0].At, got, hits[0].Matched)
		}
	}
}

// TestScanBannedRegisterWordsLeavesOrdinaryProseAlone. A false positive here
// sends a real benefit to a model to be rewritten for nothing, so the words are
// word-anchored rather than substring matches.
func TestScanBannedRegisterWordsLeavesOrdinaryProseAlone(t *testing.T) {
	clean := []string{
		"Dishonesty is not the issue here.", // substring of "honest" inside another word
		"A plain white background improves contrast.",
		"Every figure traces to a named source.",
		"A weekly guide to what is on and where to watch it.",
	}
	for _, s := range clean {
		if hits := ScanBannedRegisterWords(s); len(hits) > 0 {
			t.Errorf("false positive on %q: %s", s, DescribeRegisterViolations(hits))
		}
	}
}

// TestScanBannedRegisterOrdersByOffset is what makes hits[0].At usable as
// AcceptNegationRewrite's protectFrom. If the two arms were simply concatenated,
// a word hit late in the sentence would sort ahead of a shape hit early in it,
// and the repair would be allowed to drop facts it must keep.
func TestScanBannedRegisterOrdersByOffset(t *testing.T) {
	// shape ("X, not Y") at ~byte 46; word ("plainly") after it.
	text := "We pick the best tool for your problem, not our favourite vendor, and we say so plainly."
	hits := ScanBannedRegister(text)
	if len(hits) < 2 {
		t.Fatalf("expected both arms to fire, got %d: %s", len(hits), DescribeRegisterViolations(hits))
	}
	for i := 1; i < len(hits); i++ {
		if hits[i].At < hits[i-1].At {
			t.Fatalf("hits are not in ascending offset order: %d then %d", hits[i-1].At, hits[i].At)
		}
	}
	if hits[0].Kind != "shape" {
		t.Errorf("the earliest construction in this sentence is the shape, got %s:%s at %d",
			hits[0].Kind, hits[0].Name, hits[0].At)
	}
}

// TestBannedRegisterPathIsTheOneTheTestReads closes the gap between the constant
// the code cites in its records and the file this test actually holds it to. A
// record citing a path nobody verifies is the "a citation is not a read" shape.
func TestBannedRegisterPathIsTheOneTheTestReads(t *testing.T) {
	if filepath.Base(BannedRegisterPath) != filepath.Base(registerRelPath) {
		t.Fatalf("BannedRegisterPath cites %q but the lockstep reads %q",
			BannedRegisterPath, registerRelPath)
	}
	if _, err := os.Stat(filepath.Clean(registerRelPath)); err != nil {
		t.Fatalf("cited register is not readable: %v", err)
	}
}

// TestRegisterWordPatternsCompileUnderRE2. The patterns are copied verbatim from
// a JSON file written for a different consumer; a PCRE-only construct would
// compile there and panic here at init.
func TestRegisterWordPatternsCompileUnderRE2(t *testing.T) {
	rf := loadRegister(t)
	for _, w := range rf.BannedWords {
		if _, err := regexp.Compile(w.Pattern); err != nil {
			t.Errorf("register banned_words pattern %q does not compile under RE2: %v", w.Pattern, err)
		}
	}
	for _, s := range rf.BannedShapes {
		if _, err := regexp.Compile(s.Pattern); err != nil {
			t.Errorf("register banned_shapes pattern %q does not compile under RE2: %v", s.Pattern, err)
		}
	}
}

// ⚠⚠ TestNoNewerRegisterVersionIsUnaccountedFor — the hole in every OTHER test in
// this file, closed before it could open.
//
// Every assertion above is anchored to the literal `BANNED_REGISTER_v2.json`. The
// register's own usage rule is that "a new version is a NEW FILE line, never an
// in-place semantic change" — so the day a v2 is cut, v1 keeps existing, keeps
// saying version 1, and every test above keeps PASSING while the estate runs a
// register the Go gate does not implement. The lockstep guards drift WITHIN a
// version and is blind to the ARRIVAL of one.
//
// That is this estate's "a census goes stale BY ADDITION" shape pointed at a
// test: it enumerates from a fixed filename, so a new file is invisible to it,
// and the failure mode is a confident green.
//
// So this test does not read a filename. It ENUMERATES the register directory and
// asserts the Go constant tracks the HIGHEST version present. A v2 landing makes
// the build red until registerwords.go moves — which is the contract this package
// exists to hold, applied to the version axis rather than the word axis.
func TestNoNewerRegisterVersionIsUnaccountedFor(t *testing.T) {
	dir := filepath.Dir(registerRelPath)
	matches, err := filepath.Glob(filepath.Join(dir, "BANNED_REGISTER_v*.json"))
	if err != nil {
		t.Fatalf("globbing the register directory failed: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("no BANNED_REGISTER_v*.json found under %s — the register moved; update registerRelPath, BannedRegisterPath and this glob together", dir)
	}

	verRe := regexp.MustCompile(`BANNED_REGISTER_v(\d+)\.json$`)
	highest, highestFile := 0, ""
	for _, m := range matches {
		sub := verRe.FindStringSubmatch(filepath.Base(m))
		if sub == nil {
			t.Errorf("register file %q does not carry a parseable version — the naming convention is what this guard reads", m)
			continue
		}
		n, err := strconv.Atoi(sub[1])
		if err != nil {
			t.Errorf("unparseable version in %q: %v", m, err)
			continue
		}
		if n > highest {
			highest, highestFile = n, m
		}
	}

	if highest != BannedRegisterVersion {
		t.Fatalf(`the highest register version on disk is v%d (%s) but this package implements v%d.

A NEW REGISTER VERSION HAS LANDED AND GO DOES NOT ENFORCE IT. Every other test in
this file is anchored to the v%d filename and will keep passing regardless, which
is exactly why this one enumerates instead. To close it, in ONE commit:
  - update BannedRegisterVersion and BannedRegisterPath in registerwords.go
  - update registerRelPath in this file
  - reconcile bannedRegisterWords against the new file's banned_words
  - confirm every new banned_shapes name exists in NegationShapeNames()`,
			highest, highestFile, BannedRegisterVersion, BannedRegisterVersion)
	}

	// And the file the rest of this suite reads must BE that highest one, or the
	// lockstep is holding Go against a superseded register.
	if filepath.Base(registerRelPath) != filepath.Base(highestFile) {
		t.Errorf("the lockstep reads %s but the highest version present is %s — the suite is validating against a superseded register",
			filepath.Base(registerRelPath), filepath.Base(highestFile))
	}
}

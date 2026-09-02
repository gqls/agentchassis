// FILE: platform/orchestration/datahelpers/registerwords.go
//
// THE BANNED REGISTER, made readable from Go.
//
// The owner's copy rulings are held as versioned data at
// `docs/agent_docs/docs024_key_docs_latest/copy_quality_two_stage/AUDIT_prompts/BANNED_REGISTER_v1.json`.
// That file has three arms. Two of them already had Go readers when this file
// was written (2026-08-31):
//
//	banned_shapes           -> negationShapes in negationtells.go (7 shapes; the
//	                           register names 6, Go carries `staccato` as well)
//	banned_classes_no_regex -> by definition unmechanised; a human reads for them
//
// The third — banned_words — had NO Go reader ANYWHERE. `plainly` and `honest*`
// are in neither globalTellPhrases() nor any other list, so every mechanical
// gate in this estate was blind to the two words the owner named FIRST. This
// file is that arm, and nothing else: it is deliberately not a second home for
// the shapes, which stay where they are.
//
// ── WHY THE PATTERNS ARE COPIED AND NOT LOADED ───────────────────────────────
// The register lives under docs/ and docs/ is not in the image. A runtime read
// would work on a workstation and fail in the cluster, which is the worst of
// the two. So the patterns are duplicated here VERBATIM — byte-identical to the
// JSON's `pattern` strings, which are already Go-compatible RE2 — and
// `TestBannedRegisterWordsMatchTheRegisterFile` holds the two in lockstep in
// BOTH directions. That is the dedup-index/Go-list shape this estate already
// uses, applied BEFORE the drift rather than after it.
//
// ⚠ THE LOCKSTEP IS ON PATTERNS FOR WORDS AND ON NAMES FOR SHAPES, and the
// asymmetry is deliberate rather than an omission. The register's `banned_shapes`
// patterns are coarse documentation proxies (`,\s+not\s+\w`); the ones in
// negationtells.go are narrower and are the authority — the register's own
// negative_reveal entry says so ("broad proxy; the gate's own definition is
// narrower"). Holding those two to pattern equality would force the careful
// regex down to the loose one. Names are what must not drift.
//
// ── VERSION ─────────────────────────────────────────────────────────────────
// This file implements BANNED_REGISTER **v1**. The register's own usage rule is
// that "a new version is a new file line, never an in-place semantic change", so
// a v2 arrives as a new file and this constant is what a reader compares against.
package datahelpers

import (
	"fmt"
	"regexp"
	"sort"
)

// BannedRegisterVersion is the register version this file implements. It is
// asserted against the JSON's own `version` field by the lockstep test, so a
// register bumped without a Go change fails the build rather than drifting.
const BannedRegisterVersion = 2

// BannedRegisterPath is where the authority lives, for citation in findings and
// records. Nothing reads it at runtime — see the header for why.
const BannedRegisterPath = "docs/agent_docs/docs024_key_docs_latest/copy_quality_two_stage/AUDIT_prompts/BANNED_REGISTER_v2.json"

// bannedRegisterWord is one banned-word rule, carrying the owner's authority so
// a finding can say WHY a word is refused rather than merely that it is.
type bannedRegisterWord struct {
	Name      string // stable identifier used in records; not from the JSON
	Pattern   string // byte-identical to the register's `pattern`
	Authority string // byte-identical to the register's `authority`
	Treatment string // byte-identical to the register's `treatment`
	re        *regexp.Regexp
}

// bannedRegisterWords is banned_words from BANNED_REGISTER_v2, in file order.
//
// ⚠ THE ONE BLESSED EXCEPTION IS NOT ENCODED HERE. The register records that
// idea.uk's report hero may say "honest" (decision record D-005). A per-site
// exemption belongs to the caller, which knows the site; a scanner that carved
// it out would apply it to every site in the estate. Callers that need it must
// exempt by site, and say so.
var bannedRegisterWords = []bannedRegisterWord{
	{
		Name:      "plainly",
		Pattern:   `\bplainly\b`,
		Authority: "ruling 9, 2026-08-31",
		Treatment: "delete the label; if the sentence needs the meaning, show it by being plain",
	},
	{
		Name:      "honest",
		Pattern:   `\bhonest(?:ly|y)?\b`,
		Authority: "owner ruling 2026-07-26 (writer rule 19)",
		Treatment: "delete the label; keep being straight",
	},
	{
		Name:      "plain_words",
		Pattern:   `\b(?:in\s+)?plain\s+(?:words|english|terms|language)\b`,
		Authority: "owner ruling 2026-09-02 (v2 GO), generalising ruling 9",
		Treatment: "delete the label; a site that must SAY its words are plain is not showing it",
	},
}

func init() {
	for i := range bannedRegisterWords {
		bannedRegisterWords[i].re = regexp.MustCompile(`(?i)` + bannedRegisterWords[i].Pattern)
	}
}

// RegisterViolation is one banned-register hit in a piece of text.
//
// Shapes and words arrive in the SAME type on purpose: a caller repairing a
// point should not have to know which arm caught it, and a record that
// distinguishes them by Kind stays countable per arm afterwards. That per-arm
// count is the thing the producer's error rate is actually made of — the first
// census of this corpus put the word arm at 10% of violations and the shape arm
// at 90%, and getting that split wrong by 2.8x is what made an earlier estimate
// useless (HANDOFF_2026-08-26b §H1).
type RegisterViolation struct {
	Kind      string `json:"kind"`                // "shape" | "word"
	Name      string `json:"name"`                // shape name, or banned-word rule name
	Matched   string `json:"matched"`             // the matched fragment, verbatim
	At        int    `json:"at"`                  // byte offset of Matched in the scanned text
	Authority string `json:"authority,omitempty"` // words only; shapes carry theirs in the register
}

// ScanBannedRegisterWords finds banned_words hits only. Exported separately
// because the word arm is the half that had no reader: a caller already running
// ScanDefineByNegation can add this without re-running the shapes.
func ScanBannedRegisterWords(text string) []RegisterViolation {
	var out []RegisterViolation
	for _, w := range bannedRegisterWords {
		for _, loc := range w.re.FindAllStringIndex(text, -1) {
			out = append(out, RegisterViolation{
				Kind: "word", Name: w.Name,
				Matched:   text[loc[0]:loc[1]],
				At:        loc[0],
				Authority: w.Authority,
			})
		}
	}
	return out
}

// ScanBannedRegister finds every banned-register violation in text: the shape
// arm via ScanDefineByNegation (the authority for shapes, unchanged and not
// duplicated), plus the word arm above. Hits are returned in ASCENDING OFFSET
// order, which is what makes out[0].At usable as AcceptNegationRewrite's
// protectFrom — the earliest construction is the one everything after it may be
// dropped from.
func ScanBannedRegister(text string) []RegisterViolation {
	out := ScanBannedRegisterWords(text)
	for _, h := range ScanDefineByNegation(text) {
		out = append(out, RegisterViolation{
			Kind:    "shape",
			Name:    h.Shape,
			Matched: h.Matched,
			At:      h.SentenceStart + h.MatchInSent,
		})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].At < out[j].At })
	return out
}

// DescribeRegisterViolations renders hits for a log line or an error, in a form
// a human can act on without opening the register.
func DescribeRegisterViolations(hits []RegisterViolation) string {
	if len(hits) == 0 {
		return ""
	}
	parts := make([]string, 0, len(hits))
	for _, h := range hits {
		parts = append(parts, fmt.Sprintf("%s:%s(%q)", h.Kind, h.Name, h.Matched))
	}
	return joinWithPipe(parts)
}

func joinWithPipe(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += " | "
		}
		out += p
	}
	return out
}

// BannedRegisterWordNames is the word vocabulary, for tests and records. Order
// is stable.
func BannedRegisterWordNames() []string {
	out := make([]string, 0, len(bannedRegisterWords))
	for _, w := range bannedRegisterWords {
		out = append(out, w.Name)
	}
	return out
}

// BannedRegisterWordPatterns maps rule name -> the register pattern it carries,
// which is what the lockstep test compares against the JSON.
func BannedRegisterWordPatterns() map[string]string {
	out := make(map[string]string, len(bannedRegisterWords))
	for _, w := range bannedRegisterWords {
		out[w.Name] = w.Pattern
	}
	return out
}

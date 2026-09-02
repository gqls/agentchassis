// FILE: platform/orchestration/datahelpers/negationtells.go
//
// The define-by-negation family, as a pure scanner (bugs_open/305).
//
// WHAT IT IS. Saying what a thing is NOT in order to say what it is:
// "The registry shows you what's possible, not what survives production."
// The owner read three live pages, quoted two sentences of exactly this shape,
// and ruled that "that sort of copy never leaves this framework again".
//
// WHY A SEPARATE FILE FROM voicetells.go. That file's `strawmanCommaRe` needs a
// trailing ", but" — it matches "not X, but Y" and nothing else, which is 1.5%
// of writer sections and neither of the owner's sentences (measured over 1,503
// page-content-writer calls, 2026-08-13..19). The shapes below are the other
// 42%. voicetells.go's strawman arm now calls into here, so the site-opt-in
// voice gate and the meta-description gate inherit them; this file stays free of
// that gate's config and thresholds so the writer seam can use it unconditionally.
//
// WHY A RULE IN THE PROMPT IS NOT ENOUGH, i.e. why this is code at all. The
// house voice already says "Say what a thing IS rather than what it is not", and
// migration 228 shipped the same rule a fortnight earlier with its own header
// recording that it "did not work". The fleet_copy_quality lane's conclusion:
// "A rule can only name a form. What goes wrong is an instinct" — and
// "prescriptions become tics". So the rule is kept, and a mechanical check is
// added behind it.
//
// THE THREE THINGS THIS FILE IS SHAPED BY, all of them measured refutations of
// an earlier draft of it:
//
//  1. A REWRITE CAN DISPLACE RATHER THAN FIX. "X instead of Y", "more than just
//     Y", "isn't just Y", an em dash — all score ZERO on the five shapes and are
//     the same instinct. Baselines in the same corpus: instead of 5.9%, isn't
//     just/a 6.4%, more than (just) 10.8%. So ScanContrastNeighbours exists as a
//     SEPARATE class: never a reason to trip, only a reason to REJECT a rewrite.
//     Anything else lets a repair pass by moving the problem, which is
//     copy_quality_two_stage's "a prohibition displaces a problem rather than
//     solving it" arriving from the measurement side.
//  2. "EXEMPT ANYTHING THE PROMPT ALREADY CONTAINS" IS FATAL. The literal string
//     "rather than" appears in every rendered writer prompt — the house voice
//     uses it six times and STRICT RULE 19 uses it — so a prompt-wide exemption
//     silently exempts the whole 43% arm. Exemption here is SENTENCE-scoped and
//     matches against brief-supplied text only.
//  3. A SITE'S OWN VOICE OUTRANKS THE FLEET RULES, by the house voice's own
//     first sentence. A phrase the brief HANDS the writer is the brief's
//     decision: it is counted and left alone. Fixing it means fixing the brief.
//
// Every signal here is regex or arithmetic — no LLM, no I/O, no config.

package datahelpers

import (
	"regexp"
	"strings"
)

// NegationHit is one define-by-negation construction located in prose.
//
// Sentence and SentenceStart are the load-bearing fields: the repair operates on
// whole sentences spliced back by exact substring, so a hit that cannot name its
// own sentence cannot be repaired safely.
type NegationHit struct {
	Shape         string // x_not_y | not_x_but_y | staccato | rather_than | negative_reveal
	Matched       string // the matched fragment, verbatim
	Sentence      string // the containing sentence, tag-trimmed
	SentenceStart int    // byte offset of Sentence in the scanned text
	MatchInSent   int    // byte offset of Matched within Sentence
	Field         string // the content field this text came from; set by callers
}

// The five shapes. Go RE2 — `\b` is the word boundary here; in Postgres it is a
// BACKSPACE character and `\y` is the boundary, so never paste these into psql
// (LANDMINES.md, and it produced two confident zeroes in this bug's own census).
var (
	// "…what's possible, not what survives production" — the owner's sentence.
	// The leading character class is how RE2 does a lookbehind: the match
	// deliberately includes the char before the comma.
	// v2 (2026-09-02): the separator admits an em/en dash (spaced or unspaced)
	// and a spaced hyphen, as well as the comma. Found live by the offer lane's
	// producer gate: "straight guidance at every price point — not a default to
	// premium" passed every comma-anchored consumer — including that gate's own
	// re-scan — and shipped inside a point recorded as REPAIRED. Same shape
	// name; the register's v2 entry is the documentation proxy, this regex
	// decides (BANNED_REGISTER_v2; registerwords lockstep is names-only).
	negXNotYRe = regexp.MustCompile(`(?i)[\pL\pN)"'’](?:[,—–-]\s+|\s[—–]\s*)(?:not|never)\s+(?:just\s+|merely\s+|simply\s+|only\s+)?[\pL\pN]`)

	// "not a framework, but a system" — the shape voicetells.go already had.
	negNotXButYRe = regexp.MustCompile(`(?i)\bnot (?:just |merely |simply |about |only )?[^.;:]{2,50},\s*but\b`)

	// "Not a demo. Not a proof of concept." — the leopardess CTA that motivated
	// the original voice check.
	negStaccatoRe = regexp.MustCompile(`(?i)\bnot an? [^.]{2,40}\.\s+not an? `)

	// "persuaded rather than sold to". 43% of sections; the broadest arm, and the
	// one most likely to be narrowed once the rejection log has a week of traffic.
	negRatherThanRe = regexp.MustCompile(`(?i)\brather than\b`)

	// "instead of the one a vendor happens to sell" — the same comparison in a
	// third coat; the owner named this phrasing explicitly (Decision B).
	negInsteadOfRe = regexp.MustCompile(`(?i)\binstead of\b`)

	// "not just those with engineering teams" — the minimising comparison. The
	// negative lookahead-free guard: the ", not just" form is already x_not_y's
	// and the "not just X, but" form is not_x_but_y's; overlap is harmless (the
	// repair is sentence-keyed, and one sentence listed twice still yields one
	// replacement), so this pattern stays simple rather than clever.
	negNotJustRe = regexp.MustCompile(`(?i)\bnot (?:just|only|merely)\b`)

	// "A model directory tells you which agents exist. It doesn't tell you how
	// they hold up…" — the construction spread over two sentences, which is the
	// owner's second quoted example and is invisible to every single-sentence
	// pattern.
	//
	// ⚠ THE SUBJECT LIST IS THIRD-PERSON ONLY, AND DELIBERATELY SO. An earlier
	// version included "we", and the first live fleet run showed what that
	// costs: it flagged "we do not offer refunds", "we do not invent figures"
	// and "we do not charge for the first call". Those are statements of a
	// POLICY OR A LIMIT — which the writer's own STRICT RULE 19 explicitly asks
	// for ("name the limit, the failure mode, or what the thing cannot do") —
	// not a thing being defined by what it is not. Flagging them sends a human
	// to edit a company's stated policy, and would ask the repair to delete a
	// commitment. The mannerism the owner objected to negates a THING: "It
	// doesn't tell you how they hold up."
	negReveseRe = regexp.MustCompile(`(?i)(?:\A|[.!?]["'’)]?\s+)(?:it|this|that|they|these)\s+(?:does\s?n['’]t|does not|is\s?n['’]t|is not|wo\s?n['’]t|will not|ca\s?n['’]t|cannot|are\s?n['’]t|are not|do\s?n['’]t|do not)\b`)
)

var negationShapes = []struct {
	shape string
	re    *regexp.Regexp
}{
	{"x_not_y", negXNotYRe},
	{"not_x_but_y", negNotXButYRe},
	{"staccato", negStaccatoRe},
	{"rather_than", negRatherThanRe},
	{"negative_reveal", negReveseRe},
	// The two below joined 2026-08-31 (owner Decision B — he named "instead of"
	// verbatim in the truncation-trial ruling, and both shapes shipped untouched
	// through the gate on the canary pages because the gate could not see them).
	{"instead_of", negInsteadOfRe},
	{"not_just", negNotJustRe},
}

// The neighbour set: the same instinct in a grammar the five shapes do not
// name. NEVER a trip — only a reason to reject a rewrite that reached for one.
//
// Each entry is deliberately narrow enough that it cannot fire on an ordinary
// factual comparison: "more than 30 agents" must not count, so the qualifier
// (just/merely/simply) is required. The em dash is here because the house voice
// bans it outright, so a rewrite that introduces one has traded one violation
// for another.
var contrastNeighbourRes = []struct {
	shape string
	re    *regexp.Regexp
}{
	// `instead_of` LEFT this set 2026-08-31: owner Decision B promoted it to a
	// tripping SHAPE (he named the phrasing verbatim in the truncation-trial
	// ruling). A rewrite that reaches for it is still rejected — the shape scan
	// covers what the neighbour scan used to.
	{"more_than_just", regexp.MustCompile(`(?i)\bmore than (?:just|merely|simply)\b`)},
	{"isnt_just", regexp.MustCompile(`(?i)\b(?:is|are|was|were|does|do|did)\s?n['’]t\s+(?:just|only|merely|simply|about|a|an)\b`)},
	{"is_not_just", regexp.MustCompile(`(?i)\b(?:is|are|was|were|does|do|did)\s+not\s+(?:just|only|merely|simply|about)\b`)},
	{"unlike", regexp.MustCompile(`(?i)\bunlike\b`)},
	{"as_opposed_to", regexp.MustCompile(`(?i)\bas opposed to\b`)},
	{"far_from", regexp.MustCompile(`(?i)\bfar from\b`)},
	{"no_longer_just", regexp.MustCompile(`(?i)\bno longer (?:just|only|merely)\b`)},
	{"without_being", regexp.MustCompile(`(?i)\bwithout being\b`)},
	{"em_dash", regexp.MustCompile(`—`)},
}

// Regulatory and capability negations. These are REQUIRED text or a statement of
// what we cannot do, and the estate's rules elsewhere insist on them: the writer
// prompt's own STRICT RULE forbids promising accuracy and requires saying a tool
// can give a wrong answer. Rewriting them would be a compliance harm dressed as
// a style fix, so they are exempt wherever they appear.
var regulatoryNegationRe = regexp.MustCompile(`(?i)\bnot (?:financial|legal|medical|tax|investment|regulatory|professional) advice\b` +
	`|\bnot (?:a|an) (?:lender|broker|adviser|advisor|solicitor|accountant|insurer|regulated\s+\w+)\b` +
	`|\bnot regulated\b|\bdoes not constitute\b|\bcannot guarantee\b|\bcan(?:no|)['’]?t guarantee\b` +
	`|\bnot guaranteed\b|\bno guarantee\b|\bnot a (?:quote|offer|recommendation|substitute)\b` +
	`|\bnot intended as\b|\bcan (?:still )?be wrong\b|\bcan give a wrong answer\b|\bwe cannot tell you\b`)

// Superlatives and absolutes. A rewrite may not INTRODUCE one that the original
// did not have.
//
// This exists because the council's compliance seat pointed out (round 4) that
// the claim guard leaned entirely on checkBannedClaims, which only catches
// patterns a site has actually armed — and the register is sparse. "Say what it
// IS" is exactly the pressure that fills the slot the removed contrast leaves
// with an absolute, and an unarmed site would have had nothing standing between
// that and the page. It is deliberately a short, closed list of words that are
// almost never true and never necessary: the check is "did the rewrite reach for
// one the author had not", not "is this word banned".
var superlativeRe = regexp.MustCompile(`(?i)\b(definitive|guaranteed?|guarantees|unmatched|unrivalled|unrivaled|flawless|foolproof|industry[- ]leading|best[- ]in[- ]class|world[- ]class|cutting[- ]edge|state[- ]of[- ]the[- ]art|always|never fails?|every single|fully (?:verified|accurate|automated|managed)|100%|perfect(?:ly)?|seamless(?:ly)?)\b`)

var (
	htmlTagRe     = regexp.MustCompile(`</?[A-Za-z][^>]*>`)
	numberTokenRe = regexp.MustCompile(`\d[\d,.]*`)
	linkTokenRe   = regexp.MustCompile(`(?i)https?://[^\s"'<>]+|/[A-Za-z0-9._~/-]*\.(?:html|htm|php|pdf)|href="[^"]*"`)
	// ⚠ `\p{Lu}`, NOT `\pLu`. The one-letter form takes exactly one letter of
	// class name, so `\pLu` parses as "any letter, then a literal u" and this
	// regex matched "running" and "our" — every rewrite was rejected as
	// invented_name. Caught by a test asserting what it matched, not by reading it.
	capTokenRe = regexp.MustCompile(`\b\p{Lu}[\p{L}'’-]*\b`)
	nonAlnumRe = regexp.MustCompile(`[^\pL\pN]+`)
)

// ScanDefineByNegation finds every construction in the five shapes, each
// attributed to its containing sentence.
//
// It scans the RAW field value, HTML and all, because the repair splices by
// exact substring and a sentence recovered from stripped text is not a substring
// of the field it came from. Sentence edges are trimmed of whole tags so the
// sentence handed to a human or a model reads as prose.
// NegationShapeIsMild reports whether a detected shape is one the owner has
// ruled is only mildly objectionable.
//
// OWNER RULING 2026-08-24 (bugs_closed/305, decision D3): "`rather than` is a
// little bit of a tic." That is deliberately neither of the two answers D3 was
// framed with — it is not "a tic" (repair it like the rest) and not "ordinary
// English" (stop detecting it). The estate implements the middle as a question of
// WHO GETS FORGIVEN: the copy gate lets a page keep `page_budget` constructions
// (2 by default) and repairs the rest, and before this ruling the two survivors
// were simply whichever the scanner walked past FIRST — document order, nothing
// to do with severity. A page could therefore keep both of its "X, not Y"
// constructions and spend the gate's effort rewriting two "rather than"s further
// down, which is the ruling inverted.
//
// So mildness decides the tolerance, not detection: a mild shape can consume the
// page budget, a sharp one cannot and is always repaired. `rather than` stays
// fully detected and still counts toward `page_hits` and the density signal.
//
// Why this is a function on the vocabulary rather than a config key: which shapes
// are mild is a judgement about English recorded as a ruling, not an operational
// dial an operator should turn per site. `page_budget` remains the dial.
//
// Scale, for anyone tempted to widen this: `rather_than` was 71% of all gate
// rewrites and appears in 43% of writer sections (measured 2026-08-23/24), so
// moving one shape in or out of this set changes most of the gate's traffic.
func NegationShapeIsMild(shape string) bool {
	_, ok := mildNegationShapes[shape]
	return ok
}

// A set rather than an equality test, so this reads as a policy list.
//
// EMPTY since 2026-08-31 (owner Decision A, in-session: "repair every one").
// This repeals D3's mild-forgiveness for `rather_than`: the canary measured the
// forgiven allowance — applied per SECTION by the recorded landmine — shipping
// ~6 `rather than` per multi-section page with no repair ever attempted, and
// the truncation repair (ruling 7) is lossless, so nothing earns forgiveness
// any more. The budget machinery above stays in place, inert: with no mild
// shape, nothing can spend it. To reinstate an allowance, add a shape here and
// cite the ruling that reopens it.
var mildNegationShapes = map[string]struct{}{}

func ScanDefineByNegation(text string) []NegationHit {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	spans := negationSentenceSpans(text)
	var hits []NegationHit
	for _, s := range negationShapes {
		for _, loc := range s.re.FindAllStringIndex(text, -1) {
			// ⚠ ANCHOR ON THE CONSTRUCTION, NOT ON THE MATCH START. The
			// negative_reveal pattern deliberately begins with the PREVIOUS
			// sentence's full stop, so a raw loc[0] attributes the hit to the
			// sentence before the one that carries the fault — and the repair
			// would then hand the model a clean, true sentence to rewrite while
			// leaving the reveal untouched. Measured on the owner's own page:
			// "A model directory tells you which agents exist. It doesn't tell
			// you how they hold up…" was attributed to the first sentence.
			// Skipping the leading terminator and quotes moves the anchor onto
			// the subject, i.e. into the sentence that has to change.
			anchor := loc[0]
			for anchor < loc[1] && !isNegationAnchorRune(text[anchor]) {
				anchor++
			}
			sentStart, sentEnd := containingSentence(spans, anchor, len(text))
			sentence := text[sentStart:sentEnd]
			mis := anchor - sentStart
			if mis < 0 || mis > len(sentence) {
				mis = 0
			}
			matched := text[anchor:loc[1]]
			if mis+len(matched) > len(sentence) {
				matched = sentence[mis:]
			}
			hits = append(hits, NegationHit{
				Shape:         s.shape,
				Matched:       matched,
				Sentence:      sentence,
				SentenceStart: sentStart,
				MatchInSent:   mis,
			})
		}
	}
	return hits
}

// ScanContrastNeighbours finds the displacement set. Reported separately and
// never counted as a trip — see this file's header, point 1.
func ScanContrastNeighbours(text string) []NegationHit {
	var hits []NegationHit
	for _, n := range contrastNeighbourRes {
		for _, loc := range n.re.FindAllStringIndex(text, -1) {
			hits = append(hits, NegationHit{Shape: n.shape, Matched: text[loc[0]:loc[1]]})
		}
	}
	return hits
}

// negationSentenceSpans returns [start,end) byte spans of the sentences in text.
//
// Terminator = .!? run followed by whitespace, a tag, or end of text. A closing
// tag or a <br> also ends a sentence, because a rich_text field is
// "<p>One.</p><p>Two</p>" and its second paragraph may carry no full stop at all.
// isNegationAnchorRune reports whether a byte can START the construction: a
// letter, a digit, or the punctuation the x_not_y shape deliberately captures
// before its comma. Terminators, quotes and whitespace cannot.
func isNegationAnchorRune(b byte) bool {
	switch {
	case b >= 'a' && b <= 'z', b >= 'A' && b <= 'Z', b >= '0' && b <= '9':
		return true
	case b == ',' || b == ')' || b >= 0x80: // the captured pre-comma char, incl. multi-byte
		return true
	}
	return false
}

func negationSentenceSpans(text string) [][2]int {
	var spans [][2]int
	start := 0
	i := 0
	for i < len(text) {
		c := text[i]
		if c == '.' || c == '!' || c == '?' {
			j := i
			for j < len(text) && (text[j] == '.' || text[j] == '!' || text[j] == '?') {
				j++
			}
			// A terminator only ends the sentence if what follows is space, a
			// tag, or nothing: "v1.0" and "e.g." keep their sentence intact.
			if j >= len(text) || text[j] == ' ' || text[j] == '\n' || text[j] == '\t' || text[j] == '\r' || text[j] == '<' {
				spans = appendSentenceSpan(spans, text, start, j)
				start = j
				i = j
				continue
			}
			i = j
			continue
		}
		if c == '<' {
			if m := htmlTagRe.FindStringIndex(text[i:]); m != nil && m[0] == 0 {
				tag := strings.ToLower(text[i : i+m[1]])
				// ⚠ `</th` is NOT covered by the `</h` arm — "</th>" is `<`,`/`,`t`,`h`
				// so it fails HasPrefix("</h"). It was absent until 2026-08-24 while
				// its sibling `</td` was present, so a table's HEADER row read as one
				// sentence that CONTAINED RAW MARKUP: the probe returned
				// "Real, not simulated</th><th>Throughput" as a single x_not_y hit.
				// That is worse than a missed hit — the captured sentence is what a
				// repair splices over, so a rewrite would have eaten the cell tags.
				// Newly reachable via the 594/595 pair, which retypes five prose slots
				// to html and tells the writer to emit <table>.
				if strings.HasPrefix(tag, "</p") || strings.HasPrefix(tag, "<br") ||
					strings.HasPrefix(tag, "</li") || strings.HasPrefix(tag, "</h") ||
					strings.HasPrefix(tag, "</div") || strings.HasPrefix(tag, "</td") ||
					strings.HasPrefix(tag, "</th") || strings.HasPrefix(tag, "</tr") {
					spans = appendSentenceSpan(spans, text, start, i)
					start = i
				}
				i += m[1]
				continue
			}
		}
		i++
	}
	spans = appendSentenceSpan(spans, text, start, len(text))
	return spans
}

func appendSentenceSpan(spans [][2]int, text string, start, end int) [][2]int {
	s, e := trimTagEdges(text, start, end)
	if e > s {
		spans = append(spans, [2]int{s, e})
	}
	return spans
}

// trimTagEdges moves a span's edges inward past whitespace and whole HTML tags,
// so a sentence reads as prose while its offsets still address the raw text.
func trimTagEdges(text string, start, end int) (int, int) {
	for start < end {
		if text[start] == ' ' || text[start] == '\n' || text[start] == '\t' || text[start] == '\r' {
			start++
			continue
		}
		if text[start] == '<' {
			if m := htmlTagRe.FindStringIndex(text[start:end]); m != nil && m[0] == 0 {
				start += m[1]
				continue
			}
		}
		break
	}
	for end > start {
		if c := text[end-1]; c == ' ' || c == '\n' || c == '\t' || c == '\r' {
			end--
			continue
		}
		if text[end-1] == '>' {
			// Walk back to this tag's '<' and drop it only if the whole tag sits
			// inside the span.
			if k := strings.LastIndexByte(text[start:end], '<'); k >= 0 {
				if m := htmlTagRe.FindStringIndex(text[start+k : end]); m != nil && m[0] == 0 && start+k+m[1] == end {
					end = start + k
					continue
				}
			}
		}
		break
	}
	return start, end
}

// containingSentence picks the span holding offset pos. A match that straddles a
// boundary (the two-sentence reveal always does, since its pattern begins with
// the previous sentence's full stop) is attributed to the sentence its OWN text
// starts in, which is the one that has to be rewritten.
func containingSentence(spans [][2]int, pos, textLen int) (int, int) {
	for _, s := range spans {
		if pos >= s[0] && pos < s[1] {
			return s[0], s[1]
		}
	}
	// A match landing in trimmed-away punctuation between spans: use the next
	// span if there is one, else the whole text.
	for _, s := range spans {
		if s[0] > pos {
			return s[0], s[1]
		}
	}
	if len(spans) > 0 {
		return spans[len(spans)-1][0], spans[len(spans)-1][1]
	}
	return 0, textLen
}

// NegationExempt reports whether a hit must be left alone, and why.
//
// Two exemptions, both narrow:
//
//	regulatory        — required text or a statement of a limit (see the regex).
//	brief_supplied…   — the site's own brief handed the writer this sentence, or
//	                    a phrase within it. The brief's decision, not the
//	                    writer's mistake; "a site's own voice specification
//	                    outranks these rules".
//
// `supplied` is the brief-supplied text ONLY — never the whole rendered prompt.
// A prompt-wide test exempts every "rather than" in the estate, because the
// house voice itself uses the phrase six times.
func NegationExempt(hit NegationHit, supplied []string) (bool, string) {
	if regulatoryNegationRe.MatchString(hit.Sentence) {
		return true, "regulatory"
	}
	sent := normaliseForMatch(hit.Sentence)
	cores := negationCores(hit)
	for _, sup := range supplied {
		n := normaliseForMatch(sup)
		if n == "" {
			continue
		}
		if len(sent) >= 12 && strings.Contains(n, sent) {
			return true, "brief_supplied_sentence"
		}
		for _, c := range cores {
			nc := normaliseForMatch(c)
			if len(nc) >= 18 && strings.Contains(n, nc) {
				return true, "brief_supplied_phrase"
			}
		}
	}
	return false, ""
}

// negationCores returns candidate phrasings of the construction, widest first:
// the matched text plus k words either side, for k in {4, 3, 2}.
//
// A single fixed window does not work in either direction, and the narrow-only
// version is dangerous. Measured on the live case: the brief supplies
// "…deployed to production in days, not months" and the writer emitted
// "Multi-agent systems deployed to production in days, not months on
// Kubernetes." — the wide window carries "on Kubernetes", which the brief never
// said, so a single wide core reports NOT-supplied for a phrase the brief
// demonstrably handed over. Narrowing until something matches fixes that; the
// 18-character floor on the normalised form is what stops it narrowing to
// "s, not m" and exempting everything.
func negationCores(hit NegationHit) []string {
	s := hit.Sentence
	if s == "" {
		return []string{hit.Matched}
	}
	words, starts := splitWordsWithOffsets(s)
	if len(words) == 0 {
		return []string{hit.Matched}
	}
	wi := 0
	for i, off := range starts {
		if off <= hit.MatchInSent {
			wi = i
		}
	}
	var out []string
	for _, k := range []int{4, 3, 2} {
		lo, hi := wi-k, wi+k+1
		if lo < 0 {
			lo = 0
		}
		if hi > len(words) {
			hi = len(words)
		}
		out = append(out, strings.Join(words[lo:hi], " "))
	}
	return out
}

func splitWordsWithOffsets(s string) ([]string, []int) {
	var words []string
	var offs []int
	i := 0
	for i < len(s) {
		for i < len(s) && (s[i] == ' ' || s[i] == '\n' || s[i] == '\t' || s[i] == '\r') {
			i++
		}
		if i >= len(s) {
			break
		}
		j := i
		for j < len(s) && !(s[j] == ' ' || s[j] == '\n' || s[j] == '\t' || s[j] == '\r') {
			j++
		}
		words = append(words, s[i:j])
		offs = append(offs, i)
		i = j
	}
	return words, offs
}

func normaliseForMatch(s string) string {
	s = strings.NewReplacer("’", "'", "“", `"`, "”", `"`, "—", " ", "–", " ").Replace(s)
	s = htmlTagRe.ReplaceAllString(s, " ")
	s = nonAlnumRe.ReplaceAllString(strings.ToLower(s), " ")
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}

// AcceptNegationRewrite decides whether a proposed replacement sentence may
// stand in for the original. It is the whole reason a repair here cannot quietly
// make things worse, so it fails CLOSED: any doubt keeps the original.
//
// protectFrom is the byte offset in `from` of the construction being removed.
// Facts BEFORE it are the claim and must survive; facts after it are the
// contrasted alternative the rewrite is meant to drop. That asymmetry is the
// only reason a rule like "preserve every number" can coexist with "delete the
// contrast": "We run 1,600 orchestrations a day, not 12" must keep 1,600 and may
// lose the 12.
func AcceptNegationRewrite(from, to string, protectFrom int) (bool, string) {
	to = strings.TrimSpace(to)
	if to == "" {
		return false, "empty"
	}
	if normaliseForMatch(to) == normaliseForMatch(from) {
		return false, "unchanged"
	}
	if len(to) < len(from)*2/5 {
		return false, "gutted"
	}
	if len(to) > len(from)*22/10 {
		return false, "ballooned"
	}
	if hits := ScanDefineByNegation(to); len(hits) > 0 {
		return false, "still_" + hits[0].Shape
	}
	before := neighbourCounts(from)
	after := neighbourCounts(to)
	for shape, n := range after {
		if n > before[shape] {
			return false, "displaced_" + shape
		}
	}
	if protectFrom < 0 || protectFrom > len(from) {
		protectFrom = len(from)
	}
	for _, num := range numberTokenRe.FindAllString(from[:protectFrom], -1) {
		if !strings.Contains(to, num) {
			return false, "dropped_figure"
		}
	}
	for _, num := range numberTokenRe.FindAllString(to, -1) {
		if !strings.Contains(from, num) {
			return false, "invented_figure"
		}
	}
	for _, l := range linkTokenRe.FindAllString(from, -1) {
		if !strings.Contains(to, l) {
			return false, "dropped_link"
		}
	}
	if !sameTagSequence(from, to) {
		return false, "markup_changed"
	}
	// An absolute the original did not claim is an invented claim, whether or not
	// the site has armed a pattern for it.
	for _, w := range superlativeRe.FindAllString(to, -1) {
		if !strings.Contains(strings.ToLower(from), strings.ToLower(w)) {
			return false, "invented_superlative"
		}
	}
	fromLower := strings.ToLower(from)
	for i, tok := range capTokenRe.FindAllString(to, -1) {
		if i == 0 && strings.HasPrefix(strings.TrimSpace(to), tok) {
			continue // sentence-initial capital is not a name
		}
		if !strings.Contains(fromLower, strings.ToLower(tok)) {
			return false, "invented_name"
		}
	}
	return true, ""
}

func neighbourCounts(s string) map[string]int {
	m := map[string]int{}
	for _, h := range ScanContrastNeighbours(s) {
		m[h.Shape]++
	}
	return m
}

// sameTagSequence compares markup by ORDER, not by count.
//
// A multiset comparison was the first version and it accepts inverted nesting:
// "<b><i>x</i></b>" and "<b><i>x</b></i>" hold the same tags in the same
// numbers, and the second is malformed. Tag counts equal is not markup
// preserved (council round 1, render_guardian seat). Order equality is not a
// well-formedness proof either, but it is strictly stronger and costs nothing —
// and any doubt keeps the original sentence, so the strict direction is free.
func sameTagSequence(a, b string) bool {
	ta := htmlTagRe.FindAllString(a, -1)
	tb := htmlTagRe.FindAllString(b, -1)
	if len(ta) != len(tb) {
		return false
	}
	for i := range ta {
		if !strings.EqualFold(ta[i], tb[i]) {
			return false
		}
	}
	return true
}

// NegationShapeNames is the shape vocabulary, for tests, findings and the
// register entry. Order is stable.
func NegationShapeNames() []string {
	out := make([]string, 0, len(negationShapes))
	for _, s := range negationShapes {
		out = append(out, s.shape)
	}
	return out
}

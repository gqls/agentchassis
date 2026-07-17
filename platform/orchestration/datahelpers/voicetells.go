// FILE: platform/orchestration/datahelpers/voicetells.go
//
// Deterministic voice-tells scanner (SPEC_voice_tells_check §3a): flags copy
// that reads machine-written, per site, against that site's voice_gate config.
// Companion to the claims layer and built on the same assertion-text extractor
// (ExtractAssertionText), so both layers agree by one implementation on what
// counts as prose vs markup/attributes.
//
// Every signal here is regex or arithmetic — no LLM. The prose lane (overall
// density, hollow fluency) belongs to the claims V3 auditor, not this file.
//
// Config lives on the site's `voice` spec under a `voice_gate` key. A site
// without one is skipped entirely (opt-in by presence, mirroring the claims
// layer's evidence_base). Long-form pages (blog posts, guides) use relaxed
// thresholds — essay rhythm differs from landing copy.
//
// What this scanner must NOT do (SPEC §4): reward errors or slang, judge
// quoted third-party text (blockquote is a non-assertion element upstream),
// or edit anything. Findings carry the offending snippet so a human sees WHAT
// tripped, never just a score.

package datahelpers

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// VoiceGateConfig is the machine-readable `voice_gate` block on a site's
// `voice` spec aspect. Zero-valued thresholds fall back to defaults —
// a site opts IN by having the block, not by filling every field.
type VoiceGateConfig struct {
	Enabled bool `json:"enabled"`

	// BannedPhrases are case-insensitive regexes (literal fallback on compile
	// error). CURATED for machine matching — deliberately separate from the
	// prose `banned_language` guidance in the voice spec, whose entries carry
	// human qualifiers ("unless followed by…") that regex cannot honour.
	BannedPhrases []VoicePhrase `json:"banned_phrases"`

	// Densities and distribution trips. Zero → default.
	EmDashPer1000Words float64 `json:"em_dash_per_1000_words"` // default 3
	TriadsPerPage      int     `json:"triads_per_page"`        // default 4
	LongSentenceShare  float64 `json:"long_sentence_share"`    // default 0.30
	LongSentenceWords  int     `json:"long_sentence_words"`    // default 25
	MeanSentenceWords  float64 `json:"mean_sentence_words"`    // default 22

	// ExpectContractions: on a plain-register site, a page with zero
	// contractions across many sentences reads stiff (the v1-leopardess tell).
	ExpectContractions bool `json:"expect_contractions"`
	MinSentencesForContractionCheck int `json:"min_sentences_for_contraction_check"` // default 15

	// LongForm relaxes thresholds for essay-shaped pages. Applied when the
	// caller says the page is long-form (page_type blog-post / guide).
	LongForm struct {
		EmDashPer1000Words float64 `json:"em_dash_per_1000_words"` // default 5
		TriadsPerPage      int     `json:"triads_per_page"`        // default 8
		LongSentenceShare  float64 `json:"long_sentence_share"`    // default 0.40
	} `json:"long_form"`
}

// VoicePhrase is one banned phrase: a pattern and the reason it is banned
// (shown to the human reviewer, same shape as claims BannedClaim).
type VoicePhrase struct {
	Pattern string `json:"pattern"`
	Reason  string `json:"reason"`
}

// VoiceGate is the compiled, runnable form of the config.
type VoiceGate struct {
	cfg      VoiceGateConfig
	banned   []compiledPhrase
	longForm bool
}

type compiledPhrase struct {
	re     *regexp.Regexp
	source string
	reason string
}

// VoiceFinding is one tell located in scanned copy.
type VoiceFinding struct {
	Check       string  `json:"check"` // banned_phrase | em_dash_density | triad_density | long_sentences | no_contractions | flourish_ending | strawman
	Matched     string  `json:"matched"`
	Reason      string  `json:"reason"`
	Snippet     string  `json:"snippet"`
	Value       float64 `json:"value,omitempty"`     // measured value for density/distribution checks
	Threshold   float64 `json:"threshold,omitempty"` // the trip level it exceeded
	Occurrences int     `json:"occurrences"`
}

// ParseVoiceGate reads a site's `voice` spec JSON and compiles its voice_gate
// block. Returns (nil, nil) when the spec has no enabled voice_gate — the
// opt-in signal, mirroring ParseEvidenceBase's contract.
func ParseVoiceGate(voiceSpecJSON []byte) (*VoiceGate, error) {
	var spec struct {
		VoiceGate *VoiceGateConfig `json:"voice_gate"`
	}
	if err := json.Unmarshal(voiceSpecJSON, &spec); err != nil {
		return nil, fmt.Errorf("voice spec parse: %w", err)
	}
	if spec.VoiceGate == nil || !spec.VoiceGate.Enabled {
		return nil, nil
	}
	g := &VoiceGate{cfg: *spec.VoiceGate}
	for _, p := range append(globalTellPhrases(), spec.VoiceGate.BannedPhrases...) {
		re, err := regexp.Compile("(?i)" + p.Pattern)
		if err != nil {
			re = regexp.MustCompile("(?i)" + regexp.QuoteMeta(p.Pattern))
		}
		g.banned = append(g.banned, compiledPhrase{re: re, source: p.Pattern, reason: p.Reason})
	}
	return g, nil
}

// globalTellPhrases is the fleet-wide AI-tells list — phrases that read
// machine-written on ANY register. Per-site bans layer on top via config.
func globalTellPhrases() []VoicePhrase {
	return []VoicePhrase{
		{Pattern: `\bdive (in|into)\b`, Reason: "AI-tell verb"},
		{Pattern: `\bunlock(ing)?\b`, Reason: "AI-tell hype verb"},
		{Pattern: `\bleverage\b`, Reason: "AI-tell verb (as a verb)"},
		{Pattern: `\bseamless(ly)?\b`, Reason: "AI-tell adjective"},
		{Pattern: `\bcutting[- ]edge\b`, Reason: "AI-tell adjective"},
		{Pattern: `\bgame[- ]chang(ing|er)\b`, Reason: "hype"},
		{Pattern: `\bin today'?s\b[^.]{0,40}\b(world|landscape|market|environment)\b`, Reason: "landscape preamble"},
		{Pattern: `\bwhether you'?re\b`, Reason: "reflexive audience-sweep opening"},
		{Pattern: `\bdelve\b`, Reason: "AI-tell verb"},
		{Pattern: `\btapestry\b`, Reason: "AI-tell noun"},
		{Pattern: `\bat the end of the day\b`, Reason: "dead phrase"},
		{Pattern: `\bbest practices\b`, Reason: "corporate filler"},
		{Pattern: `\bit'?s (important|worth) (to note|noting)\b`, Reason: "throat-clearing"},
	}
}

// Sentence splitting: terminator followed by whitespace + capital/quote, or
// end of block. Deliberately naive — this is a signal, not a parser.
var sentenceSplitRe = regexp.MustCompile(`[.!?]+(?:\s+|$)`)

var contractionRe = regexp.MustCompile(`(?i)\b\w+'(s|t|re|ve|ll|d|m)\b`)

// Triads: three balanced list items joined "X, Y(,) and Z" — each item 1–4
// words. The reflex pattern, not every legitimate three-item list; density
// thresholds absorb the difference.
var triadRe = regexp.MustCompile(`\b[\w-]+(?:\s+[\w-]+){0,3},\s+[\w-]+(?:\s+[\w-]+){0,3},?\s+and\s+[\w-]+`)

// Strawman shapes: "not X, but Y" and the staccato "Not a X. Not a Y."
var strawmanCommaRe = regexp.MustCompile(`(?i)\bnot (?:just |merely |simply |about )?[^.;:]{2,50},\s*but\b`)
var strawmanStaccatoRe = regexp.MustCompile(`(?i)\bnot an? [^.]{2,40}\.\s+not an? `)

// Flourish endings: block-final sentences that open with a summarising move.
var flourishRe = regexp.MustCompile(`(?i)^(that('| i)s (why|how|the point)|and that('| i)s\b|ultimately\b|in (short|summary|essence)\b|at its core\b|simply put\b)`)

func defaultF(v, d float64) float64 {
	if v > 0 {
		return v
	}
	return d
}
func defaultI(v, d int) int {
	if v > 0 {
		return v
	}
	return d
}

// ScanVoice runs every deterministic signal over the extracted assertion
// blocks of one page. longForm selects the relaxed threshold set.
func (g *VoiceGate) ScanVoice(blocks []string, longForm bool) []VoiceFinding {
	if g == nil || len(blocks) == 0 {
		return nil
	}
	emDashTrip := defaultF(g.cfg.EmDashPer1000Words, 3)
	triadTrip := defaultI(g.cfg.TriadsPerPage, 4)
	longShareTrip := defaultF(g.cfg.LongSentenceShare, 0.30)
	longWords := defaultI(g.cfg.LongSentenceWords, 25)
	meanTrip := defaultF(g.cfg.MeanSentenceWords, 22)
	minSentences := defaultI(g.cfg.MinSentencesForContractionCheck, 15)
	if longForm {
		emDashTrip = defaultF(g.cfg.LongForm.EmDashPer1000Words, 5)
		triadTrip = defaultI(g.cfg.LongForm.TriadsPerPage, 8)
		longShareTrip = defaultF(g.cfg.LongForm.LongSentenceShare, 0.40)
	}

	var findings []VoiceFinding
	var totalWords, totalSentences, longSentences, emDashes, triads, contractions int
	var sumSentenceWords int

	for _, block := range blocks {
		// Banned phrases — every hit is a finding with its snippet.
		for _, p := range g.banned {
			locs := p.re.FindAllStringIndex(block, -1)
			if len(locs) == 0 {
				continue
			}
			first := locs[0]
			findings = append(findings, VoiceFinding{
				Check: "banned_phrase", Matched: block[first[0]:first[1]],
				Reason: p.reason, Snippet: voiceSnippet(block, first[0], first[1]),
				Occurrences: len(locs),
			})
		}

		// Strawman shapes.
		for _, re := range []*regexp.Regexp{strawmanCommaRe, strawmanStaccatoRe} {
			if loc := re.FindStringIndex(block); loc != nil {
				findings = append(findings, VoiceFinding{
					Check: "strawman", Matched: block[loc[0]:loc[1]],
					Reason:  "defines by negation ('not X, but Y') — state what the thing is",
					Snippet: voiceSnippet(block, loc[0], loc[1]), Occurrences: 1,
				})
			}
		}

		// Density inputs.
		emDashes += strings.Count(block, "—")
		for _, t := range triadRe.FindAllString(block, -1) {
			triads++
			_ = t
		}
		contractions += len(contractionRe.FindAllString(block, -1))

		sentences := sentenceSplitRe.Split(block, -1)
		var lastNonEmpty string
		for _, s := range sentences {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			lastNonEmpty = s
			w := len(strings.Fields(s))
			totalSentences++
			sumSentenceWords += w
			totalWords += w
			if w > longWords {
				longSentences++
			}
		}
		if lastNonEmpty != "" && flourishRe.MatchString(lastNonEmpty) {
			findings = append(findings, VoiceFinding{
				Check: "flourish_ending", Matched: firstWords(lastNonEmpty, 8),
				Reason:  "summarising flourish closing a block — end on the last real thing said",
				Snippet: firstWords(lastNonEmpty, 20), Occurrences: 1,
			})
		}
	}

	// Page-level densities.
	if totalWords > 0 {
		density := float64(emDashes) / float64(totalWords) * 1000
		if density > emDashTrip {
			findings = append(findings, VoiceFinding{
				Check: "em_dash_density", Matched: "—",
				Reason:    "em-dash as a rhythm — use a full stop or a comma",
				Value:     density, Threshold: emDashTrip, Occurrences: emDashes,
			})
		}
	}
	if triads > triadTrip {
		findings = append(findings, VoiceFinding{
			Check: "triad_density", Matched: "X, Y, and Z",
			Reason:    "balanced three-item lists by reflex — two examples are usually enough",
			Value:     float64(triads), Threshold: float64(triadTrip), Occurrences: triads,
		})
	}
	if totalSentences > 0 {
		share := float64(longSentences) / float64(totalSentences)
		mean := float64(sumSentenceWords) / float64(totalSentences)
		if share > longShareTrip {
			findings = append(findings, VoiceFinding{
				Check: "long_sentences", Matched: fmt.Sprintf("%d of %d sentences over %d words", longSentences, totalSentences, longWords),
				Reason:    "dense sentences — one idea per sentence",
				Value:     share, Threshold: longShareTrip, Occurrences: longSentences,
			})
		}
		if mean > meanTrip {
			findings = append(findings, VoiceFinding{
				Check: "long_sentences", Matched: fmt.Sprintf("mean sentence length %.1f words", mean),
				Reason:    "average sentence too long for the register",
				Value:     mean, Threshold: meanTrip, Occurrences: totalSentences,
			})
		}
		if g.cfg.ExpectContractions && contractions == 0 && totalSentences >= minSentences {
			findings = append(findings, VoiceFinding{
				Check: "no_contractions", Matched: fmt.Sprintf("0 contractions in %d sentences", totalSentences),
				Reason:    "stiff register — the site's voice uses contractions (it's, we'd, you're)",
				Value:     0, Threshold: 1, Occurrences: 0,
			})
		}
	}
	return findings
}

func voiceSnippet(block string, start, end int) string {
	s := start - 60
	if s < 0 {
		s = 0
	}
	e := end + 60
	if e > len(block) {
		e = len(block)
	}
	return strings.TrimSpace(block[s:e])
}

func firstWords(s string, n int) string {
	f := strings.Fields(s)
	if len(f) > n {
		f = f[:n]
	}
	return strings.Join(f, " ")
}

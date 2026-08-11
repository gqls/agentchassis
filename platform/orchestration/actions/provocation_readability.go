// FILE: platform/orchestration/actions/provocation_readability.go
//
// A DETERMINISTIC readability measure for provocations, and the reason it is
// arithmetic rather than a judgement.
//
// WHY THIS EXISTS
// Owner, 2026-08-11, on the live provocation: "almost unreadable ... make sure the
// language is written to be readable by a 5 year old or something like that ... Cut
// out the long words perhaps use ASD-STE100."
//
// Measured the same afternoon across all 28 approved bodies: median Flesch-Kincaid
// grade 9.6 for the session-written entries and 10.4 for the generator's, with the
// pool's single worst entry (15.5, 34.5 words per sentence) being one the generator
// produced. The complained-about entry sits at 11.1 — eighth worst. **It is the house
// style, not an outlier**, which is why this is a rail and not a rewrite.
//
// WHY DETERMINISTIC, AND WHY IT IS NOT IN THE JUDGE PROMPT
// Sentence length and syllable counts are counting. They need no model, they cannot
// drift between runs, and they cost nothing. This lane has now paid three times in two
// days for the same lesson — the counter-case rule, British English, and this — that a
// rule stated only in a prompt is a request, not a control. Anything measurable belongs
// here beside the other corpus-derived checks; only what genuinely needs judgement
// (metaphor, irony, whether a sentence is *comprehensible*) stays with the model.
//
// WHAT THIS DELIBERATELY CANNOT SEE, and it is the owner's actual complaint
// He rejected the pool's PLAINEST entry — Flesch-Kincaid grade 5.9 — with "I don't even
// fully understand it". Its sentences are short and its words are ordinary; what makes
// it hard is that it is a metaphor the reader has to decode ("The dashboard is not an
// input to the decision. It is the receipt."). **No arithmetic finds that.** So this
// file bounds the mechanical half and the prompt carries the rest, and nobody should
// read a passing readability score as "the reader will understand it".
//
// ASD-STE100. The owner named Simplified Technical English. Its measurable rules are
// implemented here: one topic per sentence, a hard sentence-length ceiling, the active
// voice, short words. Its **approved dictionary of ~900 words is NOT implemented** — I
// do not have the list, and inventing one would produce a rail that enforces my guess
// at a published standard while carrying its name. That gap is stated rather than
// papered over.

package actions

import (
	"fmt"
	"strings"
	"unicode"
)

// STE's ceiling for descriptive (non-procedural) text is 25 words; its procedural
// ceiling is 20. Provocations are descriptive, but the owner's brief is well below
// STE's own bar, so the ceiling here is the tighter one and the AVERAGE target is
// tighter still. Measured for context: the pool's worst entry averages 34.5.
const (
	maxSentenceWords = 20
	maxAvgWords      = 15
	maxLongWordRatio = 0.12 // words of 3+ syllables
)

type readabilityReport struct {
	Sentences     int     `json:"sentences"`
	Words         int     `json:"words"`
	AvgWords      float64 `json:"avg_words_per_sentence"`
	LongestWords  int     `json:"longest_sentence_words"`
	LongestText   string  `json:"longest_sentence"`
	LongWordRatio float64 `json:"long_word_ratio"`
	Grade         float64 `json:"fk_grade"`
	Failures      []string
}

// splitSentences is deliberately crude: terminal punctuation only.
//
// It will treat "e.g." as a sentence break and undercount. That direction is the safe
// one — it makes sentences look SHORTER than they are, so the check under-reports
// rather than manufacturing failures. A check that invents problems gets switched off.
func splitSentences(text string) []string {
	var out []string
	var cur strings.Builder
	for _, r := range text {
		cur.WriteRune(r)
		if r == '.' || r == '!' || r == '?' {
			if s := strings.TrimSpace(cur.String()); len(s) > 1 {
				out = append(out, s)
			}
			cur.Reset()
		}
	}
	if s := strings.TrimSpace(cur.String()); len(s) > 1 {
		out = append(out, s)
	}
	return out
}

func words(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && r != '\''
	})
}

// countSyllables is the standard vowel-group heuristic. It is wrong on individual
// words ("queue" scores 2) and right in aggregate, which is all a ratio needs.
func countSyllables(w string) int {
	w = strings.ToLower(w)
	n, prevVowel := 0, false
	for _, r := range w {
		isVowel := strings.ContainsRune("aeiouy", r)
		if isVowel && !prevVowel {
			n++
		}
		prevVowel = isVowel
	}
	if strings.HasSuffix(w, "e") && n > 1 {
		n--
	}
	if n < 1 {
		n = 1
	}
	return n
}

// measureReadability reports; it does not decide. The caller chooses what is fatal.
func measureReadability(text string) readabilityReport {
	r := readabilityReport{}
	sents := splitSentences(text)
	r.Sentences = len(sents)
	if r.Sentences == 0 {
		return r
	}

	totalSyl, long := 0, 0
	for _, s := range sents {
		ws := words(s)
		if len(ws) > r.LongestWords {
			r.LongestWords = len(ws)
			r.LongestText = s
		}
		r.Words += len(ws)
		for _, w := range ws {
			syl := countSyllables(w)
			totalSyl += syl
			if syl >= 3 {
				long++
			}
		}
	}
	if r.Words == 0 {
		return r
	}

	r.AvgWords = float64(r.Words) / float64(r.Sentences)
	r.LongWordRatio = float64(long) / float64(r.Words)
	r.Grade = 0.39*r.AvgWords + 11.8*(float64(totalSyl)/float64(r.Words)) - 15.59

	if r.LongestWords > maxSentenceWords {
		r.Failures = append(r.Failures, fmt.Sprintf(
			"a sentence runs to %d words (limit %d): %q",
			r.LongestWords, maxSentenceWords, truncateForReason(r.LongestText)))
	}
	if r.AvgWords > maxAvgWords {
		r.Failures = append(r.Failures, fmt.Sprintf(
			"sentences average %.1f words (limit %.0f) — split them; one idea each",
			r.AvgWords, float64(maxAvgWords)))
	}
	if r.LongWordRatio > maxLongWordRatio {
		r.Failures = append(r.Failures, fmt.Sprintf(
			"%.0f%% of words are three syllables or more (limit %.0f%%) — use shorter ones",
			r.LongWordRatio*100, maxLongWordRatio*100))
	}
	return r
}

func truncateForReason(s string) string {
	if len(s) <= 120 {
		return s
	}
	return s[:117] + "..."
}

// checkReadability adds the measure to a verdict.
//
// ADVISORY FOR NOW, AND THAT IS A DECISION WITH A DATE ON IT.
// Every one of the 28 approved entries in the pool on 2026-08-11 fails at least one of
// these thresholds. Making it fatal on the day it ships would reject every candidate
// and starve the site, and it would do so before anyone had seen what the thresholds
// actually catch. So it records first: run some rounds, read the notes, then flip it.
//
// The flip is one line and the test below pins the CURRENT state, so making it fatal
// fails that test loudly rather than silently — which is the point. Do not flip it
// without a run showing the new prompt can pass it.
func checkReadability(c provocationCandidate, v *gateVerdict) {
	body := strings.TrimSpace(c.Body)
	if body == "" {
		return
	}
	r := measureReadability(body)
	if len(r.Failures) == 0 {
		return
	}
	v.note("form", "hard_to_read", fmt.Sprintf(
		"grade %.1f, %.1f words/sentence, longest %d — %s [ADVISORY: recorded, not fatal]",
		r.Grade, r.AvgWords, r.LongestWords, strings.Join(r.Failures, "; ")))
}

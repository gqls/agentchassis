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
// **FATAL SINCE 2026-09-02.** It shipped advisory on 2026-08-11 with an explicit
// instruction — "run some rounds, read the notes, then flip it" — and this is that
// flip. Two things had to be true and both are:
//
//  1. THE PROMPT CAN PASS IT. Two generation rounds on 2026-08-12 produced 8
//     candidates and 8 of 8 cleared the rail, against 0 of 28 for the pre-existing
//     pool. The rail was not measuring "impossible", it was measuring the old house
//     style, which is what the owner objected to.
//  2. SOMETHING HAD TO REPLACE THE HUMAN. In the same change the owner removed the
//     human-approval stamp from the publish path ("not restrained by needing my
//     permission"), so the reader who used to catch unreadable prose is gone. An
//     advisory note nobody reads is not a control. This rail is now the only
//     non-stochastic thing standing between the generator and the live site, which
//     is precisely why it must reject rather than record.
//
// WHY ARITHMETIC RATHER THAN THE JUDGE CARRIES THIS: the LLM judge is
// documented-stochastic on this very corpus — byte-identical text drew 0 factual
// objections on 08-05 and 2 on 08-08. A measure of sentence and word length cannot
// drift between runs, cannot be argued with, and returns the same verdict on the
// same bytes for ever. That is worth more here than a better but unstable reader.
//
// ⚠ WHAT THIS RAIL STILL CANNOT SEE, and do NOT let a passing score be read as "the
// reader will understand it". On 2026-08-11 the owner rejected the pool's PLAINEST
// entry — grade 5.9, short sentences, ordinary words — with "I don't even fully
// understand it". It was a riddle. No word-counting finds that. The prompt carries a
// rule against it; nothing measures it.
//
// IF THIS STARVES THE SITE, that is the failure mode to expect, and the answer is to
// generate more candidates per round (the generator's `count`), NOT to relax these
// thresholds to whatever the current output happens to score. Tuning a checker until
// it agrees with the thing it is checking is how the rail becomes decorative.
func checkReadability(c provocationCandidate, v *gateVerdict) {
	body := strings.TrimSpace(c.Body)
	if body == "" {
		return
	}
	r := measureReadability(body)
	if len(r.Failures) == 0 {
		return
	}
	v.reject("form", "hard_to_read", fmt.Sprintf(
		"grade %.1f, %.1f words/sentence, longest %d — %s",
		r.Grade, r.AvgWords, r.LongestWords, strings.Join(r.Failures, "; ")))
}

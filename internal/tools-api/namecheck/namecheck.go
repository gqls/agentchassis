// Package namecheck decides whether a block of argument text is safe to make
// PUBLIC — specifically, whether it makes a checkable allegation about an
// apparently-named third party.
//
// WHY IT EXISTS
// `architecture_review/RFC_020` §5.2. The Gauntlet publishes a visitor's own prose
// and an AI verdict at a permanent public URL. The owner's brief is low risk
// appetite, keep the viral effect — so the control has to target the harm rather
// than the feature.
//
// THE SHAPE OF THE PROBLEM, AND WHY A NAME DETECTOR ALONE IS THE WRONG ANSWER
// The obvious control is "refuse anything naming a real person". On this product
// that is close to refusing everything: the ruled audience is food, music and film
// canon, cities and generational habits (PLAN §11.1), where naming things IS the
// argument. "Christopher Nolan is overrated" names a real person and is pure
// opinion; blocking it costs the feature and buys nothing.
//
// What is actually dangerous is narrower and has a recognisable shape:
//
//	A NAMED ENTITY + A CHECKABLE ALLEGATION ABOUT IT.
//
// "Nolan is overrated" is a judgement. "Nolan stole the script" is an assertion of
// fact about an identifiable person, and it is the second that creates exposure.
// This is the same split the provocation gate already makes — the thesis is exempt
// by design, the factual assertions inside it are not.
//
// WHAT THIS IS AND IS NOT
// This is layer A: DETERMINISTIC, cheap, no network call, no model, and therefore
// unable to fail open. It is deliberately NOT the whole control. The natural layer
// B is a model judging the same question, exactly as the provocation gate stacks
// them — and until that exists this layer's misses are real. It is written to be
// conservative in the direction that matters: a false positive costs one visitor a
// share button, a false negative is the incident RFC_020 exists to prevent.
//
// It makes NO attempt to decide whether an allegation is TRUE, or whether a named
// entity is a real person. Both are out of reach here and neither changes the
// decision: publishing an unverifiable allegation about an apparent person is the
// thing being declined.
package namecheck

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// Finding is one reason a text was refused. Match is the literal span, so an
// operator reading a refusal can see what tripped it — a refusal whose reason
// cannot be inspected is one nobody can tune.
type Finding struct {
	Kind  string `json:"kind"`
	Match string `json:"match"`
	Term  string `json:"term,omitempty"`
}

// proximityWords is how close an allegation term must be to a named entity to
// count as being ABOUT it. Chosen for the length of an ordinary clause: far
// enough to span "Nolan, who by then had two hits, stole the script", short
// enough that an allegation in a different sentence about a different subject
// does not attach itself to whatever name appeared last.
const proximityWords = 12

var (
	// An honorific followed by a capitalised word. The strongest single signal
	// that a specific person is meant, and it survives lowercase surnames.
	reHonorific = regexp.MustCompile(
		`\b(?:Mr|Mrs|Ms|Miss|Dr|Prof|Professor|Sir|Dame|Lord|Lady|Rev|President|Senator|Chancellor|Mayor|Judge|Officer)\.?\s+[A-Z][\p{L}'’-]+`)

	// A named organisation by legal suffix. Companies can sue, and an allegation
	// about a business is the shape most likely to be both checkable and costly.
	reOrgSuffix = regexp.MustCompile(
		`\b[A-Z][\p{L}&'’.-]*(?:\s+[A-Z][\p{L}&'’.-]*)*\s+(?:Ltd|Limited|PLC|plc|LLC|Inc|Incorporated|GmbH|LLP|Group|Holdings)\b`)

	// A social handle names an account, which is an identifiable party.
	reHandle = regexp.MustCompile(`@[A-Za-z0-9_]{2,}`)
)

// opensSentence reports whether the byte at i begins a sentence, so that a token
// capitalised only by position ("The Guardian ran it") can be told from a name
// ("blamed The Guardian"). Used ONLY together with sentenceOpeners: position
// alone is not evidence, or "Christopher Nolan stole it" would be discarded for
// opening its own sentence — which is what the first version of this file did.
//
// It scans BACKWARDS past whitespace and opening punctuation, decoding runes
// rather than bytes because the skip set includes typographic quotes, which are
// multi-byte (a byte-wise version did not compile, which is how that was caught).
func opensSentence(text string, i int) bool {
	rest := text[:i]
	for len(rest) > 0 {
		r, size := utf8.DecodeLastRuneInString(rest)
		switch r {
		case ' ', '\t', '\n', '\r', '"', '\'', '(', '[', '“', '‘', '‛', '„':
			rest = rest[:len(rest)-size]
			continue
		case '.', '!', '?', ';', ':':
			return true
		default:
			return false
		}
	}
	return true // start of the text
}

// allegationTerms are words that turn a mention into a checkable claim about
// someone. Deliberately about CONDUCT and CHARACTER-AS-FACT, not about quality:
// "overrated", "boring", "derivative" and "terrible" are absent on purpose,
// because a site built for argument must permit them.
var allegationTerms = []string{
	"fraud", "fraudulent", "scam", "scammer", "con artist", "conman",
	"steal", "stole", "stolen", "steals", "stealing", "theft", "plagiarise", "plagiarize", "plagiarised", "plagiarized", "plagiarising", "plagiarist",
	"lied", "lying", "liar", "dishonest", "deceived",
	"corrupt", "corruption", "bribe", "bribed", "kickback",
	"criminal", "crook", "convicted", "arrested", "prosecuted", "guilty",
	"embezzle", "embezzled", "embezzlement", "laundering", "laundered", "launders",
	"defrauded", "defraud", "defrauding", "tax evasion", "evaded tax",
	"abuser", "abusive", "assaulted", "assault", "harassed", "harassment",
	"groomed", "grooming", "paedophile", "pedophile", "predator",
	"racist", "antisemitic", "homophobe", "homophobic", "misogynist", "bigot",
	"cheated", "cheating", "rigged", "faked", "faking", "forged", "forging",
	"falsified", "falsifying", "bribing", "harassing", "assaulting",
	"addict", "alcoholic", "junkie",
	"affair", "adulterer", "cheated on",
	"incompetent quack", "malpractice", "negligent",
}

// standaloneTerms are damaging enough that they are refused wherever they appear,
// with or without a detectable name. A text containing one of these is not an
// argument this product needs to publish, and requiring a name match first would
// let "he is a paedophile" through on a pronoun.
var standaloneTerms = []string{
	"paedophile", "pedophile", "groomed", "grooming", "predator",
	"rapist", "raped",
}

// ---------------------------------------------------------------------------
// Negation — added 2026-08-10 after the council's `reuse_agent` seat objected
// ---------------------------------------------------------------------------
//
// The seat was RIGHT and it found a real defect, not a process complaint: without
// this, "Nolan did not steal the script" was flagged as an allegation. That is a
// DEFENCE of the named person, and refusing to publish it is the opposite of what
// this package is for.
//
// THE ALGORITHM IS DELIBERATELY THE ONE ALREADY IN THIS ESTATE —
// `platform/orchestration/datahelpers.NegationGuard`, factored out under
// `bugs_open/222` with the explicit doctrine: "Two vocabularies, one algorithm."
// A bounded backwards window, trimmed to the current clause, tested against a cue
// regex. `check_tool_fabrication_action.go` already builds its own guard with its
// own cues on that sanctioned pattern, and this is the second.
//
// WHY THE TYPE IS NOT IMPORTED, measured rather than asserted: `datahelpers`
// drags goquery, cascadia and five tdewolff minify packages into a service that
// parses no HTML — 12+ heavy transitive dependencies for a struct of three
// fields, in a binary that ships to a single small VM by scp.
//
// ⚠ THAT LEAVES THE ALGORITHM DUPLICATED, and this comment is not a licence for
// it. The clean fix is extracting `NegationGuard` into a leaf package both sides
// can import; it is recorded as a follow-up in RFC_020 rather than done here,
// because moving a symbol out of `datahelpers` is a platform change with its own
// review. If you are reading this because you are about to write a THIRD copy:
// do the extraction instead.
var (
	// Allegation-appropriate cues. Deliberately NOT datahelpers' vocabulary,
	// which excludes bare "no"/"without" because in marketing prose those are
	// intensifiers. Here they are genuine negators — "no evidence Nolan stole
	// it" must not read as an allegation.
	negationCueRe = regexp.MustCompile(
		`(?i)\b(?:not|never|nor|cannot|no|without|denies|denied|deny|unfounded|baseless|` +
			`false(?:ly)?|untrue|disproven|debunked|acquitted|cleared|exonerated|` +
			`unable to|fails? to|failed to|refuses? to)\b` +
			`|[a-z]n['’‘]t\b` +
			`|\b(?:cant|dont|doesnt|didnt|isnt|arent|wasnt|werent|hasnt|havent|hadnt|wont|couldnt|shouldnt|wouldnt)\b`)

	// A cue must be in the SAME CLAUSE as the term, or "Nolan stole it, and no
	// one minds" would read as negated.
	negationBoundary = ".!?;:,<>\n\r\t\u2013\u2014"
)

// negationWindowBytes bounds the backwards scan. Wider than datahelpers' 64
// because an allegation cue can sit further from its negator ("there is no
// credible evidence that he ever plagiarised").
const negationWindowBytes = 96

// negated reports whether the term at byte offset pos is negated by a cue in the
// same clause. Multibyte-safe: the window is trimmed to a rune boundary, which is
// the bug datahelpers' comment records having cost real effort to get right.
func negated(text string, pos int) bool {
	start := pos - negationWindowBytes
	if start < 0 {
		start = 0
	}
	for start < pos && !utf8.RuneStart(text[start]) {
		start++
	}
	window := text[start:pos]
	if i := strings.LastIndexAny(window, negationBoundary); i >= 0 {
		window = window[i+1:]
	}
	return negationCueRe.MatchString(window)
}

// Scan reports why text must not be published, or nil if nothing was found.
//
// Nil means "this layer found nothing", NEVER "this text is safe" — see the
// package comment. The caller decides what to do with that distinction.
func Scan(text string) []Finding {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	var out []Finding
	lower := strings.ToLower(text)

	// 1. Standalone terms, regardless of any name.
	for _, t := range standaloneTerms {
		if i := strings.Index(lower, t); i >= 0 && !negated(text, i) {
			out = append(out, Finding{Kind: "standalone-allegation", Match: t, Term: t})
		}
	}

	// 2. Drive from the ALLEGATION, then look for a name near it.
	//
	// The loop runs this way round after the first version ran the other way and
	// was wrong twice. Driving from named entities meant (a) a name opening a
	// sentence was discarded along with "The Guardian…", so "Christopher Nolan
	// stole the script" passed, and (b) single-word names were never matched at
	// all, so "Nolan plagiarised it" passed. Both are exactly the case this
	// exists to catch.
	//
	// Inverting it also narrows the false-positive surface for free: capitalised
	// tokens are only ever examined in the neighbourhood of an allegation, so the
	// ordinary naming that IS the product is never inspected.
	for _, occ := range allegationOccurrences(lower) {
		if negated(text, occ.start) {
			continue
		}
		if name, ok := nameNear(text, occ, proximityWords); ok {
			out = append(out, Finding{Kind: "named-allegation", Match: name, Term: occ.term})
		}
	}
	return dedupe(out)
}

type occurrence struct {
	term  string
	start int
	end   int
}

func allegationOccurrences(lower string) []occurrence {
	var out []occurrence
	for _, t := range allegationTerms {
		from := 0
		for {
			i := strings.Index(lower[from:], t)
			if i < 0 {
				break
			}
			abs := from + i
			out = append(out, occurrence{term: t, start: abs, end: abs + len(t)})
			from = abs + len(t)
		}
	}
	return out
}

// sentenceOpeners are capitalised only because they begin a sentence. A token
// here, in that position, is not a name.
//
// This is an allow-list, which is the shape that hides what it was written to
// catch — so it is kept to grammar words that can never be a subject of an
// allegation, and it applies ONLY sentence-initially. Mid-sentence, every one of
// these is still eligible ("blamed The Guardian" keeps Guardian).
var sentenceOpeners = map[string]bool{
	"the": true, "this": true, "that": true, "these": true, "those": true,
	"it": true, "he": true, "she": true, "they": true, "we": true, "you": true,
	"a": true, "an": true, "and": true, "but": true, "so": true, "if": true,
	"when": true, "where": true, "why": true, "how": true, "what": true,
	"every": true, "most": true, "some": true, "all": true, "no": true,
	"nobody": true, "everyone": true, "everybody": true, "after": true,
	"before": true, "in": true, "on": true, "at": true, "for": true, "to": true,
	"from": true, "by": true, "with": true, "without": true, "because": true,
	"since": true, "although": true, "while": true, "yes": true, "perhaps": true,
	"maybe": true, "actually": true, "honestly": true, "frankly": true,
	"given": true, "once": true, "now": true, "then": true, "there": true,
	"here": true, "our": true, "my": true, "your": true, "his": true,
	"her": true, "their": true, "its": true, "i": true, "as": true, "of": true,
}

var reToken = regexp.MustCompile(`[\p{L}][\p{L}'’-]*|@[A-Za-z0-9_]{2,}`)

// nameNear reports the first apparent name within proximityWords of an
// allegation. A "name" is a capitalised token, an @handle, or an honorific-led
// or legally-suffixed span — minus tokens capitalised only by sentence position.
func nameNear(text string, occ occurrence, n int) (string, bool) {
	lo, hi := windowBounds(text, occ, n)
	window := text[lo:hi]

	// Strong signals first: these are unambiguous even where capitalisation is not.
	for _, re := range []*regexp.Regexp{reHonorific, reOrgSuffix, reHandle} {
		if m := re.FindString(window); m != "" {
			return m, true
		}
	}

	for _, loc := range reToken.FindAllStringIndex(window, -1) {
		tok := window[loc[0]:loc[1]]
		r, _ := utf8.DecodeRuneInString(tok)
		if !isUpper(r) {
			continue
		}
		// Capitalised only because it opens a sentence?
		if opensSentence(text, lo+loc[0]) && sentenceOpeners[strings.ToLower(tok)] {
			continue
		}
		return tok, true
	}
	return "", false
}

func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z' || (r > 127 && strings.ToUpper(string(r)) == string(r) && strings.ToLower(string(r)) != string(r))
}

// windowBounds returns byte offsets spanning n words either side of occ.
func windowBounds(text string, occ occurrence, n int) (int, int) {
	lo := occ.start
	for w := 0; w < n && lo > 0; w++ {
		j := strings.LastIndexAny(strings.TrimRight(text[:lo], " \t\n\r"), " \t\n\r")
		if j < 0 {
			lo = 0
			break
		}
		lo = j
	}
	hi := occ.end
	for w := 0; w < n && hi < len(text); w++ {
		rel := strings.IndexAny(strings.TrimLeft(text[hi:], " \t\n\r"), " \t\n\r")
		if rel < 0 {
			hi = len(text)
			break
		}
		skipped := len(text[hi:]) - len(strings.TrimLeft(text[hi:], " \t\n\r"))
		hi = hi + skipped + rel
	}
	return lo, hi
}

// ScanAll scans several texts as one decision. Used for a round, whose risk is
// spread across the visitor's prose AND the model's own output — the verdict is
// the service's text, not the visitor's, so it is checked on exactly the same
// terms rather than trusted for being ours (RFC_020 §1.4).
func ScanAll(texts ...string) []Finding {
	var out []Finding
	for _, t := range texts {
		out = append(out, Scan(t)...)
	}
	return dedupe(out)
}

func dedupe(in []Finding) []Finding {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[Finding]bool, len(in))
	out := make([]Finding, 0, len(in))
	for _, f := range in {
		if seen[f] {
			continue
		}
		seen[f] = true
		out = append(out, f)
	}
	return out
}

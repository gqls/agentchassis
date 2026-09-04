// FILE: platform/delivery/vocabulary.go
//
// The customer-letter placeholder vocabulary, and the guard derived from it.
//
// WHY THIS EXISTS (bugs_open/475; council c8ed56d2, approved round 2).
// Until this file, the token set lived in send_delivery_email_action.go's
// fillTemplate (a strings.Replacer over a closed vocabulary) and the rule that
// a named token must actually be producible lived in a hand-written slice in
// EACH sender. Two senders, one vocabulary, two mirrors — and the contract
// keeping them in step was a COMMENT in one of them:
//
//	"If fillTemplate's vocabulary grows, this list must grow with it."
//
// OWNER RULING 2026-08-02 §2 already ruled on that shape: "A comment is not a
// control on a tree this many sessions share." It was demonstrated on
// 2026-09-04, when two sessions independently chose two names for one new
// placeholder ({{instructions_link}} / {{instructions_url}}) and it was caught
// by a cross-session message rather than by anything mechanical.
//
// So: ONE declaration. Apply substitutes from the same map Check validated, and
// a token cannot exist in the filler while being absent from the guard.
//
// FOR A FUTURE ADOPTER (architecture seat's advisory, same round): the Check
// ordering and the NeverReason/FromClaim semantics were settled in council
// round c8ed56d2 — read that verdict rather than reverse-engineering intent
// from the types. The follow-up sender (bugs_open/477) is the second caller and
// converts on its own lane's commit; ADOPTION IS PER-CALLER BY DESIGN, copying
// ActionInputSpec's model (opt-in, "driven by the coverage report, not by a
// flag day"), because a fleet-wide flag day for a validator is how you turn an
// inert defect into an outage — see datahelpers/action_inputs.go and
// bugs_open/101 D2.
//
// WHAT THIS IS NOT. It checks PLACEHOLDERS, never PROSE. The sentence that
// started bugs_open/475 — "The ZIP comes with instructions" — carries no
// placeholder and is invisible here. This narrows the class by making
// customer-facing facts into checked links; it does not eliminate it.
//
// PRIOR ART, checked (reuse_agent + prior_art seats, round 2). Two adjacent
// mechanisms exist and neither is reused:
//   - datahelpers.ActionInputSpec (ConfigKeys/CheckConfig) — declared STEP-CONFIG
//     KEYS per action definition, checked per definition.
//   - content_components.input_schema + missingRequiredValueFields /
//     missingRequiredLLMFields — declared TEMPLATE FIELDS per component, checked
//     against content_data at render/discovery time, remedied by a work item.
//
// Both answer "does the DATA supply what the TEMPLATE declared". This answers a
// different question — "do the CODE and the CONFIG agree about which tokens
// exist" — and it must answer it BEFORE an irreversible act rather than as an
// audit afterwards. The reuse_agent seat's wider point stands and is recorded
// rather than dismissed: this is the THIRD declared-vocabulary-plus-coverage
// mechanism on the platform, and generalising the three is a real candidate —
// but they differ in unit, lifetime and consequence, so it is architecture-scope
// work and not a rider on a customer-facing bug fix.
package delivery

import (
	"fmt"
	"sort"
	"strings"
)

// Token is one placeholder a customer letter may name. The set is CLOSED: a
// template author cannot invent one, and the estate cannot grow one without
// every ADOPTING sender saying what it does with it (see Fill.Check).
//
// Named Token, not Link, deliberately. The members do NOT share a provenance
// and they are not all links: TokenDays is AdvertisedLiveWindowDays, a
// compile-time constant whose value is deliberately NOT the obvious one (we
// advertise 30 while LiveLinkWindow serves 42 — see handover.go, and never
// derive it from an expiry). A "placeholder -> link" shape would have dropped
// precisely the entry whose correct value a reader would guess wrong.
type Token string

const (
	TokenLiveSite     Token = "{{live_site}}"
	TokenConfirmLink  Token = "{{confirm_link}}"
	TokenZipLink      Token = "{{zip_link}}"
	TokenInstructions Token = "{{instructions_link}}"
	TokenDomainRent   Token = "{{domain_rent_link}}"
	TokenDomainBuy    Token = "{{domain_buy_link}}"
	TokenStripePortal Token = "{{stripe_portal_link}}"
	TokenDays         Token = "{{days}}"
)

// Vocabulary is the single source of truth. Both Apply and every adopting
// sender's pre-claim guard are derived from it, so they cannot disagree.
//
// NOTE ON NAMING, because it reads like a slip and is not: the placeholder is
// {{instructions_link}} while its step-config KEY is "instructions_url". The
// estate pairs a *_url key with a *_link placeholder three times already
// (domain_rent, domain_buy, stripe_portal). Renaming the key would have made it
// the first *_link config key in the estate.
var Vocabulary = []Token{
	TokenLiveSite,
	TokenConfirmLink,
	TokenZipLink,
	TokenInstructions,
	TokenDomainRent,
	TokenDomainBuy,
	TokenStripePortal,
	TokenDays,
}

// Availability is what ONE sender says about ONE token on THIS dispatch.
//
// VALUES COME FROM THE CALLER. The same token resolves differently in different
// senders — {{live_site}} is step config in one and an input in the other;
// {{confirm_link}} is minted inside Claim in one and minted by the caller after
// its own claim in the other — so nothing about provenance belongs in the
// Vocabulary. (bugs_open/477 lane, 2026-09-04: an earlier draft modelled
// provenance per token and could not have expressed their caller at all.)
type Availability struct {
	// Value is the resolved substitution. Empty means this dispatch has none.
	Value string

	// Source names where the value should have come from, for the refusal text
	// ("zip_presigned_url", "domain_rent_url config"). Operators read this.
	Source string

	// NeverReason, when set, means this sender can NEVER produce this token, at
	// any time, BY CONSTRUCTION — not "it happens to be empty today". The
	// worked case is {{zip_link}} in the scheduled follow-up sender: it has no
	// zip step and no presign to mint, so the token in its template is an
	// AUTHOR ERROR to catch at dispatch, not a value someone might supply.
	//
	// The reason string is the point. Flattening this to "empty => refuse"
	// behaves identically today and loses the sentence that stops a later
	// session "fixing" it by wiring a presign into a scheduled follow-up.
	NeverReason string

	// FromClaim marks a token whose value is PRODUCED BY the caller's claim and
	// is therefore unknowable when Check runs. {{confirm_link}} is the case:
	// the token is minted inside Claim, so pre-claim it legitimately has no
	// value, and refusing on that would refuse every delivery.
	//
	// Check exempts these from the empty-value refusal; Apply then REFUSES if
	// one is still empty. Without that second half a placeholder would ship to
	// a customer as a literal or a blank — the exact defect this file exists to
	// prevent, arriving through the exemption. (editquality seat, round 2: the
	// mutation step was unspecified in the submitted sketch.)
	FromClaim bool
}

// Fill is a sender's complete declaration over the Vocabulary for one dispatch.
type Fill map[Token]Availability

// Check refuses BEFORE the caller's irreversible statement (delivery.Claim, or
// ClaimFollowup — each caller has its own, which is why ordering is asserted at
// the CALLER and not here).
//
// Four refusals, in this order, each catching something the next cannot:
//
//  1. MALFORMED — Value and NeverReason both set. A caller asserting a token is
//     both unproducible-by-construction and has a value is confused; say so at
//     once rather than silently picking one.
//  2. UNKNOWN — a {{...}} in the template that the Vocabulary does not contain.
//     This is what moves an ahead-of-the-binary template's failure to BEFORE
//     the stamp. A body_template is CONFIG and goes live on apply, while the
//     vocabulary is compiled in; without this, an unknown token survives the
//     fill, trips the caller's post-fill "{{" scan, and that scan runs AFTER
//     the claim — burning a handover, or in the follow-up sender the customer's
//     only follow-up. See LANDMINES.md.
//  3. COVERAGE — a Vocabulary token this Fill does not declare. Unconditional:
//     it does NOT depend on the template naming the token, so growing the
//     Vocabulary makes an unteached adopter refuse loudly and pre-claim rather
//     than latently.
//  4. AVAILABILITY — a token the template NAMES whose value this dispatch
//     cannot produce (empty, or NeverReason set).
//
// ⚠ ACCEPTED TRADE-OFF (guardian seat, round 2): a refusal here BLOCKS a paid
// customer's delivery, where the previous behaviour would have shipped the
// email with an empty string in it. That is deliberate — a blocked delivery is
// visible, recoverable and pre-stamp, whereas the empty string is invisible,
// reaches the customer, and no post-fill scan can see it because the fill
// succeeded. But it IS a behavioural change on a money-adjacent path, so the
// refusals above are ordered to fire on author error (1-3) before dispatch
// state (4), and every message names what to fix.
func (f Fill) Check(template string) error {
	// 1. MALFORMED.
	for _, tok := range sortedTokens(f) {
		a := f[tok]
		if a.NeverReason != "" && a.Value != "" {
			return fmt.Errorf("delivery vocabulary: %s is declared with BOTH a value (%q) and a never-reason (%q); a sender that can never produce a token must not also supply one — fix the sender's Fill", tok, a.Value, a.NeverReason)
		}
	}

	// 2. UNKNOWN — a token in the template the compiled vocabulary lacks.
	if unknown := unknownTokens(template); len(unknown) > 0 {
		return fmt.Errorf("delivery vocabulary: body_template names %s, which this binary's vocabulary does not contain. The template is CONFIG (live on apply) and the vocabulary is COMPILED IN, so this usually means the template was migrated ahead of the image that knows the token — check the running build with service_binary_capabilities and roll before applying. Nothing was stamped", strings.Join(unknown, ", "))
	}

	// 3. COVERAGE — every Vocabulary token must be declared by this sender.
	var undeclared []string
	for _, tok := range Vocabulary {
		if _, ok := f[tok]; !ok {
			undeclared = append(undeclared, string(tok))
		}
	}
	if len(undeclared) > 0 {
		return fmt.Errorf("delivery vocabulary: this sender does not declare %s. Every token in delivery.Vocabulary needs an Availability — supply a value, or a NeverReason saying why this sender can never produce it. Nothing was stamped", strings.Join(undeclared, ", "))
	}

	// 4. AVAILABILITY — only for tokens the template actually names.
	for _, tok := range sortedTokens(f) {
		a := f[tok]
		if !strings.Contains(template, string(tok)) {
			continue
		}
		if a.NeverReason != "" {
			return fmt.Errorf("delivery vocabulary: body_template names %s but this sender can never produce it: %s. Nothing was stamped — remove it from the template", tok, a.NeverReason)
		}
		if a.Value == "" && !a.FromClaim {
			src := a.Source
			if src == "" {
				src = "its source"
			}
			return fmt.Errorf("delivery vocabulary: body_template names %s but %s is empty: the email would carry a blank where a link should be. Nothing was stamped — fix the template or supply the link and re-dispatch", tok, src)
		}
	}
	return nil
}

// CoverageErrors reports how a sender's Fill fails to match the Vocabulary:
// tokens it does not declare, and tokens it invents. Empty means exact.
//
// Exported and PURE (no testing import in production code) so that every
// adopting sender's own test package can assert coverage directly. That
// assertion is the mechanism this file exists for — without a caller it would
// be a guarantee nobody checks. Check performs the same coverage test at
// dispatch; this is the same question asked in CI, where it is free.
func (f Fill) CoverageErrors() []string {
	var out []string
	for _, tok := range Vocabulary {
		if _, ok := f[tok]; !ok {
			out = append(out, "does not declare "+string(tok))
		}
	}
	known := make(map[Token]bool, len(Vocabulary))
	for _, tok := range Vocabulary {
		known[tok] = true
	}
	for _, tok := range sortedTokens(f) {
		if !known[tok] {
			out = append(out, "declares "+string(tok)+", which is not in the Vocabulary")
		}
	}
	return out
}

// Claimed fills in the tokens marked FromClaim, once the caller's claim has
// produced them. Returns a new Fill; the receiver is not mutated, so a caller
// cannot accidentally leave Check and Apply reading different maps.
//
// A token not marked FromClaim is rejected: this is the post-claim step, and
// quietly overwriting a pre-claim value here would reintroduce exactly the
// "Check validated one thing and Apply substituted another" gap the file
// exists to close.
func (f Fill) Claimed(values map[Token]string) (Fill, error) {
	out := make(Fill, len(f))
	for tok, a := range f {
		out[tok] = a
	}
	for tok, v := range values {
		a, ok := out[tok]
		if !ok {
			return nil, fmt.Errorf("delivery vocabulary: Claimed given %s, which this sender never declared", tok)
		}
		if !a.FromClaim {
			return nil, fmt.Errorf("delivery vocabulary: Claimed given %s, which is not marked FromClaim — post-claim values may only fill tokens the claim produces", tok)
		}
		a.Value = v
		out[tok] = a
	}
	return out, nil
}

// Apply substitutes the vocabulary into the template. It reads the SAME map
// Check validated, which is the whole point of the file.
//
// It returns an error rather than a bare string because of the FromClaim
// exemption: Check deliberately lets a claim-produced token through with no
// value, so something must assert that the claim actually produced it. That
// assertion is here, and it is the last gate before a customer sees the text.
func (f Fill) Apply(template string) (string, error) {
	var stillEmpty []string
	pairs := make([]string, 0, len(f)*2)
	for _, tok := range sortedTokens(f) {
		a := f[tok]
		if a.FromClaim && a.Value == "" && strings.Contains(template, string(tok)) {
			stillEmpty = append(stillEmpty, string(tok))
		}
		pairs = append(pairs, string(tok), a.Value)
	}
	if len(stillEmpty) > 0 {
		return "", fmt.Errorf("delivery vocabulary: %s is produced by the claim but is still empty at fill time — the claim did not supply it, and substituting would put a blank in a customer's letter. Call Fill.Claimed with the claim's outputs before Apply", strings.Join(stillEmpty, ", "))
	}
	return strings.NewReplacer(pairs...).Replace(template), nil
}

// unknownTokens returns every {{...}} in the template that is not in the
// Vocabulary, deduplicated and sorted.
//
// Deliberately a scan for the {{...}} SHAPE rather than a check of the known
// tokens: the question is "what does this template name that we do not know",
// and only a shape scan can answer it. A loop over Vocabulary answers the
// opposite question and would return nothing for a template full of inventions.
func unknownTokens(template string) []string {
	known := make(map[string]bool, len(Vocabulary))
	for _, t := range Vocabulary {
		known[string(t)] = true
	}
	seen := make(map[string]bool)
	var out []string
	rest := template
	for {
		i := strings.Index(rest, "{{")
		if i < 0 {
			break
		}
		j := strings.Index(rest[i:], "}}")
		if j < 0 {
			// An unterminated "{{" cannot be a token; the caller's post-fill
			// scan is what catches a malformed template.
			break
		}
		tok := rest[i : i+j+2]
		if !known[tok] && !seen[tok] {
			seen[tok] = true
			out = append(out, tok)
		}
		rest = rest[i+j+2:]
	}
	sort.Strings(out)
	return out
}

// sortedTokens gives every iteration over a Fill a stable order, so a refusal
// message names the same token every time rather than whichever the map yielded
// first. A test asserting on error text is otherwise flaky by construction.
func sortedTokens(f Fill) []Token {
	out := make([]Token, 0, len(f))
	for tok := range f {
		out = append(out, tok)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

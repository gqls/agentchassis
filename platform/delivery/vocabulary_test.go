// FILE: platform/delivery/vocabulary_test.go
//
// The control that makes vocabulary drift fail in CI rather than in a
// customer's inbox.
//
// ⚠ THE ORDERING TEST IS NOT HERE, AND THAT IS THE POINT (editquality seat,
// council c8ed56d2 round 2). The submitted plan proposed proving "Check runs
// before the claim" in this package with sqlmock and no expectations. That test
// would have been INERT: Fill.Check takes no database handle, so it cannot
// touch the database by construction, and mock.ExpectationsWereMet() with zero
// expectations is satisfied whatever happened — including by a caller that
// claims first. "Pre-claim" is a property of a CALLER, so it is asserted at
// each caller against that caller's own irreversible statement
// (delivery.Claim in send_delivery_email_action.go; ClaimFollowup in the
// follow-up sender). See TestSendDeliveryEmailChecksVocabularyBeforeClaiming.
package delivery

import (
	"strings"
	"testing"
)

// AssertCoversVocabulary fails if a sender's Fill omits a Vocabulary token or
// invents one. Every ADOPTING sender's test calls it; that call is what turns a
// vocabulary addition red for a sender that has not been taught.
func AssertCoversVocabulary(t *testing.T, f Fill) {
	t.Helper()
	for _, problem := range f.CoverageErrors() {
		t.Errorf("sender %s", problem)
	}
}

// AssertNeverProduces asserts a sender has declared a token unproducible BY
// CONSTRUCTION, with a reason.
//
// Offered for the architecture seat's objection (round 2) that NeverReason
// otherwise relies on a human reading a string correctly for ever. The
// follow-up sender's {{zip_link}} is the case: a session "fixing" it by wiring
// a presign into a scheduled follow-up turns this red instead of reading past
// a sentence.
func AssertNeverProduces(t *testing.T, f Fill, tok Token) {
	t.Helper()
	a, ok := f[tok]
	if !ok {
		t.Fatalf("sender does not declare %s at all", tok)
	}
	if a.NeverReason == "" {
		t.Errorf("%s should be declared unproducible-by-construction, but carries no NeverReason", tok)
	}
	if a.Value != "" {
		t.Errorf("%s carries a value (%q) as well as a never-reason — see Fill.Check's MALFORMED refusal", tok, a.Value)
	}
}

// completeFill is a Fill covering the whole Vocabulary, for tests that want to
// vary one entry. Deliberately built from Vocabulary rather than a literal: a
// hand-written literal here would need updating alongside every addition, which
// is the defect this package exists to remove.
func completeFill() Fill {
	f := make(Fill, len(Vocabulary))
	for _, tok := range Vocabulary {
		f[tok] = Availability{Value: "x", Source: "test"}
	}
	return f
}

func TestCheckRefusesATokenTheVocabularyDoesNotKnow(t *testing.T) {
	// The config-ahead-of-the-binary case: a body_template migrated before the
	// image that knows the token. Before this refusal the token survived the
	// fill and tripped the caller's post-fill "{{" scan, which runs AFTER the
	// claim has stamped. See LANDMINES.md.
	err := completeFill().Check("Your site {{live_site}} and {{some_new_token}}")
	if err == nil {
		t.Fatal("expected a refusal for an unknown token")
	}
	if !strings.Contains(err.Error(), "{{some_new_token}}") {
		t.Errorf("refusal must name the offending token, got: %v", err)
	}
	if !strings.Contains(err.Error(), "Nothing was stamped") {
		t.Errorf("refusal must tell the operator nothing was stamped, got: %v", err)
	}
}

func TestCheckRefusesAnUndeclaredVocabularyToken(t *testing.T) {
	// COVERAGE, and it is unconditional: the template does not name the missing
	// token, and the refusal fires anyway. This is what makes growing the
	// Vocabulary loud for an unteached sender rather than latent.
	f := completeFill()
	delete(f, TokenDays)
	err := f.Check("Your site is live at {{live_site}}")
	if err == nil {
		t.Fatal("expected a coverage refusal for an undeclared token")
	}
	if !strings.Contains(err.Error(), string(TokenDays)) {
		t.Errorf("refusal must name the undeclared token, got: %v", err)
	}
}

func TestCheckRefusesBothValueAndNeverReason(t *testing.T) {
	f := completeFill()
	f[TokenZipLink] = Availability{Value: "https://x", NeverReason: "no presign here"}
	err := f.Check("{{zip_link}}")
	if err == nil || !strings.Contains(err.Error(), "BOTH") {
		t.Fatalf("expected a MALFORMED refusal, got: %v", err)
	}
}

func TestCheckRefusesANamedTokenThisSenderCanNeverProduce(t *testing.T) {
	f := completeFill()
	f[TokenZipLink] = Availability{NeverReason: "a scheduled follow-up has no step that can mint a presign"}
	err := f.Check("Your files: {{zip_link}}")
	if err == nil {
		t.Fatal("expected a refusal for a never-producible token")
	}
	if !strings.Contains(err.Error(), "no step that can mint a presign") {
		t.Errorf("the REASON must reach the operator, not just the token: %v", err)
	}
}

func TestCheckIgnoresANeverProducibleTokenTheTemplateDoesNotName(t *testing.T) {
	// A sender may legitimately declare a token unproducible for ever; that is
	// only an error if a template actually names it.
	f := completeFill()
	f[TokenZipLink] = Availability{NeverReason: "no zip step"}
	if err := f.Check("Your site is live at {{live_site}}"); err != nil {
		t.Fatalf("a never-producible token the template does not name must not refuse: %v", err)
	}
}

func TestCheckRefusesANamedTokenWithNoValue(t *testing.T) {
	f := completeFill()
	f[TokenDomainRent] = Availability{Source: "domain_rent_url config"}
	err := f.Check("Rent: {{domain_rent_link}}")
	if err == nil || !strings.Contains(err.Error(), "domain_rent_url config") {
		t.Fatalf("refusal must name the SOURCE the operator has to fix, got: %v", err)
	}
}

func TestFromClaimTokenPassesCheckAndIsCaughtByApply(t *testing.T) {
	// The hole the editquality seat found in the submitted sketch:
	// {{confirm_link}} is minted INSIDE the claim, so it is legitimately empty
	// when Check runs. If nothing asserted it afterwards, a literal or a blank
	// would ship to the customer through the very exemption that lets delivery
	// work at all.
	f := completeFill()
	f[TokenConfirmLink] = Availability{FromClaim: true}
	tpl := "Press here when you have moved: {{confirm_link}}"

	if err := f.Check(tpl); err != nil {
		t.Fatalf("a claim-produced token must pass Check pre-claim: %v", err)
	}
	if _, err := f.Apply(tpl); err == nil {
		t.Fatal("Apply must REFUSE a claim-produced token the claim never supplied")
	}

	filled, err := f.Claimed(map[Token]string{TokenConfirmLink: "https://links.example/c/tok"})
	if err != nil {
		t.Fatalf("Claimed: %v", err)
	}
	body, err := filled.Apply(tpl)
	if err != nil {
		t.Fatalf("Apply after Claimed: %v", err)
	}
	if !strings.Contains(body, "https://links.example/c/tok") {
		t.Errorf("the claimed value must reach the body, got %q", body)
	}
	if strings.Contains(body, "{{") {
		t.Errorf("no placeholder may survive, got %q", body)
	}
}

func TestClaimedRefusesATokenNotProducedByTheClaim(t *testing.T) {
	// Quietly overwriting a pre-claim value here would reintroduce the exact
	// "Check validated one thing, Apply substituted another" gap.
	f := completeFill()
	if _, err := f.Claimed(map[Token]string{TokenDomainRent: "https://pay"}); err == nil {
		t.Fatal("Claimed must refuse a token not marked FromClaim")
	}
}

func TestApplyAndCheckReadTheSameTokenSet(t *testing.T) {
	// The invariant, asserted directly rather than trusted: anything Check
	// accepts, Apply substitutes completely. If a future edit adds a token to
	// Vocabulary and forgets Apply, this fails.
	f := completeFill()
	var b strings.Builder
	for _, tok := range Vocabulary {
		b.WriteString(string(tok))
		b.WriteString(" ")
	}
	tpl := b.String()
	if err := f.Check(tpl); err != nil {
		t.Fatalf("a template naming every token, with every token supplied, must pass: %v", err)
	}
	body, err := f.Apply(tpl)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if strings.Contains(body, "{{") {
		t.Errorf("Apply left a placeholder standing — Vocabulary and Apply have drifted: %q", body)
	}
}

func TestUnknownTokensToleratesAnUnterminatedBrace(t *testing.T) {
	// A malformed template must not panic or loop; the caller's post-fill scan
	// owns that case.
	if got := unknownTokens("hello {{ world"); len(got) != 0 {
		t.Errorf("an unterminated {{ is not a token, got %v", got)
	}
}

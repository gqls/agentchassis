// FILE: paired_test.go
//
// These tests exist because "the positions are sealed" is a CLAIM ABOUT
// BEHAVIOUR, and this codebase has a standing rule that such a claim is not
// evidence until something runs. A comment saying the seal holds is worth
// nothing; a test that tries to read a colleague's answer and fails to is
// worth something.
//
// The tests are written to BREAK the seal, not to demonstrate it working.
package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

var (
	t0 = time.Date(2026, 7, 31, 9, 0, 0, 0, time.UTC)
	// Distinctive so a substring search cannot accidentally pass — the
	// lesson from WRONG_CALLS 2026-07-31, where "Nobody actually" matched
	// two different provocations and I believed the wrong one.
	secretA = "ZZALICE-SECRET-POSITION-QQ"
	secretB = "ZZBOB-SECRET-POSITION-QQ"
	secretC = "ZZCAROL-SECRET-POSITION-QQ"
)

func threeWay(t *testing.T, rule RevealRule, quorum int, deadline time.Time) (*Session, string, string, string) {
	t.Helper()
	s, err := NewSession("Fran", "Remote work killed mentorship.",
		[]string{"Alice", "Bob", "Carol"}, rule, quorum, deadline, t0)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	return s, s.participants[0].Token, s.participants[1].Token, s.participants[2].Token
}

// ---------------------------------------------------------------------------
// The seal
// ---------------------------------------------------------------------------

// The central test. Two people have committed; the third has not, so the
// session is still sealed. Serialise exactly what a participant's page would
// be built from and assert nobody else's words are in the bytes.
func TestSealedViewBytesCannotContainAnotherPosition(t *testing.T) {
	s, ta, tb, tc := threeWay(t, RevealWhenAllCommitted, 0, time.Time{})

	if err := s.Commit(ta, secretA, t0); err != nil {
		t.Fatal(err)
	}
	if err := s.Commit(tb, secretB, t0.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}

	for name, tok := range map[string]string{"Alice": ta, "Bob": tb, "Carol": tc} {
		sealed, revealed, err := s.ViewFor(tok, t0.Add(2*time.Minute))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if revealed != nil {
			t.Fatalf("%s got a RevealedView while one participant has not committed", name)
		}
		blob, err := json.Marshal(sealed)
		if err != nil {
			t.Fatal(err)
		}
		body := string(blob)

		// Everyone else's secret must be absent from your bytes.
		others := map[string]string{"Alice": secretA, "Bob": secretB, "Carol": secretC}
		delete(others, name)
		for who, secret := range others {
			if strings.Contains(body, secret) {
				t.Errorf("SEAL BROKEN: %s's sealed view contains %s's position: %s", name, who, body)
			}
		}
	}

	// ...and your own position IS returned to you, which is not a leak.
	sealed, _, _ := s.ViewFor(ta, t0.Add(2*time.Minute))
	if sealed.YourPosition != secretA {
		t.Errorf("Alice cannot see her own position back: %q", sealed.YourPosition)
	}
}

// Structural half: SealedPeer is the only description of another participant
// that a sealed view can hold, so it must have nowhere to put their words.
// If someone later adds a Position field "just for the organiser", this fails.
func TestSealedPeerTypeHasNowhereToPutAPosition(t *testing.T) {
	rt := reflect.TypeOf(SealedPeer{})
	allowed := map[string]reflect.Kind{"Name": reflect.String, "Committed": reflect.Bool}
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		want, ok := allowed[f.Name]
		if !ok {
			t.Errorf("SealedPeer gained field %q (%s) — a pre-reveal type must not describe another participant beyond name and committed-ness", f.Name, f.Type)
			continue
		}
		if f.Type.Kind() != want {
			t.Errorf("SealedPeer.%s changed kind to %s", f.Name, f.Type.Kind())
		}
	}
}

// The organiser is the most tempting person to give an early read to.
func TestOrganiserCannotSeePositionsBeforeOrAfterReveal(t *testing.T) {
	s, ta, tb, _ := threeWay(t, RevealWhenAllCommitted, 0, time.Time{})
	_ = s.Commit(ta, secretA, t0)
	_ = s.Commit(tb, secretB, t0)

	blob, _ := json.Marshal(s.OrganiserView("http://x", t0))
	for _, secret := range []string{secretA, secretB} {
		if strings.Contains(string(blob), secret) {
			t.Errorf("organiser view leaked a position: %s", blob)
		}
	}
	// The organiser still gets what they need to facilitate.
	ov := s.OrganiserView("http://x", t0)
	if ov.Committed != 2 || ov.Total != 3 {
		t.Errorf("organiser should see 2/3 committed, got %d/%d", ov.Committed, ov.Total)
	}
	if ov.Rows[0].Link == "" {
		t.Error("organiser should get the participant links to distribute")
	}
}

// ---------------------------------------------------------------------------
// Reveal
// ---------------------------------------------------------------------------

// Atomic: there is no window in which one participant can read the positions
// and another cannot. If reveal were per-participant, the last to poll would
// read everyone else's answer before writing their own.
func TestRevealIsAtomic(t *testing.T) {
	s, ta, tb, tc := threeWay(t, RevealWhenAllCommitted, 0, time.Time{})
	_ = s.Commit(ta, secretA, t0)
	_ = s.Commit(tb, secretB, t0.Add(time.Minute))

	for _, tok := range []string{ta, tb, tc} {
		if _, rv, _ := s.ViewFor(tok, t0.Add(90*time.Second)); rv != nil {
			t.Fatal("revealed before the last participant committed")
		}
	}

	revealMoment := t0.Add(2 * time.Minute)
	_ = s.Commit(tc, secretC, revealMoment)

	// READ AT THREE DIFFERENT TIMES. This is the load-bearing part of the
	// test and the first version got it wrong: reading all three at the same
	// `now` cannot distinguish a session-wide reveal from a timestamp stamped
	// per read, so the test passed against a mutation that made reveal
	// per-participant. Caught by mutation-testing the guard, not by review.
	readAt := []time.Time{t0.Add(3 * time.Minute), t0.Add(4 * time.Minute), t0.Add(9 * time.Minute)}
	var stamps []time.Time
	for i, tok := range []string{ta, tb, tc} {
		_, rv, err := s.ViewFor(tok, readAt[i])
		if err != nil {
			t.Fatal(err)
		}
		if rv == nil {
			t.Fatal("a participant is still sealed after everyone committed")
		}
		if len(rv.Positions) != 3 {
			t.Fatalf("expected 3 positions, got %d", len(rv.Positions))
		}
		stamps = append(stamps, rv.RevealedAt)
	}
	for i, got := range stamps {
		// The reveal happened when the LAST person committed — not when each
		// participant happened to open the page.
		if !got.Equal(revealMoment) {
			t.Errorf("reveal is not atomic: reader %d saw RevealedAt=%v, want the commit moment %v (a per-read stamp would return %v)",
				i, got, revealMoment, readAt[i])
		}
	}
}

// The rule that makes a deadline safe: silence does not buy you a free read.
func TestNonResponderDoesNotReceiveTheReveal(t *testing.T) {
	deadline := t0.Add(time.Hour)
	s, ta, tb, tc := threeWay(t, RevealAtDeadline, 0, deadline)
	_ = s.Commit(ta, secretA, t0)
	_ = s.Commit(tb, secretB, t0)
	// Carol says nothing.

	after := deadline.Add(time.Minute)

	_, rv, _ := s.ViewFor(ta, after)
	if rv == nil {
		t.Fatal("Alice committed and should see the reveal after the deadline")
	}
	if len(rv.Positions) != 2 || len(rv.DidNotCommit) != 1 || rv.DidNotCommit[0] != "Carol" {
		t.Errorf("expected 2 positions and Carol listed as silent, got %+v", rv)
	}

	sealed, rvC, _ := s.ViewFor(tc, after)
	if rvC != nil {
		t.Fatal("SEAL BROKEN: Carol never committed and can read everyone's positions")
	}
	blob, _ := json.Marshal(sealed)
	for _, secret := range []string{secretA, secretB} {
		if strings.Contains(string(blob), secret) {
			t.Errorf("Carol's view leaked a position: %s", blob)
		}
	}
}

func TestQuorumReveals(t *testing.T) {
	s, ta, tb, _ := threeWay(t, RevealAtQuorum, 2, time.Time{})
	_ = s.Commit(ta, secretA, t0)
	if _, rv, _ := s.ViewFor(ta, t0); rv != nil {
		t.Fatal("revealed at 1 of quorum 2")
	}
	_ = s.Commit(tb, secretB, t0.Add(time.Minute))
	if _, rv, _ := s.ViewFor(ta, t0.Add(time.Minute)); rv == nil {
		t.Fatal("should have revealed at quorum 2")
	}
}

// Edge case worth pinning: the deadline fires and NOBODY answered. The
// session is technically open and there is nothing to show — and because
// non-responders do not receive the reveal, every participant correctly
// still gets a sealed view. An empty room reveals nothing to anybody.
func TestDeadlineWithNoCommitsRevealsNothingToAnyone(t *testing.T) {
	deadline := t0.Add(time.Hour)
	s, ta, _, _ := threeWay(t, RevealAtDeadline, 0, deadline)
	s.Tick(deadline.Add(time.Minute))
	if _, rv, _ := s.ViewFor(ta, deadline.Add(time.Minute)); rv != nil {
		t.Error("a participant who never committed received a reveal of an empty room")
	}
}

func TestForceRevealIsTheOrganisersEscapeHatch(t *testing.T) {
	s, ta, _, _ := threeWay(t, RevealWhenAllCommitted, 0, time.Time{})
	_ = s.Commit(ta, secretA, t0)
	s.ForceReveal(t0.Add(time.Hour))
	_, rv, _ := s.ViewFor(ta, t0.Add(time.Hour))
	if rv == nil {
		t.Fatal("force reveal did not open the session")
	}
	if len(rv.DidNotCommit) != 2 {
		t.Errorf("expected two silent participants, got %v", rv.DidNotCommit)
	}
}

// ---------------------------------------------------------------------------
// Commit rules
// ---------------------------------------------------------------------------

func TestCommitIsFinal(t *testing.T) {
	s, ta, _, _ := threeWay(t, RevealWhenAllCommitted, 0, time.Time{})
	if err := s.Commit(ta, secretA, t0); err != nil {
		t.Fatal(err)
	}
	if err := s.Commit(ta, "changed my mind", t0.Add(time.Minute)); err != ErrAlreadyCommit {
		t.Errorf("expected ErrAlreadyCommit, got %v", err)
	}
	sealed, _, _ := s.ViewFor(ta, t0.Add(time.Minute))
	if sealed.YourPosition != secretA {
		t.Error("a final position was overwritten")
	}
}

func TestCannotCommitAfterReveal(t *testing.T) {
	s, ta, tb, tc := threeWay(t, RevealAtQuorum, 2, time.Time{})
	_ = s.Commit(ta, secretA, t0)
	_ = s.Commit(tb, secretB, t0)
	if err := s.Commit(tc, secretC, t0.Add(time.Minute)); err != ErrAlreadyOpen {
		t.Errorf("expected ErrAlreadyOpen once positions are public, got %v", err)
	}
}

func TestUnknownTokenIsRejected(t *testing.T) {
	s, _, _, _ := threeWay(t, RevealWhenAllCommitted, 0, time.Time{})
	if _, _, err := s.ViewFor("not-a-real-token", t0); err != ErrUnknownToken {
		t.Errorf("expected ErrUnknownToken, got %v", err)
	}
	if err := s.Commit("not-a-real-token", "hello", t0); err != ErrUnknownToken {
		t.Errorf("expected ErrUnknownToken, got %v", err)
	}
}

func TestEmptyPositionRejected(t *testing.T) {
	s, ta, _, _ := threeWay(t, RevealWhenAllCommitted, 0, time.Time{})
	if err := s.Commit(ta, "   ", t0); err != ErrEmptyPosition {
		t.Errorf("expected ErrEmptyPosition, got %v", err)
	}
}

func TestSessionNeedsTwoParticipants(t *testing.T) {
	if _, err := NewSession("Fran", "p", []string{"Solo"}, RevealWhenAllCommitted, 0, time.Time{}, t0); err == nil {
		t.Error("a paired provocation with one participant should be rejected")
	}
}

// Tokens must not be guessable — this prototype's ONLY access control.
// See README: that is a known weakness, not a design.
func TestTokensAreDistinctAndLong(t *testing.T) {
	s, _, _, _ := threeWay(t, RevealWhenAllCommitted, 0, time.Time{})
	seen := map[string]bool{}
	for _, p := range s.participants {
		if len(p.Token) < 20 {
			t.Errorf("token too short to resist guessing: %q", p.Token)
		}
		if seen[p.Token] {
			t.Errorf("duplicate token %q", p.Token)
		}
		seen[p.Token] = true
	}
}

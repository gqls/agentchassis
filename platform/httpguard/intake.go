package httpguard

import (
	"strconv"
	"strings"
	"time"
)

// DefaultMinFill is the floor on how fast a human fills and submits a form.
// Ported from idea.uk, where it has run in production against real traffic.
const DefaultMinFill = 2500 * time.Millisecond

// IntakeVerdict says whether a form submission looks automated, and why.
type IntakeVerdict struct {
	Bot    bool
	Reason string // "honeypot" | "too-fast" | "" when human
}

// CheckIntake applies the two cheap bot gates that cost nothing and catch most
// unsophisticated form spam.
//
//   - HONEYPOT: a field hidden from humans by CSS. Any value at all means a bot
//     filled the form blind.
//   - TIMING: the page's JS posts the on-screen duration as a client-side delta
//     (immune to clock skew between client and server). Faster than minFill is a
//     bot.
//
// FAIL OPEN on a missing or unparseable elapsed value, so a visitor with
// JavaScript disabled is never blocked. That asymmetry is deliberate: the cost of
// turning away a real customer is much higher than the cost of letting one bot
// through to the limiter behind this.
//
// THE CALLER MUST RESPOND IDENTICALLY to a bot verdict and a success. idea.uk
// returns the byte-identical success page, because a distinguishable rejection
// tells the author exactly which gate to tune. Returning 403 here would undo the
// gate.
func CheckIntake(honeypotValue, elapsedMillis string, minFill time.Duration) IntakeVerdict {
	if strings.TrimSpace(honeypotValue) != "" {
		return IntakeVerdict{Bot: true, Reason: "honeypot"}
	}
	raw := strings.TrimSpace(elapsedMillis)
	if raw == "" {
		return IntakeVerdict{} // no JS: fail open
	}
	ms, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return IntakeVerdict{} // unparseable: fail open
	}
	if ms < 0 {
		return IntakeVerdict{} // nonsense: fail open rather than guess
	}
	if time.Duration(ms)*time.Millisecond < minFill {
		return IntakeVerdict{Bot: true, Reason: "too-fast"}
	}
	return IntakeVerdict{}
}

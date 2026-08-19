// FILE: cmd/regcheck/main.go
//
// Operator CLI for the regulated-identity guard (register CGV-033).
//
// It answers the two questions an operator actually has, using the SAME code the
// deploy gate runs — not a re-implementation that can drift:
//
//  1. Is this site attested, and if not, WHICH required field is missing?
//  2. Would this sentence be refused on this site?
//
// WHY IT EXISTS. The guard is Go, so it is inert until a chassis roll — but the
// attestation is DATA, written by hand into site_specs the moment a client proves
// they are authorised. Without this, an operator recording an attestation has no
// way to check they got the shape right until the next roll, and the failure mode
// is silent: an attestation with a mistyped FRN reads as "no attestation", which
// looks exactly like a site nobody has attested yet. This makes that visible now.
//
// Usage:
//
//	regcheck -evidence <evidence_base.json>
//	regcheck -evidence <evidence_base.json> -claim "We are authorised and regulated by the FCA."
//	kubectl ... psql -At -c "SELECT data FROM site_specs WHERE ..." | regcheck -evidence -
//
// Exit codes: 0 = attested (or, with -claim, the claim is allowed); 1 = not
// attested / the claim is refused; 2 = bad input. Scriptable as a gate.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func main() {
	evidencePath := flag.String("evidence", "", "evidence_base JSON file, or '-' for stdin")
	claim := flag.String("claim", "", "optional sentence to test against the guard")
	flag.Parse()

	if *evidencePath == "" {
		fmt.Fprintln(os.Stderr, "regcheck: -evidence is required (a file, or '-' for stdin)")
		os.Exit(2)
	}

	var raw []byte
	var err error
	if *evidencePath == "-" {
		raw, err = io.ReadAll(os.Stdin)
	} else {
		raw, err = os.ReadFile(*evidencePath)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "regcheck: cannot read evidence: %v\n", err)
		os.Exit(2)
	}
	if len(strings.TrimSpace(string(raw))) == 0 {
		fmt.Fprintln(os.Stderr, "regcheck: evidence is empty — an empty evidence_base means NOT attested")
		os.Exit(1)
	}

	eb, err := datahelpers.ParseEvidenceBase(raw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "regcheck: evidence_base does not parse: %v\n", err)
		os.Exit(2)
	}

	attested := eb.RegulatedAttested()
	fmt.Printf("attested: %v\n", attested)

	// Name the missing field rather than just saying no. "Not attested" with no
	// reason is the answer that sends an operator round the loop again.
	if !attested {
		if eb == nil || eb.Regulated == nil {
			fmt.Println("reason:   no `regulated` block in the evidence_base")
		} else {
			for _, m := range missingFields(eb.Regulated) {
				fmt.Printf("missing:  %s\n", m)
			}
		}
	} else {
		fmt.Printf("firm:     %s (FRN %s)\n", eb.Regulated.FirmName, eb.Regulated.FRN)
		fmt.Printf("attested: by %s on %s\n", eb.Regulated.AttestedBy, eb.Regulated.AttestedAt)
		fmt.Printf("evidence: %s\n", eb.Regulated.Evidence)
	}

	if *claim == "" {
		if attested {
			os.Exit(0)
		}
		os.Exit(1)
	}

	findings := datahelpers.ScanAllBannedClaims([]string{*claim}, eb)
	if len(findings) == 0 {
		fmt.Printf("\nclaim ALLOWED: %q\n", *claim)
		os.Exit(0)
	}
	fmt.Printf("\nclaim REFUSED: %q\n", *claim)
	for _, f := range findings {
		fmt.Printf("  pattern: %s\n  reason:  %s\n", f.Pattern, f.Reason)
	}
	os.Exit(1)
}

// missingFields mirrors RegulatedAttestation.Attested's requirements. It is a
// REPORTING helper only — Attested remains the single decision point, so the two
// cannot disagree about whether the gate opens, only about how the failure is
// described.
func missingFields(r *datahelpers.RegulatedAttestation) []string {
	var out []string
	if strings.TrimSpace(r.FirmName) == "" {
		out = append(out, "firm_name — the authorised entity's registered name")
	}
	if strings.TrimSpace(r.FRN) == "" {
		out = append(out, "frn — the Financial Services Register firm reference number")
	} else if !isFRNShape(strings.TrimSpace(r.FRN)) {
		out = append(out, fmt.Sprintf("frn — %q is not 6 or 7 digits", r.FRN))
	}
	if strings.TrimSpace(r.AttestedBy) == "" {
		out = append(out, "attested_by — who checked it")
	}
	if strings.TrimSpace(r.Evidence) == "" {
		out = append(out, "evidence — what proof was seen (e.g. 'email 2026-08-19 + FS Register entry')")
	}
	if strings.TrimSpace(r.AttestedAt) == "" {
		out = append(out, "attested_at — when it was checked (YYYY-MM-DD or RFC3339)")
	} else if !parsesAsDate(r.AttestedAt) {
		out = append(out, fmt.Sprintf("attested_at — %q is not a readable date", r.AttestedAt))
	}
	if len(out) == 0 {
		out = append(out, "(no field is missing — if this prints, Attested and missingFields disagree; that is a bug)")
	}
	return out
}

func isFRNShape(s string) bool {
	if len(s) < 6 || len(s) > 7 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// parsesAsDate reuses the package's own date parser via a round trip: an
// attestation identical to a valid one EXCEPT for the date tells us whether the
// date is what is wrong, without exporting the parser or copying its layouts.
func parsesAsDate(s string) bool {
	att := &datahelpers.RegulatedAttestation{
		FirmName: "x", FRN: "123456", AttestedBy: "x", Evidence: "x", AttestedAt: s,
	}
	return att.Attested()
}

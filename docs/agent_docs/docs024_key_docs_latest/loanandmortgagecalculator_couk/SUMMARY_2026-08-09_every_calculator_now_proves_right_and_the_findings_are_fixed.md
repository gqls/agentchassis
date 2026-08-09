# SUMMARY — 2026-08-09 — every calculator now proves right, and the findings are fixed

**What we're trying to do.** loanandmortgagecalculator.co.uk carries 23
calculators about consumer credit and tax. The standing goal is that every
number they show is provably correct — checked against independent maths and
published rules, not merely unchanged since yesterday — and that the site is
managed by the framework so corrections are systematic, not hand-patches.

**Where we've come from.** The 08-08 milestone built the independent oracle
the owner asked for and it found what the goldens never could: eight of the
23 tools were wrong. Six loan calculators failed on a 0% interest rate — each
carried its own private copy of the repayment maths and every private copy
was missing the zero case, so users saw £NaN, a wrong recommendation, or a
stale answer dressed as a fresh one. And the stamp-duty tool was applying a
First Time Buyer relief cap that expired in March 2025, under-quoting a real
tax bill by £5,000, while also charging a £2,000 surcharge on purchases the
rules exempt. The 08-08 summary ended with the findings filed
(`bugs_open/224`, `bugs_open/225`) and nothing yet changed.

**What we've done.** Both findings are now fixed, live, and verified — by two
sessions working the same night, without file overlap. The 0% fix
(`bugs_open/224`) deleted the six private formula copies: every loan tool now
calls the same shared engine the mortgage tools always used, with one new
balloon-payment helper, so this class of defect has one place to be right in.
Every tool also now always writes an answer (or a clear blank) — never a
leftover one. The stamp-duty fix (`bugs_open/225`, owner-approved figures)
replaced the hand-rolled branches with a dated-constants port of the oracle's
own HMRC rules, on this site and on the byte-identical twin page at
mortgagecalculator.co.uk. Along the way three of the lane's own tools were
hardened: the repo↔database repair script would have destroyed a decomposed
page if run as documented, the deploy script crashed polling its own
bookkeeping, and a comparator quirk that makes untouched pages read as
"diverged" is now a recorded landmine with its control.

**Where we are now.** The full oracle estate — 176 checks across all 23
tools, boundary vectors included — is green: PASS 170, FAIL 0, plus six
recognised rounding-convention notes. The checker itself was proven able to
fail (mutation controls, parse self-test) in the same sessions that read it
green. Both bug files stay in `bugs_open/` with full evidence, per the
owner's ruling. The site remains 39 pages verbatim, two decomposed; the
calculators are still the clobber-sensitive class for any future whole-site
rebuild.

**Where we're going.** The `--emit-criteria` step the validation plan gated
on "224 and 225 both fixed" is now unblocked — the tools' correct answers can
be pinned into the platform's acceptance record. The decomposition programme
can resume (stamp-duty needs its pinned ref moved past the fix first, as the
225 session recorded). And the two structural findings this work surfaced —
the diagnosis loop and landmine verifier cannot see non-Go artefacts
(`bugs_open/223`, and 224's no-verdict run) — belong to other lanes but
remain the estate-level gap this lane keeps hitting.

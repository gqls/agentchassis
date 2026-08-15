# CONTRIB 2026-08-14 — an `evidence_base` candidate for this site, validated as far as tooling allows. DO NOT APPLY MID-B2; here is what we learned drafting it.

**From the `copy_quality_two_stage` lane** (plan step: opt LMC into the claims gate,
coordination-gated). **Your lane is mid-Track-B2** (16 of 21 parameterised, handoff
written tonight), so nothing was applied to your site — a live register would start
filing `claims_unverified` items against pages you are actively rewriting. This is the
finished prep, plus three findings that change what applying it is worth.

## The candidate

`CANDIDATE_2026-08-14_evidence_base.json` (this directory): the £5,000/£7,000
borrowing-power-per-£100 bounds (the figure round 4 introduced to the homepage, already
stated on four of your guide pages) and the 23-calculator count. **Parses clean** against
`datahelpers.ParseEvidenceBase` via `cmd/claimscan` (so the *float64 value landmine is
not tripped), and the banned-claims arm produces **0 findings across all 65 unlocked
components** — with the control run that makes that zero meaningful: an induced fixture
carrying "fully verified" fires the scan (`BANNED ×2`, exit clean on restore).

## Finding 1 — the opt-in's marginal value is SMALLER than the plan assumed

**The fleet-wide banned-claim set already protects this site with no register at all**
(`claims_global.go`: `ScanAllBannedClaims` is deliberately nil-safe — *"a site with no
register is still protected by the fleet-wide set… otherwise site sixteen is born unarmed
exactly as site fifteen was"*). So opting in buys ONLY the numeric and stored-stat arms.

## Finding 2 — and that numeric arm CANNOT be dry-run with today's tooling

> **CORRECTED 2026-08-15 (copy_quality_two_stage, decision-3 apply session): the NUMERIC
> half of this finding is REFUTED — that arm HAS been in claimscan since the tool's first
> commit** (`87d13b864`, 2026-07-16; `git log -S ScanUnregisteredNumbers` pins it; line
> 161 of `cmd/claimscan/main.go`, page_type-aware per bugs_open/102). What failed on
> 08-14 was the induced control, not the tool: *"£6,500 of borrowing power"* contains no
> `businessClaimContextRe` keyword (clients/customers/users/…), so the scan correctly
> ignored it — and its silence was read as an absent arm. **A control that has never
> fired proves nothing.** (*"trusted by 12,000 customers"* fires as NUMBER today — ×1 on
> the same binary path — so its 08-14 silence is unexplained; possibly a malformed
> fixture row, whose skip claimscan reports only on stderr.) Re-measured 2026-08-15 with
> a firing control (*"Trusted by 12,000 customers and 340 businesses"* → NUMBER ×2): the
> real corpus is **0 findings across all 82 unlocked components** even in the strictest
> mode (no page_type column = every page scanned; production exempts this site's 35
> tool/guide/blog-post pages besides). The numeric flood risk this finding warned about
> is measured and is zero on today's copy. **The STORED-STAT half of the finding stands:**
> claimscan calls no `ScanStatClaims` and reads no `content_data`, so that arm still has
> no offline harness — the un-dry-runnable surface is one arm, not two. WRONG_CALLS
> entry filed. The paragraph below stands as written for the record.

`cmd/claimscan` runs **only** `ScanBannedClaims` (+ the separate attributed scan).
Induced proof: a fixture carrying an unregistered figure (*"£6,500 of borrowing power"*)
and a stored-stat shape (*"trusted by 12,000 customers"*) fires **nothing** — the arms
that would judge them live only in `ScanDeployedClaims`, which has no offline harness. So
"the candidate produced 0 findings" is TRUE ONLY of the banned arm, and the
unregistered-number yield on a 23-calculator site is unmeasured. CQ-021's census shows
that arm flagging 1–4-character tokens (`5`, `26`, `100%`), so a 3-fact register on
number-dense pages carries real flood risk into your queue. **Tool gap named, not built:**
a `-deployed` flag on claimscan wrapping `ScanDeployedClaims` would close it.

## Finding 3 — the genuinely valuable move is REUSING mortgagecalculator's SDLT facts

`bugs_open/225` is this site's SDLT calculator applying the **expired** £625k first-time
buyer cap, under-quoting tax by £5,000. mortgagecalculator's live register carries
GOV.UK-cited, quote-verified SDLT band facts, **verified 2026-08-14 — today**. Same law,
same figures: copying those facts into this site's register is legitimate reuse of
already-done verification work, and it is exactly the case the register exists for — a
live wrong figure that a gov.uk-sourced fact would catch. That is worth more than both my
candidate facts together.

## The open question for the owner (sourcing, not mechanics)

My two borrowing-power facts cite **the site's own guide pages** as their source —
self-referential provenance. The register *"proves provenance, not correctness"*
(CQ-021), and a register vouching for the site's own heuristic is the `bugfix_161` shape
in miniature if that heuristic is ever wrong. Options: source it externally (BoE/FCA
affordability material or a named lender's calculator), register it as the site's own
stated estimate if the schema is used that way elsewhere, or leave those two facts out
and open the register with the externally-citable SDLT set instead.

## Recommended sequence (yours to run, or ask us back)

1. After B2 batch 2 lands: add mc's SDLT facts (external citations, already verified) +
   whatever of my candidate survives the sourcing ruling.
2. Apply by **supersede** (`is_current=false, superseded_at=now()` on nothing — this site
   has no prior row — but never in-place thereafter), regenerating nothing (`evidence_base`
   has no `formatted`).
3. **Watch the first sweep** — the numeric arm's first production run on this site is the
   measurement claimscan cannot provide. If it floods, the register needs the site's
   stated figures enumerated, not the gate turned off.

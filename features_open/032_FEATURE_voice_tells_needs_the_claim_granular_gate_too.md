# 032 — `voice_tells` needs the claim-granular gate too (or a shared helper)

**Filed:** 2026-08-12 · **Lane:** `bugfix_168_deployed_asset_path` (retraction sweep)
**Raised by:** council `bug_historian` (MEDIUM) and `reuse_agent` (MEDIUM), independently,
round 6 of correlation `b67eb26a-14ef-45d7-b755-3e489fd57ef0`.

## The gap

`claims_unverified` now has two gates before a parked item may close:

1. **copy-changed** — an EXAMINED component edited after the finding was filed (owner ruling
   2026-08-09);
2. **claim-granular** — every text the finding cited has gone from the slot it was cited from
   (council round 6).

**`voice_tells` (CQ-020) has neither.** It has the identical moving-standard hole: a site's voice
gate is *data*, so a page can stop tripping the check because somebody edited the gate rather than
the copy, and the item retracts with the copy untouched.

## Why it was left, and why that is not a finished answer

The asymmetry is deliberate and was argued, not overlooked: `voice_tells` is live and
council-approved, and **its surface is style rather than truth** — the distinction the council seats
themselves drew when they escalated `claims_unverified` to the owner. A wrong `resolved` on a voice
finding costs a stylistic regression; on a claims finding it withdraws a live statement that a
factual assertion is unsupported.

`bug_historian` accepted the reasoning and objected to the *shape*: this is 016b §9's

> "One call site of a shared judgement gets the rigorous fix; the sibling stays heuristic."

and its point is that **the un-hardened sibling is the one that bites later**. An accepted risk with
no tracking artefact is indistinguishable, six months on, from an oversight. Hence this file.

## What to build

`reuse_agent` named the better version. Both gates are currently **claims-specific helpers inlined
in `revalidate_unverified_claims.go`**:

- `unverifiedClaimsVerdict`'s `filedAt` / `NewestComponentUpdate` comparison, and
- `claimStillOnPage` + `flaggedClaimsFromSpec`.

If `voice_tells` (or a seventh type) needs the same protection later, the default path is a second
bespoke implementation — **two independent answers to "has the underlying content actually changed
since this finding was filed?", free to drift**. That is the dual-path problem this platform's
founding incident was about.

So the work is, in order:

1. **Lift the copy-changed comparison into a shared helper** — `revalidate_review_queue_action.go`
   beside `parkedReviewItem`, or `datahelpers`. It needs only `filedAt` and a newest-update
   timestamp, so it is type-agnostic already.
2. **Adopt it in `revalidateVoiceTells`.** `VoicePageScan` would need `NewestComponentUpdate`, which
   is the same three lines `ClaimsPageScan` carries.
3. **Then decide on claim-granularity for voice separately** — it is *not* obviously right there.
   A voice finding cites a phrase (`spec.findings[].matched` exists on that type too), but voice
   findings are about register and tone, and a rewrite that fixes the tone may legitimately keep the
   phrase. **Measure before building**, exactly as round 6 did for claims: run the demand control
   over `still_holds` voice items and see whether the cited text is still present when the finding
   genuinely holds.

## Do not skip the measurement

Round 6's lesson, and it applies directly here: the council seat that asked for the claim-granular
gate proposed comparing the cited **snippet**. Measured against the population where the answer is
known, the snippet saw a present claim in **18 of 41** cases against the token's **40 of 41** — and
in a gate a missed match reads as *"the copy changed"*, which **grants** closure. The seat was right
about the defect and wrong about the remedy. **Assume the same could be true here.**

## Related

- `docs/agent_docs/docs026_concept_register/register/content-quality.md` — **CQ-020** (voice),
  **CQ-021** (claims, both gates, with the measurements)
- `docs024_key_docs_latest/bugfix_168_deployed_asset_path/NOTES_deployed_asset_path.md` — the
  demand-control method
- `LANDMINES.md` — "Comparing a work item's flagged text against live copy PAGE-WIDE always finds
  it", which is the trap any implementation here will hit

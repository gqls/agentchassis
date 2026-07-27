# 104 — `banned_claims` is per-site only, so 10 of 15 live sites are unarmed

**Filed** 2026-07-27 from the oufe.com workstream.
**Severity** medium-high — not a code defect. The engine works; it is pointed at
almost nothing. The two most exposed sites in the estate are among the unarmed.
**Class** coverage / platform configuration.
**Status** OPEN. Not a new idea — see "This decision is already filed" below.

## Measurement

```sql
SELECT s.domain, jsonb_array_length(COALESCE(ss.data->'banned_claims','[]'::jsonb)) AS bans
  FROM sites s
  LEFT JOIN site_specs ss ON ss.site_id=s.id AND ss.aspect='evidence_base' AND ss.is_current
 WHERE s.status NOT IN ('pool','archived') ORDER BY bans DESC NULLS LAST;
```

2026-07-27: **5 of 15** live sites carry a single banned-claim pattern —
leopardessconsulting (19), oufe (28 after mig 226), vonc (9), relojistas (9),
fundamentallyai (6). The other ten carry **zero**, including:

- **vetcomparison.uk** — the site that published fabricated prices for 3,124 named
  real vet practices and required a contemporaneous legal record;
- **idea.uk** — the only site taking real money (£29 reports);
- robot-hands.com, ai-agent-orchestration.com, gamesdesign.co.uk — which have
  `facts[]` and a `writer_block` but no patterns.

## Why this is the defect and not something else

`ScanBannedClaims` (`platform/orchestration/datahelpers/claims.go:284-325`) is a
bare case-insensitive regex over prose blocks. It has no numeric gating —
`businessClaimContextRe` and `isExcludedNumber` gate only `ScanUnregisteredNumbers`
(`claims.go:365,369`). It will catch **whatever patterns a site is given**, about
anyone, numeric or qualitative.

So the engine is general and the wiring is sound. Every reader is keyed on
`site_id` (`validate_page_content.go:967`, `check_unverified_claims.go:229`,
`refresh_evidence_base_action.go:213`), there is no `site_id IS NULL` global row,
no network-level inheritance, and no merge of a base set into per-site sets. **A
pattern learned on one site cannot reach any other, and every new site is born
unarmed.**

Concretely, the lesson this estate paid for twice — fabricated social proof,
invented figures, promises of our own infallibility — is written down five times
and enforced five times, on the five sites where somebody happened to write it.

## This decision is already filed, and its precondition has lapsed

`docs/agent_docs/docs024_key_docs_latest/claims_verification/SPEC_claims_verification.md:250-252`
poses exactly this question and names exactly this use case:

> "Should `banned_claims` be fleet-shareable (some patterns are universal:
> 'Awards Won', invented-client shapes) or strictly per-site? **Proposal:
> per-site only until two sites have evidence bases.**"

Restated in `PLAN_2026-07-16_claims_verification.md:138-139` under what is
deliberately not built. The constraint behind it is landmine 7
(`SPEC:220-223`): *"Do NOT block builds fleet-wide on a layer only one site has
data for"* — entirely correct when n=1.

**There are now 8 sites with evidence bases.** The precondition lapsed months ago
and nothing re-opened the question, because a deferral with a numeric trigger and
no watcher is indistinguishable from a decision.

## The precedent to copy is one directory away

`platform/orchestration/datahelpers/voicetells.go` already solves this exact
shape for the sibling engine:

- `globalTellPhrases()` (`:121-137`) — a hardcoded fleet-wide list;
- unioned with the per-site list at `:109`:
  `append(globalTellPhrases(), spec.VoiceGate.BannedPhrases...)`;
- still gated by the per-site opt-in, so it only fires where the gate is enabled.

Two further precedents for fleet-wide patterns with **no** opt-in at all:
`checkMetaCommentary` (`validate_page_content.go:273,1003-1023`, severity blocker)
and `check_placeholder_contact.go:95-124`.

## Fix candidates, ordered by what closes the door

1. **Mirror `globalTellPhrases` into `ParseEvidenceBase`** (`claims.go:121-128`):
   a `globalBannedClaims()` list unioned with the per-site set. Universal patterns
   become default-on for every armed site, and a lesson learned anywhere protects
   everywhere. Small, precedented, one image roll.
   *Residual:* still gated by `eb != nil`, so the seven sites with no register at
   all remain uncovered.
2. **1, plus make the universal set apply even without a register** — i.e. treat
   the global list like `checkMetaCommentary`, which needs no opt-in. Closes the
   door completely, and is the only candidate that protects a brand-new site on
   its first build. Bigger blast radius: it changes what every site's build gate
   can block, so it wants a council round and a dry-run count of what it would
   have fired on across live pages before it is switched on.
3. **Write the patterns into each site by hand** (config, no code). Zero risk,
   immediate, and it is what mig 226 did for oufe — but it is ten more UPDATEs,
   it does not survive site number sixteen, and it is exactly the state that
   produced this bug.

Recommend 1 now and 2 behind a measured dry-run, because the whole point is that
the next site should not have to be armed by someone remembering.

## Suggested starting content for the universal set

Tested against real copy in `sql_for_agents/226_overclaim_patterns_oufe.sql` —
10/10 fabrication shapes blocked, 13/13 legitimate sentences passed. The
overclaimed-reliability family is a good first candidate for universality because
**no site should ever assert it**: completeness-of-exclusion, verification-of-
everything, invitations to rely, self-accuracy claims, infallibility claims.

The line those patterns draw generalises: **a site may describe what it does; it
may not claim what that guarantees.**

## How to verify a fix

Pick a site with **no** register today. Add a page whose copy asserts "every claim
on this site is verified", and build it. Under candidate 1 with a register seeded,
and under candidate 2 with none, the build must fail with a `claims` blocker.
Then confirm a legitimate process sentence ("we cite each figure and date it")
still builds — **induce both, because a checker that fires on everything is as
useless as one that fires on nothing.**

## Related

- `sql_for_agents/226` (arms oufe), `227` (corrects the false premise that no
  scanner could catch this class).
- `bugs_open/083_HANDOFF_2026-07-26_detected_findings_never_reach_a_handler.md` —
  the other half of the reach problem: even where a site *is* armed, the
  post-deploy sweep that would catch drift effectively never runs.
- `docs024_key_docs_latest/WRONG_CALLS.md` 2026-07-26 — how a wrong belief about
  this engine's reach nearly produced a redundant subsystem.

---

## Triage 2026-07-27, later the same day — the title figure has already moved: 8 unarmed, not 10

Verification sweep, not a fix. **Re-run the § Measurement query before quoting this file** —
it drifted within hours of being written, which is itself the point.

```
 oufe.com 28 · leopardessconsulting.co.uk 19 · ai-agent-orchestration.com 10 ·
 vonc.com 9 · relojistas.com 9 · fundamentallyai.com 6 · finetuning.uk 3
 gaswholesalers.com 0 · robot-hands.com 0 · gamesdesign.co.uk 0 · idea.uk 0 ·
 vetcomparison.uk 0 · system.internal 0 · dartsonline.com 0 · webdesign.co.uk 0
```

**7 of 15 armed** (was 5): ai-agent-orchestration.com gained 10 and finetuning.uk gained 3,
by hand, since this was filed. `site_specs` now holds **9** current `evidence_base` rows.

Nothing about the defect changed — and the drift is the argument for candidate 1, not
against it. Two sites were armed by somebody remembering, one at a time, which is precisely
the state § "Fix candidates" says produces the bug. **The two named worst cases are still at
zero: vetcomparison.uk and idea.uk** — the site that published fabricated prices for 3,124
named practices, and the only site taking real money.

### The filed decision, located and quoted verbatim (this file's § "already filed" is correct)

`claims_verification/SPEC_claims_verification.md` § 10 "Open questions for the owner", q2:

> "Should `banned_claims` be fleet-shareable (some patterns are universal: "Awards Won",
> invented-client shapes) or strictly per-site? Proposal: per-site only until two sites have
> evidence bases."

Restated in `PLAN_2026-07-16_claims_verification.md` under what is deliberately not built:

> "**No fleet-wide banned list.** Per-site until at least two sites have evidence bases
> (spec open question 2, unchanged)."

**The precondition is a number and the number is 9.** So this is not a technical decision
waiting on evidence — it is an **owner call on a deferral whose own trigger has fired**, and
nothing watches it. Nobody needs to re-derive the design; candidate 1 (mirror
`globalTellPhrases`) is precedented one directory away.

### One thing that changes how much candidate 1 buys today

The post-deploy half of this layer does not run at all — `check_unverified_claims` is
reachable only through `improvement-sweep`, disabled since 2026-05-02
(`bugs_open/083`; `claims_unverified` has **0** live rows and **1** ever, from 2026-07-17).
So arming a site fleet-wide arms its **build gate** and nothing else. That is still the
half that matters for new copy, and it should be said plainly when the value is estimated.

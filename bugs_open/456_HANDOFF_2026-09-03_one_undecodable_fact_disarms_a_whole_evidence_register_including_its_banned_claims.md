# 456 — one undecodable fact disarms a whole evidence register, banned claims included

**Filed 2026-09-03** while resuming `bugs_closed/161`'s residual. **Two live sites are
affected today and both were found by audit, not by any signal the platform emits.**

> ## STATUS 2026-09-03 — root cause FIXED AT SOURCE and committed; the reporting half is written and HELD
>
> | half | state |
> |---|---|
> | **the parse** (a bad fact costs that fact, not the register) | **FIXED, committed `3f221f99f`**, council corr `c2d1d570-f6f4-4ce0-9820-e5d79501200e`. Inert until the next chassis roll. |
> | **the report** (raise `malformed_evidence_fact`; count the facts nothing re-proves) | **written, NOT committed** — see §7. Another session has an uncommitted `platform/livespec` helper wired into `refresh_evidence_base_action.go`; committing that file now would take their unfinished call site as a passenger and break HEAD for the fleet. |
> | **the two malformed rows** | **NOT repaired — they belong to their lanes**, and after the roll they no longer cost anything but themselves. See §6. |

## The one-sentence version

`ParseEvidenceBase` decoded a register's `facts` as one array, so **a single fact whose
`value` is text returned an error for the WHOLE base** — and every caller treats that error
as "this site has no register", which switches off the site's `banned_claims` too, though
bans have never depended on facts.

This is `bugs_closed/161`'s mechanism inverted. 161 was a register that **vouched for a
false claim**. This is a register that **protects nothing at all**, silently, and reads
exactly like a site that never opted in.

## What is false, and the control that proves it

`[MEASURED 2026-09-03, all 27 live registers, by running the real `ParseEvidenceBase`
through `cmd/regcheck` — not by reading the code and inferring]`

| site | inert `banned_claims` | facts | deployed pages | broken since |
|---|---|---|---|---|
| `finetuning.uk` | **3** | 10 | 49, last deployed 2026-09-03 | **2026-08-24** |
| `noted.co.uk` | **7** | 1 | 12 | **2026-08-25** |

The other 25 parse. Reproduce:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At \
  -c "SELECT sp.data::text FROM site_specs sp JOIN sites s ON s.id=sp.site_id
      WHERE sp.aspect='evidence_base' AND sp.is_current AND s.domain='noted.co.uk';" > eb.json
go run ./cmd/regcheck -evidence eb.json -claim "Your notes are end-to-end encrypted and we are fully GDPR compliant."
```

**Before `3f221f99f`** — the guard is not merely quiet, it is absent:

```
regcheck: evidence_base does not parse: evidence_base unmarshal:
  json: cannot unmarshal string into Go struct field EvidenceFact.facts.value of type float64
```

**The demand control, which is what makes this evidence rather than a story.** The same
register with **only the one malformed fact repaired** — nothing else changed — refuses that
same sentence:

```
claim REFUSED: "Your notes are end-to-end encrypted and we are fully GDPR compliant."
  pattern: end[- ]?to[- ]?end encrypt|zero[- ]?knowledge|we (have )?no access to your
  reason:  Not built. E2E encryption and zero-knowledge are specific, verifiable
           architectures with real cost; claiming either without building it is the most
           harmful possible lie on a notes product, because a user would reasonably store
           secrets on the strength of it.
  pattern: gdpr[- ]compliant|fully compliant|iso ?27001|soc ?2
  reason:  Compliance postures are attestations with auditors and dates behind them.
```

`finetuning.uk`'s inert bans are the same shape — its own owner's ruling:

```
claim REFUSED: "We cut 80% of quote preparation time and saved 40 hours."
  reason: 2026-07-27 owner ruling: '~80% reduction in quote preparation time' had nothing
          on record behind it and was removed. It must not return.
```

**So a ban an owner personally ruled on has been unenforceable for ten days, and the site
that most needs its security absolutes policed had that list switched off entirely.**

## The cause is a MISSING CAPABILITY, not a typo — which is why it recurs

`EvidenceFact.Value` is a `*float64` (`datahelpers/claims.go`). Authors are legitimately
registering facts whose value is **text**, because that is what the fact is:

| site | fact | `value` |
|---|---|---|
| finetuning.uk | `ft-licence-mistral7b` | `"Apache 2.0"` |
| finetuning.uk | `ft-licence-phi35mini` | `"MIT"` |
| finetuning.uk | `ft-licence-llama33` | `"Llama 3.3 Community License"` |
| finetuning.uk | `ft-booking-hours` | `"customer picks a time, 9am-5pm UK time"` |
| noted.co.uk | (unnamed) | `"30 days"` |

Nobody did anything careless. There is no shape in the register for a text-valued fact, and
the failure mode for reaching for one was **total, silent disarmament of the site's guard
list**. `[MEASURED]` finetuning.uk's count of such facts went **0 → 3 → 7 → 8** across
2026-08-24 → 08-26 (`site_specs` history, `jsonb_typeof(f->'value')='string'`) as one author
kept doing the reasonable thing. It was still growing when this was filed.

`noted.co.uk`'s row has a **second, independent defect the old decode never even reached**:
its `source` is a bare string, not an object. Go stops at the first type error, so `value`
is what the message names.

## Why nothing caught it

All three claims-gate callers fail **open**, each for a locally defensible reason:

- `validate_page_content.go:1448` — `logger.Warn("Failed to parse evidence_base spec — claims checks skipped")`, returns nil.
- `validate_page_content_stats.go:143` — returns nil for the stat audit.
- `discovery_checks/check_unverified_claims.go:318` — *"A malformed evidence base is a real defect on an opted-in site — surface it as a check error (logged, not fatal to the run)."*

The third **names this exact defect and calls it real** — and its remedy is a log line. The
only trace anywhere is a `Warn` in a pod that is replaced daily, and **no work item, ever**.
A site with a broken register is therefore indistinguishable, to every dashboard and every
reader, from a site that never opted in.

⚠ **The daily sweep could not have reported it either**, and that is a third, separate
defect: `refreshOneSiteEvidence`'s `if res.FactsChecked == 0 { return res, nil }` sits ahead
of **every** work-item raise in the function — including the `invalid_banned_claim_pattern`
raise added 2026-09-02. A register that fails the typed parse tends also to have nothing the
sweep can check, so the one mechanism that visits every register daily returned early on
precisely the sites that needed reporting.

## The fix

**Committed (`3f221f99f`), inert until the next chassis roll.** Facts decode **one at a
time**; a fact that fails is skipped and recorded in a new `EvidenceBase.MalformedFacts`;
`banned_claims`, both attestations and the citation codes decode as before. It can only ever
make a register parse where it previously did not — the top-level decode is the same call on
the same bytes minus the facts array.

**Ordering hazard, checked BEFORE proposing, because `bugs_closed/161` recorded it.** 161's
own landmine says *repair the copy first, then arm* — a banned claim is BLOCKER severity and
would make pages unsaveable. Arming these 10 bans is safe: `[MEASURED 2026-09-03]` both
sites' deployed component text scores **zero violations** against their own ban lists.
**Limit, stated:** that check reads stored `rendered_html` with tags stripped, so `<title>`
and JSON-LD are outside it, per the LANDMINES entry on prose-only sweeps.

## What is still owed

1. **§7's held commit** — the sweep half: raise `malformed_evidence_fact` (per-fact key,
   severity high), count `FactsUnverifiable`, widen the early return. Written and tested;
   waiting on another session's `livespec` helper to land. **Whoever picks this up: check
   `git status platform/livespec/` first.**
2. **The two malformed rows are their lanes' to repair**, not this one's — `noted_rebuild`
   wrote theirs on 2026-08-25 (`apply_privacy_copy.py`), and finetuning.uk's belongs to
   whoever owns that register. Either express the fact numerically (`value: 30`, unit in the
   claim) or drop `value` and put the wording in `claim` with an `attested_by` source. After
   the roll this costs only the fact itself, so it is no longer urgent — but until it is
   done, those facts support no claim.
3. **A text-valued fact shape.** The honest end state, and a shared-vocabulary addition worth
   its own round. Deliberately NOT built here: with the parse fixed, a text fact is skipped
   and reported rather than catastrophic, so the pressure is off. Named as an open option in
   RFC_025's addendum.

## Relations

- `bugs_closed/161` — the parent. This is its mechanism inverted, found while re-verifying it.
- `bugs_open/288` / `register_guards_code_phase_b` — owns this function and RFC_025's stage 2.
  Its §6.6 names the 27 unverifiable facts as an adoption gap; nobody had noticed the sweep
  cannot **count** them.
- `RFC_025` — stage 1's nudge is what §7's held half extends to every unverifiable fact.
- `bugs_open/033` — the `needs_human_review` queue (**1,389** open rows as of 2026-09-03).
  The new item goes there deliberately: the architecture seat asked, on 288's round, that no
  further bespoke `doc_notes` bypasses be invented.
- `CLM-003` / `CLM-007` / `CLM-014` — concept register, `claims-verification.md`.

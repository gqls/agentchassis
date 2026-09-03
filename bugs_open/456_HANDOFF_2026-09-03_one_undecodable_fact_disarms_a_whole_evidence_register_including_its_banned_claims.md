# 456 — one undecodable fact disarms a whole evidence register, banned claims included

> ⚠ **THE NUMBER 456 IS AMBIGUOUS — RESOLVE THIS FILE BY SLUG.** A second, unrelated bug was
> filed as 456 the same day by the `finetuning_uk_service` lane
> (`456_HANDOFF_2026-09-03_writer_emitted_a_malformed_closing_tag_and_it_reached_the_served_page_unchecked.md`).
> Numbers are never reassigned, so both stand. **A bare "456" in a commit message is more
> likely to mean theirs than this one** — their own commit already says "456 §4" meaning theirs.
> This file is `one_undecodable_fact_disarms_a_whole_evidence_register`; `git log` the PATH,
> never the number. Added to CLAUDE.md's collision list 2026-09-03.

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

> **⚠ CORRECTED 2026-09-03, by the council's `debug_historian` seat (corr `c2d1d570`), and it
> was my overclaim.** The before/after runs below are an **offline reproduction against a copy
> of the live register**, using a locally built `cmd/regcheck`. They are NOT a post-deploy
> observation: at the time of writing, `3f221f99f` is committed and **inert until the next
> chassis roll**. The distinction matters because this repo has a standing failure mode of
> reading "confirmed live" into exactly this shape of evidence. What the runs DO establish is
> the causal claim — same register, same binary, one fact repaired, opposite outcome — which
> is what a control is for. What they do not establish is that anything has changed in
> production yet. **The post-roll check is in §8.**

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

`finetuning.uk`'s inert bans are the same shape — its own owner's ruling (same offline repro):

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

## 8. Post-deploy verification — owed, and NOT yet done

Named because the council's `debug_historian` seat pointed out the submission had no
post-roll step, and `ParseEvidenceBase` runs inside the **agent-chassis binary**
(`validate_page_content`, `check_unverified_claims`), not in any tool a session runs by hand.
Per CLAUDE.md, ask the artefact, not git and not the tag:

```bash
# 1. which commit is the running chassis built from (per SERVICE, and it is a STARTUP line)
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor 3f221f99f <that sha> && echo "the parse fix is in the running build"
git merge-base --is-ancestor e5b41dc31 <that sha> && echo "the reporting half is in too"

# 2. the demand control: the two registers must move from FAILING to PARSING.
#    Re-run the 27-register census in §2 — 25 must be unchanged, 2 must flip.
#    A census that shows 27 OK proves nothing on its own unless you know it
#    showed 25 before (and see the stdin trap in the §2 recipe).

# 3. the finding must SURFACE, which is the half no unit test can prove:
#    two malformed_evidence_fact items, one per site, naming the fact.
SELECT s.domain, swi.item_key, swi.spec->>'fact_id', swi.spec->>'bans_at_stake'
FROM site_work_items swi JOIN sites s ON s.id = swi.site_id
WHERE swi.item_type = 'malformed_evidence_fact';

# 4. and the counter, on a site with NO malformed facts, as the positive control:
#    relojistas.com must report facts_unverifiable = 12, gamesdesign 1, oufe absent.
```

**A second sweep must REFRESH, not duplicate.** The key is per fact, so a second daily pass
over an unrepaired register must not add rows.

## 9. Answers to the council's advisory objections (corr `c2d1d570`, APPROVED round 1)

Recorded here rather than left in the verdict, because two of them are open questions a later
session should be able to find.

- **`reuse_agent`: "no evidence you searched for an existing tolerant-decode helper, or an
  existing malformed-data item_type."** Fair — the submission asserted novelty without showing
  the search. Both were then run. **No tolerant/partial array-decode helper exists** anywhere
  in `platform/`, `internal/` or `pkg/` (searched for the shape and for the naming). The two
  candidate item types are **not** substitutes: `required_fields_missing` is a
  content-completion item with its own **automated handler** (`required-fields-missing-handler`,
  `platform/livespec/unarmed_completers.go`) that auto-closes on convert/resolve/stale, so
  routing a broken register through it would hand the finding to a completer that cannot fix it;
  `cta_tel_malformed` is specific to CTA phone numbers.
- **`bug_historian` [medium]: the same all-or-nothing decode shape may exist at other
  `site_specs` aspects** (`content_direction`, `design_intent`, `imagery_style_guide`), and this
  fix is bespoke to `EvidenceBase` rather than a shared helper. **Genuinely open, and not
  closed here.** A first look found no equivalent named typed parser for those aspects, so the
  defect is not obviously present — but that is an absence I did not prove, and the seat is
  right that the *mechanism* (a Go typed decode of a jsonb array voided by one bad element) is
  generic. **Follow-up sweep owed; do not read this bullet as clearance.**
- **`architecture` [low]: `malformed_evidence_fact` has no retirement condition.** Correct, and
  the seat named the trigger precisely: this item type converts a silent total failure into a
  recurring manual queue for a capability the schema still cannot express. `EvidenceFact.Value`
  is a `*float64`; a typed union (string or number) would close it for good. **The stated
  trigger: when this item type's volume is next measured, if it is still growing, that is an
  RFC, not another sweep tweak.** Today's expected volume is 2.
- **`guardian` [medium]: heavy concurrent edit activity on `refresh_evidence_base_action.go`.**
  Real, and it bit: at commit time that file carried an uncommitted `livespec` call site from
  another lane, so the reporting half was **held out of the first commit** and landed
  separately (`e5b41dc31`) once their helper was committed (`1802359a6`). The `sql`, `citation`
  and `artifact_check` arms are byte-identical.
- **`editquality` / `guardian` [low]: the embed-and-shadow decode is subtle.** A comment now
  names both ways a later edit breaks it silently, and points at the test that catches it.

## 10. The diagnosis loop: UNVERIFIABLE, and the reason is MY symptom, not the finding

Run per the 2026-07-31 ruling (a `bugs_open/` file asserting a cross-cutting root cause goes
through the loop, or the session states plainly what it substituted). Intake correlation
`6ab3fc01-c743-4cc8-b0c8-7dc6cda3d58d`.

**Verdict as returned, verbatim and not softened:** `status = UNVERIFIABLE`,
`conclusion = NOT CONFIRMED (stopped: scope-not-narrowing)`, `is_fix = false`. **It is not a
REFUTED and it is not a CONFIRMED.** It assembled the evidence bundle — `ParseEvidenceBase`,
`checkAttestationStaleness`, `checkBannedClaimPatterns`, `createStaleAttestationItem`,
`createInvalidBannedClaimPatternItems`, `composeWriterBlock` — and then stopped without
grading anything.

**Why, and it is my error.** CLAUDE.md's symptom-authoring rule is *one coherent bug per run*.
I submitted **two** mechanisms in one symptom: the all-or-nothing decode voiding the register,
AND the residue arm dropping facts uncounted with the early return in front of every raise.
Those share a subject and nothing else — different files, different failure modes, different
evidence. The loop had no single scope to narrow toward, so it burned a full run and returned
no verdict on either half. **A two-part hypothesis is not gradable, and "they are both about
the evidence register" is not coherence.**

**What this bug rests on instead, declared as the ruling requires.** Not the loop, and not
first-hand reading either — a **before/after control on the live registers**: the same binary
and the same register, with one fact repaired, produces the opposite outcome (§2). That is
stronger than a code-read for this particular claim, because the claim is about what the
decoder *does*, and the control could have come out otherwise. What the loop could have added
— an independent look for a cause I had not considered — is the thing I did not get, and I
should have got it by splitting the symptom in two.

**If you re-run it, run it twice:** one symptom for the decode-voids-the-register mechanism,
one for the residue arm and the early return.

## 11. INDEPENDENT CORROBORATION, same day, different consumer — and it raises the typed-union case from "tidy" to "owed"

Found while checking whether anything had moved under this lane. **The `mortgagecalculator.co.uk`
adoption lane hit the same root cause on 2026-09-03, in a completely different reader, with no
knowledge of this bug** (commit `7991c3191`):

> `verify_criteria.py`'s `load_register_bands()` read EVERY fact in the site's `evidence_base`
> and `float()`'d it. The register now holds 18 facts, of which 5 are `CIT-*` citations with no
> scalar value, so the loader raised `ValueError` at import time and **`install_fences.py` could
> not run at all.**

That is this bug's mechanism one consumer over: **a reader assuming every fact carries a
number, meeting a fact that legitimately does not, and taking a whole tool down with it.** They
also hit a second shape worth knowing — *"a fact whose value is JSON null makes `->>` return SQL
NULL, and `psql -tA` prints that row with NO trailing separator"*, so a `split('\t')` unpacked
into the wrong arity.

**Why this matters here, and it changes the argument.** §7.3 and RFC_025 §12.5 record the
text-valued fact shape as an *open option* — the honest end state, not urgent, deliberately
deferred. Two independent consumers breaking on the same assumption within one day is a
different claim from "one register is malformed":

| | this bug | the mcalc lane, same day |
|---|---|---|
| consumer | `ParseEvidenceBase` (Go, every claims gate) | `verify_criteria.py` (Python, the acceptance fences) |
| trigger | a **text** value (`"MIT"`, `"30 days"`) | a **value-less** citation fact, and a **null** one |
| blast radius | the site's whole register, bans included | the tool could not start |
| found by | an audit of all 27 registers | the tool falling over |

**`EvidenceFact.Value` being a bare `*float64` is not a local wart; it is a seam that every
reader of the register has to defend itself against, separately, and two of them did not.**
Neither lane's fix generalises: mine makes one bad fact survivable inside the Go parser, theirs
skips unparseable facts inside one Python loader. The next reader starts from zero again.

**So the retirement condition in §9 is met sooner than I expected.** The architecture seat's
trigger was "if the item type's volume is still growing when next measured". A second, wholly
independent consumer failing on the same assumption is stronger evidence than volume would have
been, because it shows the defect is in the **contract**, not in the data. Recorded rather than
acted on — promoting it is an RFC and belongs to whoever owns the vocabulary — but a future
session should not have to rediscover that this happened twice in one day.

**Told to:** the `finetuning` lane (asked for their view, since 8 of their 10 facts want the
shape) and `register_guards_code_phase_b` (owner of RFC_025 stage 2). The mcalc lane's own fix
stands and is not affected by anything here.

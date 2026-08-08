# HANDOFF (2026-08-08) — provocation gate: two calibrations run, the second one moved the problem, and the fix is committed but not yet built

**Cold start? Read this file, then only §3 of
`HANDOFF_2026-08-05_calibration_failed_and_the_spec_is_why.md`** (its diagnosis is
still correct; its "what to do" is superseded by the rulings below).

Predecessors, in order: `HANDOFF_2026-08-02_continue_here.md` (the publisher half),
`HANDOFF_2026-08-05_…` (first calibration). **Everything either of them asks for is
either done or superseded here.**

---

## 1. One-paragraph state

The whole provocation pipeline is **built, council-approved and live in the chassis**;
**nothing is wired to publish**; and vonc.com is still serving a **26 July**
provocation while promising a daily one five times on its home page. Two live
calibrations have now run. The second scored **6 of 9** on the owner's real
provocations (up from 4 of 9) and **4 of 4** on the deliberately bad set. The three
remaining rejections are understood, and the fix for two of them is **committed but
not yet in a build**. One is a content gap that belongs to the framework.

## 2. What is live, proven at the artefact

| thing | state | evidence |
|---|---|---|
| `gate_provocation`, `generate_provocations`, `schedule_provocations` | live, chassis **v1.0.1264**, both replicas | pod-grep, below |
| council verdict | **APPROVED r1**, 6 advisory objections, all discharged or recorded | corr `28056723-b2a3-4057-b92f-482b7f7a0e72` |
| wired to publish | **NO** — 0 agent defs reference them (excluding the calibration agent) | re-checked 2026-08-08 14:24Z |
| production pool | 9 approved, newest `publish_on` **2026-07-26**, untouched | re-checked after every run |

**Provenance for v1.0.1264 — the best control this lane has had.** `not_contestable`
1, `one_sided` 1 (both ADDED by the 08-06 rulings), and **`not_two_sided` = 0** — a
string the change genuinely REMOVED, which is a true negative control rather than the
synthetic one used previously. Pre-existing control `gate_provocation` = 3. Identical
on both replicas.

## 3. Owner rulings so far (all applied)

1. **2026-08-06 — "the one sided provocation is better."** Two-sidedness is no longer
   fatal; recorded as `one_sided`. Its old justification was false anyway (5 of 9
   carry a counter-case, not 9 of 9).
2. **2026-08-06 — "can we just remove that rhetoric?"** Applied as a minimal deletion
   to production. **See §4: it did not hold.**
3. **2026-08-06 — "we want it all done through the framework, so we don't want you
   writing things yourself."** General. Saved as memory
   `the-framework-writes-the-content-not-you`. Applied immediately to the calibration
   fixture, which is now **generated from the pool, not typed.**

## 4. The second calibration, and why it is the interesting one

Run on v1.0.1264, `orchestration_id=65afc0b3-48c4-450d-9780-489f2696b8ad`, model
`claude-sonnet-5`, all 13 judged.

```
must APPROVE (the 9 real)   approved 6   rejected 3      (was 4/9)
must REJECT  (the 4 bad)    rejected 4                   PASS, again
```

Trending slop is now caught by **`not_contestable`**, which is exactly what that
criterion was added for after ruling 1 removed the only thing catching it.

**I predicted 8 of 9 and got 6. The two extra rejections are the finding:**

- **`four-day-week-productivity-myth` — still rejected, and the deletion is why we
  know.** The requested clause was removed cleanly. The model then flagged the
  *adjacent clause of the same sentence*: "recruit organisations that already believed
  and run them for six months" is "an overstated generalization". **Removing rhetoric
  phrase by phrase is whack-a-mole** — the provocation *is* a rhetorical
  generalisation about pilots, so there is always another clause.
- **`nobody-reads-terms-of-service` — newly rejected**, on "Reading takes an hour" (a
  figure of speech) and "every study that frames it as apathy" (a rhetorical
  generalisation). Neither is a fabrication.

**Diagnosis: the factual check was never catching what it was built for.** PLAN §4 and
`bugs_closed/043` aimed at generated copy **inventing** quantitative claims. The
implementation was penalising **unsupportedness**, which is the ordinary register of
argumentative prose. A provocation that must cite every generalisation is not
writable, and the failure is **silent** — the pool starves while the gate reports
itself working.

### 4a. The judge is STOCHASTIC, and this changes what "calibrated" means

`nobody-reads-terms-of-service` drew **no** factual objection on 05 Aug and **two** on
08 Aug from **byte-identical text** ("Reading takes an hour" present in both runs;
verified by query, not by memory). So:

> **A single green calibration is not evidence. It has to pass repeatedly.**
> Whoever declares this gate calibrated should run it at least three times and require
> all three, or state plainly that they did not.

## 5. What is committed and NOT yet built

`103fa6e30` narrows the factual criterion to **fabrication**:

- *reportable*: an invented study/institution/person; an invented statistic presented
  as sourced; a specific checkable statement that is actually false.
- *NOT reportable*: idiomatic quantities; sweeping generalisations about a category;
  anything merely unsupported / uncited / unverified / "overstated".
- The line in the prompt that carries it: **"The test is INVENTED, not UNCITED."**

**This is a correction toward the stated intent, not a relaxation.** Fabrication stays
fatal — that is the half `bugs_closed/043` was about. Safety, contestability and the
deterministic layers are all untouched.

> ### THE ONE THING OWED: a fresh chassis build, then re-run.
> Go changes are inert until an image is rebuilt and rolled. `103fa6e30` is committed;
> the running chassis is v1.0.1264, which predates it.
> After the roll: pod-grep for `INVENTED, not UNCITED` (positive) and
> `invents a statistic, study, quantity or named source` (**negative, expect 0** — the
> old wording), then dispatch as in §6.

**Expected on the third run, stated up front so it can be judged rather than
admired:** **8 of 9** approved; the ninth — `group-chats-replaced-friendship` —
rejected as `body_too_short`, because its body is **zero characters** in the
production pool. **That is a POOL defect, not a gate defect, and by ruling 3 the
framework must generate that body — do not write it.** Bad set 4 of 4. Anything else
is a new finding; write it down rather than adjusting the gate to match.

## 6. Commands

```bash
# scorecard
SELECT CASE WHEN source_ref LIKE '%must-approve half%' THEN 'must APPROVE' ELSE 'must REJECT' END,
       status, count(*) FROM provocations
 WHERE domain='calibration.vonc.com' AND gated_at IS NOT NULL GROUP BY 1,2 ORDER BY 1,2;

# why each rejection (the interesting half — §10.3)
SELECT slug, (SELECT string_agg(r->>'detail', ' // ') FROM jsonb_array_elements(gate_verdict->'reasons') r
              WHERE (r->>'fatal')::bool)
  FROM provocations WHERE domain='calibration.vonc.com' AND status='rejected' ORDER BY slug;

# reset for a fresh round (the gate never re-judges a judged row, on purpose)
UPDATE provocations SET status='draft', gated_at=NULL, gate_verdict=NULL
 WHERE domain='calibration.vonc.com';
# re-copy the must-approve half if production text has changed since:
#   DELETE ... WHERE source_ref LIKE '%must-approve half%'; then the INSERT…SELECT in migration 319.
```

Dispatch: `agent_type='provocation-gate-calibration'` on
`system.agent.generic.requests`. Envelope: copy `097_TRIGGER_council_review_v1.sh`
lines 168–188. **One-line payload — `kcat -P` splits stdin on newlines into separate
messages.** No dispatch within ~300s of a chassis restart; it is silently dropped.

**The harness cannot touch production, by data not by care:** `calibration.vonc.com`
is absent from `sites`, and `render_provocation_feed` calls `assertKnownDomain`, which
refuses a domain that is not a site. Migration **319** asserts that with a `RAISE` and
refuses to commit if the domain ever appears in `sites`.

## 7. Owed, in order

1. **Build + roll, then re-run** (§5). Three runs, not one (§4a).
2. **The empty body for `group-chats-replaced-friendship`** — framework's job, ruling 3.
3. **Then wire it**: an `agent_definitions` row + a `scheduled_tasks` row. **The
   `architecture` seat asked that the WIRING submission get its own council round** —
   the existing approval does not cover it.
4. **Carried, still unfixed, still correct:** `nextPublishDates` will collide with
   RFC_013's per-category index change; **both halves must become category-aware in
   the same change** or a category is silently never scheduled. Landmine on VONC-012;
   contributed to their RFC as §8.
5. `bugs_open/098`-adjacent, tiny: `/blog/provocation.html` is `status='active'` with
   0 components and `deployed_at IS NULL` — a never-built plan row, not a lost file.
   Belongs to whoever owns vonc's page inventory.

## 8. Traps this lane has paid for (read before touching anything)

- **A roll is not evidence.** Pod-grep an added string **and a removed one**. v1.0.1264
  finally had a real negative control (`not_two_sided` = 0); earlier runs had to use a
  synthetic one because the changes were purely additive.
- **The feed is read SERVER-SIDE.** `tools-api round.go FetchProvocation` takes the
  whole `today` object and persists it as the round's provocation. Never "seal" a
  provocation by emptying `today`.
- **An artefact rollback is not a rollback here.** Restoring `provocations.json` is
  undone within 6h when the publisher re-derives from the pool.
  `builder/rollback_provocation.sh` retires the ROW instead; dry-run by default, and
  the preview IS the real UPDATE inside a rolled-back transaction.
- **A fixture you compose to exercise a rule will exercise the rule.** The first
  calibration passed 9/9 against bodies I had written myself while claiming
  "verbatim". `WRONG_CALLS.md` 2026-08-05. Generate fixtures; do not type them.
- **A positional reference into generated data is a hidden dependency on row order.**
  Regenerating the fixture moved the zero-body row to index 0 and broke seven
  fail-closed tests — which still "passed" as rejections, and were caught only because
  each asserts *why* it rejected. Use `aGoodCandidate(t)`.
- **The shared tree may not compile and it may not be you.** On 08-06 another session
  had five files uncommitted and mid-edit. Test against `git archive HEAD` with only
  your files overlaid, and reap the extraction afterwards.
- **`cd` in a compound command resets the session cwd** and can leave you in a deleted
  directory. Use absolute paths.

# HANDOFF (2026-08-05, evening) — the live calibration FAILED, and the gate is not the thing that is wrong

**Start here.** The generative half is built, council-approved, and live in the
chassis. The §10.6 calibration has now been run against a real model, and it
**failed** — in a way that indicts the specification rather than the code, and
which needs **two owner decisions** before anything can be wired.

Do not "fix" the gate to make the calibration pass. See §4.

---

## 1. State, all of it verified rather than assumed

| thing | state | evidence |
|---|---|---|
| gate / generator / scheduler code | **LIVE**, chassis `v1.0.1254`, both replicas | pod-grep with controls, §1a |
| council verdict | **APPROVED round 1**, 6 advisory objections, none high | corr `28056723-b2a3-4057-b92f-482b7f7a0e72` |
| wired to publish? | **NO — 0 `agent_definitions` reference the 3 actions** | re-checked 20:57Z |
| §10.6 live calibration | **RUN, and FAILED** | §2 |
| vonc.com's daily promise | still FALSE — serving 26 Jul, 10 days stale | `today.date` on the served feed |

### 1a. Deploy provenance (the roll was another session's, so it was proven, not trusted)

`strings /app/agent-chassis | grep -c` on **both** replicas, identical:
`gate_provocation` 3 · `generate_provocations` 3 · `schedule_provocations` 2 ·
`judge_unavailable` 1 · positive control `render_provocation_feed` 1 ·
synthetic negative control 0. (`bugs_open/153`: a roll is not evidence your fix
shipped.)

---

## 2. The calibration result

Harness: migration **319**, agent `provocation-gate-calibration`, domain
`calibration.vonc.com`, model **`claude-sonnet-5`**, dispatched
`orchestration_id=f0d3f9fe-25b0-433b-b582-ea9b1c0a9144`. All 13 candidates judged.

```
must APPROVE (the 9 real)   approved 4   rejected 5     <-- FAIL
must REJECT  (the 4 bad)    rejected 4                  <-- PASS, 4/4
```

**The bad set is caught perfectly.** Every one of the four kinds §10.6 names — a
bare insult, a factual claim dressed as opinion, a one-sided political take,
trending slop — was rejected. The false-negative direction, the one that puts a
falsehood on a live homepage, is clean.

**Five of your nine published provocations were rejected**, which is the
false-positive direction: a gate that refuses the entries the owner actually
published would silently starve the site.

| slug | verdict | fatal rule |
|---|---|---|
| ai-never-funny-on-purpose | approved | — |
| data-driven-decisions-arent | approved | — |
| fiction-makes-you-worse-at-facts | approved | — |
| remote-work-killed-mentorship | approved | — |
| **four-day-week-productivity-myth** | rejected | `factual_problem_in_body` |
| **group-chats-replaced-friendship** | rejected | `body_too_short` |
| **nobody-reads-terms-of-service** | rejected | `not_two_sided` |
| **nobody-wants-personalised-internet** | rejected | `not_two_sided` |
| **privacy-is-already-over** | rejected | `not_two_sided` |

---

## 3. Why — and the judge is not the problem

I tested PLAN §4's central form claim against the whole live corpus rather than
against the handful of entries it was written from:

```sql
SELECT slug,
  CASE WHEN COALESCE(body,'') ~* '(the counter is|the rebuttal is|against that|the defence is)'
       THEN 'HAS explicit counter' ELSE 'no explicit counter' END, length(COALESCE(body,''))
FROM provocations WHERE domain='vonc.com' AND status='approved';
```

**5 of 9 have an explicit counter-case. 4 do not.** And the correlation with the
verdicts is exact: every entry with a counter-case was approved on that criterion;
every entry without one was rejected. `group-chats-replaced-friendship` has a
**body of 0 characters** in the live pool.

> ### PLAN §4 says: *"genuinely two-sided — **every** `detail_body` makes the case
> and then makes the counter-case"*. **That is false about the live corpus.** It is
> true of the five entries it was evidently written from.

I encoded that claim as a **fatal** rule. The gate then did exactly what it was
told, and the model judged the text correctly. **A well-calibrated judge applied to
a wrong specification looks identical to a broken judge** — the only way to tell
them apart was to read the four rejected bodies, which is why the reasons are
persisted (§10.3 earning its place immediately).

### 3a. The one genuinely debatable case

`four-day-week-productivity-myth` **has** a counter-case and was rejected for a
factual problem. The model's words:

> `"measure self-reported output"` — Overgeneralized claim about all four-day week
> pilots; some (e.g., UK 2022 pilot) used company-reported KPIs/revenue alongside
> self-report, not purely self-reported output.

That is a defensible objection. **It is also the exact phrase PLAN §4 uses as its
worked example of a *legitimate* factual claim inside a good provocation.** So the
plan's own illustration of the thesis/prose split is the thing the split rejects.
The underlying tension is real: a provocation's supporting prose is *rhetorical*
and will always generalise a little, and a competent fact-checker will flag that.

---

## 4. DO NOT relax the rules to make this pass

The tempting fix — drop `not_two_sided` to advisory, lower `minBodyLen` — is
`fixing-a-checker-to-agree-with-a-broken-site`, which this estate has an explicit
practice against. It would also destroy the only evidence the gate works: a
calibration you tuned until it passed measures the tuning, not the gate.

**Two decisions are the owner's, and neither is an engineering call:**

1. **Is two-sidedness a real requirement for a provocation?**
   - *If yes* — the gate is correct as built, and **4 of the 9 already-published
     provocations do not meet the bar you want for new ones.** That is a coherent
     position (the bar rises; the old entries stay as they are). The calibration
     corpus must then be reduced to the 5 that meet it, and §10.6's "all 9" becomes
     "all 5", recorded as a deliberate narrowing with this evidence attached.
   - *If no* — `not_two_sided` becomes a recorded-but-not-fatal reason, joining
     `interesting`/`current` in the advisory block. The gate keeps safety and
     factual checks as the only fatal ones.
2. **Is "the pilots measure self-reported output" acceptable rhetoric or a factual
   problem?** Your answer sets how strict `factual_problem_in_body` should be. If
   rhetorical generalisation is acceptable, the judge prompt needs to say so
   explicitly — it currently invites exactly this objection by asking for anything
   "specific and checkable that is false".

---

## 5. What I got wrong, and it is the most transferable thing here

**My unit calibration was measuring prose I wrote myself.**
`provocation_gate_action_test.go` states the nine are *"copied from the live pool
on 2026-08-05 … reproduced verbatim rather than paraphrased: a paraphrase would
calibrate the gate against my idea of a provocation instead of against the
owner's."*

**The titles and teasers are verbatim. The BODIES are not.** Eight of the nine live
rows have an empty `body` column, and I composed long-form bodies — complete with
tidy "The counter is…" turns — for the test. So the stubbed calibration passed 9/9
against text that demonstrated the very property the live text lacks. **I wrote the
counter-cases I then tested for.**

The live run is what caught it, which is precisely the argument §10.6 makes for
existing. Logged in `WRONG_CALLS.md`.

---

## 6. To pick this up

```bash
# the scorecard
SELECT CASE WHEN source_ref LIKE '%must-approve half%' THEN 'must APPROVE' ELSE 'must REJECT' END,
       status, count(*) FROM provocations
 WHERE domain='calibration.vonc.com' AND gated_at IS NOT NULL GROUP BY 1,2 ORDER BY 1,2;

# why each one was rejected (the interesting half)
SELECT slug, status,
       (SELECT string_agg(r->>'detail', ' // ') FROM jsonb_array_elements(gate_verdict->'reasons') r
         WHERE (r->>'fatal')::bool)
  FROM provocations WHERE domain='calibration.vonc.com' AND gated_at IS NOT NULL ORDER BY slug;

# re-run a fresh round (the gate never re-judges a judged row, on purpose)
UPDATE provocations SET status='draft', gated_at=NULL, gate_verdict=NULL
 WHERE domain='calibration.vonc.com';
# then dispatch agent_type 'provocation-gate-calibration' on
# system.agent.generic.requests (envelope: see 097_TRIGGER, one-line payload —
# kcat -P splits stdin on newlines into separate messages)
```

**The harness cannot touch production, by data rather than by care:**
`calibration.vonc.com` is absent from `sites`, and `render_provocation_feed` calls
`assertKnownDomain`, which refuses a domain that is not a site. Migration 319
asserts that isolation with a `RAISE` and refuses to commit if the domain ever
appears in `sites`.

## 7. Owed, in order

1. **The two decisions in §4.** Everything else waits on them.
2. Re-run the calibration after whichever change they imply. It must pass before
   anything is wired — §10.6, and the compliance seat restated it independently.
3. Then, and only then, wire it: an `agent_definitions` row plus a `scheduled_tasks`
   row. **The `architecture` seat asked that the WIRING submission get its own
   council round** — this approval does not cover it.
4. Carried over, still unfixed and still correct: `nextPublishDates` will collide
   with RFC_013's per-category index change; both halves must become category-aware
   in the same change. Recorded as a landmine on VONC-012 and contributed to their
   RFC as §8.

## 8. Nothing was risked

Production pool re-checked after the run: **9 approved, newest `publish_on`
2026-07-26 — unchanged.** No live row was read for update; the calibration copies
live under a different domain, and `(domain, slug)` uniqueness makes the copies
incapable of colliding with the originals.

---

# ANSWERED 2026-08-06 — both owner decisions given, code changed, RE-RUN OWED

**§4's two decisions are settled. This handoff's diagnosis stands; its "what to do"
is now history. Read this section first.**

## Ruling 1 — two-sidedness is NOT required

> *"I think the one sided provocation is better, we can continue that way for now
> at least."*

So `not_two_sided` is no longer fatal. It is recorded as an advisory note
(`one_sided`) alongside interesting/current, because the ruling is explicitly
provisional and "how many are one-sided" stays worth asking.

**The calibration immediately found a consequence I had not predicted.** Removing
two-sidedness removed the **only** criterion that was rejecting §10.6's *trending
slop* sample — the suite promptly approved "AI is changing everything". The ruling
did not license letting slop through; it opened a gap.

**Filled with CONTESTABILITY, fatal:** could a reasonable, informed person argue the
opposite? That is precisely what separates the one-sided provocations the owner
prefers from filler — *"Privacy is already over"* takes a disputable position;
*"AI is changing everything"* states something nobody disputes. All nine live
entries satisfy it, so it costs the corpus nothing.

`minBodyLen` was **re-justified rather than left standing**: its old reason ("too
short for a case AND a counter-case") was withdrawn by the ruling, and a floor whose
stated reason has gone is a number nobody can defend. It now means *there must be
something to judge*, sized against the measured live spread (326..607 for the eight
non-empty bodies).

## Ruling 2 — remove the rhetoric, don't loosen the check

> *"'the pilots measure self-reported output' — can we just remove that rhetoric?"*

Applied to the **production** pool as a minimal deletion, in a guarded transaction:

```
BEFORE  The pilots recruit organisations that already believed, run them for six
        months with everyone watching, and measure self-reported output.
AFTER   The pilots recruit organisations that already believed and run them for six
        months with everyone watching.
```

Nothing added, one connective repaired. Verified in-transaction: the phrase is
absent and the next sentence ("That is a design which cannot return a negative
result") is intact. **The factual check stays strict** — which is the right half to
keep strict, since it is the one guarding against falsehoods on a live page.

## Ruling 3 (general, and it outlives this lane)

> *"we want it all to be done through the framework, so we don't want you writing
> things yourself."*

Applied first to my own worst habit here: **the calibration fixture is now
GENERATED, not typed.** It is serialised straight from
`SELECT slug,title,teaser,COALESCE(NULLIF(body,''),detail_body) …` rather than
hand-transcribed. The previous version claimed "verbatim" and its bodies were mine
(§5).

That regeneration **broke seven tests, correctly**: `realProvocations[0]` is now the
zero-body row, so fail-closed tests were rejecting on `body_too_short` before the
judge ran. They still "passed" as rejections — caught only because each one asserts
*why* it was rejected. Replaced by `aGoodCandidate(t)`, which states the property
instead of trusting a row order that comes from a live table. Saved as a memory:
`the-framework-writes-the-content-not-you`.

## State now, and the ONE thing owed

- Code committed (`3b473b8dd`); unit suite green (run in an isolated `git archive
  HEAD` tree, because another session has five files uncommitted and mid-edit and
  the shared tree does not compile — none of it mine).
- Production pool: rhetoric removed, 9 approved, newest `publish_on` 2026-07-26.
- Calibration corpus **refreshed from the corrected production text** and reset to
  ungated: 9 real + 4 bad, all `draft`, guarded so a stale copy cannot pass.

> ### OWED: a fresh chassis build, then re-run the calibration.
> The rulings changed **Go code**, which is inert until an image is rebuilt and
> rolled. The running chassis is `v1.0.1254`, which still has two-sidedness fatal and
> no contestability check — **a re-run against it would measure the old gate.**
> After the roll: pod-grep for `not_contestable` and `one_sided` (with a positive
> control), then dispatch `provocation-gate-calibration` as in §6.

**Expected on the re-run, so it can be judged rather than admired:** 8 of 9 real
approved; `group-chats-replaced-friendship` rejected as `body_too_short` because its
body is genuinely empty in the pool — **a defect in the POOL, not the gate, and one
the framework should fix by generating a body, not me by writing one**; 4 of 4 bad
rejected, with slop now caught by `not_contestable` rather than `not_two_sided`.
Anything else is a finding.

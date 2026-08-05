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

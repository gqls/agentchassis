# HANDOFF 2026-08-12 — bugfix 213, continue here (supersedes the 08-10 handoff)

**Read this first, then `NOTES…md` (bottom third) for the missteps and mutation
matrices, then `PLAN…md` §"OWNER DECISIONS" for the rulings and their reasoning.**
The 08-10 handoff is still accurate about the instance fix and about D3's brief; it is
superseded only on "what happens next".

---

## THE ONE-PARAGRAPH STATE

The instance fix (**WII-013**) is live and pod-proven on the chassis. The class
detector (**WII-015**, `cmd/verifier-remit-check`) is built, council-APPROVED, deployed
as a daily CronJob, and proven at the artefact — it evaluates 12 verified item_types,
finds 0, and names the one it suppressed. **D3 is DONE.** What is left is **D1**: the
new item_type `dark_section_audit` has no verifier, so the rotation now re-finds these
defects and closes them **unchecked** — measured, not feared: 14 items filed 2026-08-11,
**all 14 complete, 0 of 14 carrying any `_verification` key.** D2 depends on D1 and is
otherwise a no-op. The bug file itself now records that its **closure criterion has
become unsatisfiable**, with three costed options and a note that the choice is the
owner's.

---

## START HERE — D1, and what today's evidence changed about it

**The ruling (owner, 2026-08-11): build a verifier over `spec.acceptance_test` using the
`criteria_check` vocabulary (RFC_002), and settle the handler routing.** Unchanged. What
changed is that it is no longer a theoretical gap.

### The evidence to open with (all [MEASURED 2026-08-12], all re-runnable)

```sql
-- the bleed: filed by the sweep on 08-11, closed the same day, graded by nothing
SELECT status, count(*), count(*) FILTER (WHERE result ? '_verification') AS carry_verification
FROM site_work_items WHERE item_type='dark_section_audit' GROUP BY 1;
--  complete | 14 | 0

-- the control that makes that asymmetry mean something (same window, same producer label)
SELECT count(*) FILTER (WHERE result ? '_verification') AS carry, count(*) AS total
FROM site_work_items WHERE item_type='hardcoded_section_colors';
--  9 of 9 producer-A rows carry one
```

The mechanism is read, not inferred (the `bugfix_122` lane read it and I checked the
same lines): `verifyBeforeComplete` resolves by `checks.GetVerifier(itemType)`
(`complete_work_item_verification.go:70`); an **unregistered** type is documented to
complete as before (`:16`) and no `_verification` key is written at all. The
`out_of_scope` branch (`:112`) needs a **registered** verifier that declines. So
`dark_section_audit` completions are untouched by construction.

### Do these two things in this order

1. **MEASURE BEFORE RE-ROUTING — this is still the first task and it is still not done.**
   Check each live `acceptance_test` against `ReplaceHardcodedColors`' actual remit.
   Only gamesdesign's already-`var()` case is confirmed OUTSIDE it; several others name
   inline `style` attributes and `rgba(0,0,0` literals that may be INSIDE it. "The fixer
   cannot repair these" is `[UNVERIFIED]` as a generalisation — generalising from the one
   worked instance is the exact move this bug exists to punish.
2. **Then the verifier, and treat its design question as the whole job.** Putting a
   browser / computed-style evaluation on the **completion path** is what this estate has
   deliberately kept free even of HTTP probes (`verifier_coverage_test.go:171` records the
   standing objection). It needs its own council round. **Argue from `contrast_failure`,
   and note it is now ONE decision, not two:** both types are classified `catMechanical`
   in the coverage guard with the *same* stated posture ("verification needs a browser;
   the NEXT audit plus the dedup key is the re-detection, and two-strike escalates a
   persistent pairing"). The `bugfix_122` lane has **226 `contrast_failure` rows parked**
   on exactly this question and has withdrawn its earlier claim that they unpark when 213
   closes — they unpark when `contrast_failure` gets a verifier, or when someone rules it
   does not need one. **So D1's council round decides both lanes.** Talk to that lane
   before you submit; do not decide it twice.
3. If you give the new type a verifier, the `sql_for_agents/220` claim-timeout exclusion
   must move in the **same** change — `TestRegisteredVerifiersMatchClaimTimeoutExclusion`
   enforces both directions — and its coverage-guard entry must move out of
   `itemTypesWithoutVerifiers` (the guard refuses a type that is both).

---

## DECISION OWED BY A HUMAN — this bug can no longer close on its own terms

`bugs_open/213`'s recorded criterion is *"stays OPEN until a `hardcoded_section_colors`
item without `spec.check` reaches completion and lands `triaged`/`failed` with the
scope-mismatch error"*. [MEASURED 2026-08-12] the design-audit producer's newest
surviving row under that type is **2026-08-09**; Half A moved that producer to
`dark_section_audit` permanently. **The fix removed the traffic that would have
demonstrated the fix.** `out_of_scope` has been 0 for two days and can now only become
non-zero if a NEW producer converges on that type — which is precisely the event
`verifier-remit-check` exists to catch.

Options, written into the bug file, none of them a thread's call:
**(a)** accept the unit + mutation proof and close, recording the one unexercised branch;
**(b)** exercise it deliberately with one synthetic no-`spec.check` row on a throwaway
site driven to completion (real proof, real dispatch, needs an owner's yes);
**(c)** leave OPEN, accepting that the file stops describing a reproducible defect.

---

## WHAT IS LIVE, AND HOW TO CHECK IT WITHOUT RE-DERIVING ANYTHING

| thing | state | how to check |
|---|---|---|
| WII-013 instance fix | LIVE, chassis `v1.0.1290` | binary probe recipe in the 08-10 handoff (both replicas, two-sided control) |
| WII-015 detector | LIVE, CronJob `25 7 * * *` UTC, image `v1.0.1289` | `doc_notes WHERE source='verifier-remit-check'` — one row per run, clean or not |
| detector findings | **0, correctly** | `site_work_items WHERE item_type='verifier_remit_gap'` |
| the gate (`out_of_scope`) | **0, and now unfireable** — see above | `site_work_items WHERE result->'_verification'->>'status'='out_of_scope'` |

Run the detector yourself in seconds: `go run ./cmd/verifier-remit-check --dry-run`, and
`--ignore-remit` for the disconfirmability control (writes refused) — it reproduces the
original bug as a live finding, which is what makes the daily 0 an honest 0. **From a
terminal it never writes** (no `PG_CLIENTS_HOST`); it says so in its own report.

---

## OPEN, FILED, NOT THIS LANE'S TO ADOPT

- **`RFC_024`** — there are **nine** CronJob meta-checks with no shared harness, and
  three council seats have asked for a consolidation pass across two rounds. Filed with
  the population, what is duplicated, four costed options and a recommendation (2 then 3,
  not 4). Nobody has picked it up.
- **12 live item_types are in neither half of `verifier_coverage_test.go`** (89 rows;
  `lock_blocked_change` 37, `chrome_divergence_overwritten` 19, `save_refused_incomplete`
  16, and nine more). Contributed as a census comment in that file with its refresh query;
  deliberately **not adopted** into either map, because that file's own rule is that
  adding a type is a commitment about someone else's producer. It belongs to
  `bugs_open/021` §INSTANCE 2.
- **Two `design-audit` rows vanished from `site_work_items` between 08-11 and 08-12**
  (8 → 6 under the old type; `gamesdesign.co.uk` and `vonc.com` now have none while
  retaining 306 and 219 other work items, so not a site cleanup). No standing pruner
  exists in code. Unexplained. It matters beyond curiosity: **the detector's retraction
  cannot tell a fix from a deletion**, so if a shrink ever crosses a family boundary it
  will close a real finding on what looks like a positive observation. Recorded as a
  limitation in WII-015; the cheap guard is to store each family's row count at filing
  time and refuse to retract on a SHRUNK census rather than a converged one. Not built.

---

## TRAPS THIS LANE PAID FOR (do not re-pay them)

1. **`grep -Lq PATTERN` prints the files that DO match** — `-q` wins over `-L`, silently,
   and you get a confident inverted answer. It nearly became a council argument. Now in
   `LANDMINES.md`; use `grep -rLE`, and open one named file before quoting the count.
2. **A census of an ABSENCE is only as good as its spelling.** My first blind-spot
   extraction said 14; a strict regex had missed entries the file spells differently. The
   fix is to over-collect the "known" set so the bias runs toward *fewer* findings.
3. **`status='deferred'` is NOT undispatchable** — 316 live rows carry it, only 15 carry
   the empty-`handler_agent` half that makes the pair a lock. In `LANDMINES.md`.
4. **A distinct-key-set count is not a producer count** (it fires on four single-producer
   types), and Jaccard invents a producer for `page_canonical_collision`. If you touch the
   detector's clustering, re-read WII-015 before changing the threshold — it sits in a
   measured empty band, not a tuned one.
5. **Council rounds are cheap and legibility is most of the cost.** Round 1's gating
   objection was about a file I had already written but had not listed. List what you
   built; do not argue.
6. **Prove a CronJob image without exec** (a Completed pod cannot be exec'd and these
   images carry no build-provenance log line): image label
   `org.opencontainers.image.revision` = your commit, pod `imageID` digest = the pushed
   digest. Recipe in the RUNBOOK.

---

## COMMIT/REVIEW STATE

Five commits today: `ef1374426` (detector, `Council-Submitted`), `74ac4ed3a` (round-2
changes, `Council-Reviewed: fc082c4a-4b00-4835-8ffe-11a55e53f47a`, APPROVED),
`d7dfc77f4`, `5f0af2331` (docs), `a4befccd7` (the reply into `bugs_open/213`, the guard
census, the WII-015 limitation). The last one touches `platform/` and will therefore list
as un-reviewed in the `098` report: it is a **comment-only** contribution to a `_test.go`
file with no behaviour change, and a council round on a comment would spend the
architecture seat's signal on nothing. Stated here rather than left for someone to
wonder about.

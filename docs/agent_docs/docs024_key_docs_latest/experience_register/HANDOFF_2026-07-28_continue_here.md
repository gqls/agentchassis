# HANDOFF — experience register, 2026-07-28 (end of a long session)

**Read this first, then `NOTES_experience_register.md` from the 2026-07-27 entries onward.**
`SUMMARY_2026-07-28b` is the plain-prose read-out for the owner. Owner rulings are in
`PLAN_2026-07-24` §2 — **do not relitigate them.**

---

## 1. What this is, in one paragraph

A library of small reusable user-experience contracts — what a component or journey must DO,
specific enough to be machine-checked — held once and forked per site. It exists because a promise
nobody wrote down cannot be checked: if no record says this card should lead somewhere real, a card
leading nowhere is not detectably wrong.

## 2. State — everything below is LIVE and verified unless marked otherwise

| piece | state |
|---|---|
| Schema (`experience_patterns`, `site_experiences`, `experience_invariants`) | **LIVE**, migrations 218/230/239 applied + recorded |
| 9 entries + 9 travelling docs | **LIVE**, all via the validating write path, all `draft` |
| Write path `write_experience_pattern` + writer agent (mig 238) | **LIVE** |
| Bind path `bind_site_experience` | **DEPLOYED v1.0.1194**, never called by a workflow |
| Consumer `verify_site_experience` | **DEPLOYED v1.0.1194**, never called by a workflow |
| Approval council `experience-approval-council` (migs 259–262) | **LIVE and has ruled once** |
| **Applying a verdict to an entry's status** | **DOES NOT EXIST** — see §4.1 |

Pod-grep on v1.0.1194 (positive + negative control): `bind_site_experience` 12,
`verify_site_experience` 16, `requires Tier 4` 1, `deferred:` 1, `experience-nonsense-xyz` 0.

**Nothing has run through the real orchestration path except the write path and the council.**
Bind and verify have only been exercised **locally**, against `git archive HEAD` plus my files.

## 3. The two findings worth carrying out of this workstream

**3.1 The harness gap is measured and ranked.** 38 deferrals across the nine entries, grouped by
the capability that blocks them:

| blocked on | clauses | entries |
|---|---|---|
| **attribute assertion** | **13** | **9 of 9** |
| waits + retries (300 ms vs a measured 8–23 s) | 8 | 2 |
| cross-page status · empty-region | 3 · 3 | 3 · 2 |
| zero-count · threshold · focus · ordering | 2 each | — |

Attribute assertion is the **anti-dead-control rule** — so *we cannot check the rule the register
exists to enforce*, and one capability unblocks a third of everything. The browser-runner half
overlaps `gauntlet_dead_cta` P5: **coordinate, do not fork.**

**3.2 IMPLEMENTED is not SATISFIABLE.** The capability table answers whether a check *type* exists.
Whether a check can *succeed* depends on how the page is built — and **write-time validation sees
the template, never the page**. `asset_loads` is implemented and can only ever fail for a component
whose loader is an external bundle. Only a real run catches that class, which is why `dry_run` is a
first-class mode, not a debugging aid.

## 4. What to do next, in order

### 4.1 Apply-a-verdict action — the gate is still shut
The council **records** a verdict; nothing can write `status='approved'`. So every entry is still
`draft` ⇒ every fork is `proposed` ⇒ nothing reaches `verified`/`proven` except via `dry_run`.

Build `apply_experience_verdict`: read the latest `council_report` for the entry, and promote to
`approved` **only if** decision is `approved` **and** `unreadable == 0` **and** the entry's
`updated_at` is unchanged since submission. That last condition is the known gap stated in
migration 259's header — the council reviews the entry as it stood at submission.
Migration 230's CHECK already makes a zero-executable-check approval unrepresentable, so the DB
backs the action rather than trusting it.

### 4.2 Wire bind + verify
A workflow naming `bind_site_experience` then `verify_site_experience`. **Migration 238 is the
precedent** for seeding one, and its guard is the model: it fails if a later edit routes around the
paired step. Guardian's objection on corr `2e71f640` asked that this come back as an explicit
`config_change` naming the owning pipeline — honour that.

Bind `feed-driven-teaser-list` to vonc.com (`site_id 9ec3b9ee-5b08-461b-b4f8-9e1e03579c74`, page
`/provocations/index.html`). Real selectors, read off the served page:

```
list_section    [data-component="provocations-archive-list"]
list_container  .provocations-archive__list
item_template   .provocations-archive__item
feed_path       /data/provocations.json      <- NOT /provocations/…, that 404s
```

Expected result today (proven locally): **1 executable check passes, 4 held back with reasons,
verdict `verified`.** Thin, and the system says so.

### 4.3 Act on the council's REVISE
Corr `ec91c7e4-1b2c-4329-be19-4231cdfa553b`. Highest-value first:
- `requires_invariant` is **empty** while the entry restates `no-inert-control` three times —
  reference it instead.
- "no activation handler" (a third of the `must_not`) has **no check at all**, not even deferred.
- the degraded-state clause, "stay hidden", and the hide-the-slot clause have no check and **no
  deferral note** — silent gaps, not declared limitations.
- `feed_loads` is in `criteria_template` **and** `deferred_checks` — resolve the contradiction.

Then resubmit: `./260_TRIGGER_experience_approval_v1.sh feed-driven-teaser-list`.

### 4.4 Then
Harness work in the §3.1 order (attribute assertion first); then the planner side
(`experience_brief` in `plan_site`, bindings after `sync_pages`, the reconcile guard); then the
remaining harvest (`hub-spoke-index` — which is also what `features_open/001` needs, see
`COORDINATION_2026-07-28_packaged_topic_features.md`).

## 5. Landmines — each cost real time

- **Never test a shape against a fixture you wrote.** I invented the `contract` shape; all nine real
  entries disagreed; every test passed because the fixture was mine and *called*
  `validHarvestedEntry()`. **A name is not provenance.** Fixtures are the real files now.
- **A guard whose two sides are both yours proves nothing.** Migration 259's guard checked my list
  against my own list and passed while the consumer contract was broken. Check against the
  *consumer*, and against an independent witness where one exists.
- **`run-migrations.sh --apply` applies EVERY pending file**, not yours — 20 on 07-27, 19 other
  threads', several parked deliberately. Apply one with `psql -f`, then `--record-only`.
- **Every migration needs a `DO $$ … RAISE EXCEPTION $$` guard inside the transaction**, and
  **prove it by running it standalone before the migration, when it should fail.**
- **A council run has THREE outcomes.** The third is a `FAILED` row from the 4-hour stale-step
  reaper with no verdict in it. Poll with `error` in the SELECT; never infer your run's state from
  the queue depth (that cost a false "still queued" report 14 h after the run died).
- **Check `unreadable` before reading any verdict.** A filtered seat counts as abstained; an
  unreadable one is an opinion you were owed and lost.
- **`output_tokens == max_tokens` means CUT.** Raising the ceiling postpones it; bound the *ask*.
- **`grep -c` prints `0` and exits 1** — `|| echo 0` appends a second zero and breaks the compare.
- **Five hand-maintained lists went stale this week**, all mine. Derive from the real thing, or make
  the build fail when someone adds something and does not classify it.

## 6. Commands

```bash
# seed / re-seed entries (precheck greps the pod in BOTH directions, refuses the wrong binary)
./240_TRIGGER_seed_harvested_entries_v1.sh [CC-001]

# put one entry through the approval council
./260_TRIGGER_experience_approval_v1.sh <pattern-name>

# read a verdict — metadata FIRST, for unreadable
SELECT metadata::text FROM diagnosis_artifacts WHERE correlation_id='<corr>' AND kind='council_report';
SELECT body        FROM diagnosis_artifacts WHERE correlation_id='<corr>' AND kind='council_report';

# register state
SELECT name, kind, status, executable_checks, jsonb_array_length(deferred_checks) AS deferred
FROM experience_patterns ORDER BY name;

# entries vs travelling docs — a gap means a write landed and its doc did not
SELECT (SELECT count(*) FROM experience_patterns) AS entries,
       (SELECT count(*) FROM doc_plans WHERE subject_type='experience-pattern' AND is_current) AS docs;
```

## 7. Branch note

Another session moved the working tree to `087_towards_multiple_domains` mid-session. Verified:
`087` branched from `086_experience_loop` at `b82b3d8b4`, **after** every commit in this workstream,
so the whole chain is reachable. No reset was needed or made. Re-check `git branch --show-current`
before assuming.

# HANDOFF — vigilant designer + offer analyser (2026-08-25b)

**COLD-START = this file + `bugs_open/395` §8 + register `CLM-024` and `WII-033` + `features_open/030` §10 + `features_open/034`.**
**This supersedes `HANDOFF_2026-08-25_continue_here.md`**, which is still correct on history and out
of date on §1 — the consumer it asked for is **built, approved and committed**.

> **Re-run every liveness claim here before acting.** This branch takes hundreds of commits a day.
> Verify against `git archive <resolved-sha>` — never the working tree (measured again today: another
> lane's untracked `audit_tone_pending_proposal_bound_test.go` appeared mid-test-run and failed the
> package; `component_selector.go` is unformatted; and `platform/livespec` fails at **plain HEAD**,
> see §5) and never the moving name `HEAD`.

## The one-line state

> **The loop is closed in code and NOT YET in the world. Gate 1c reads an item's own acceptance
> predicate before `complete` — and it RECORDS rather than refuses, because no live negative control
> exists. Go, so inert until a roll. The next real step is one row: a completion carrying
> `outcome='permitted'`.**

## What is DONE — do not re-take any of this

| | state |
|---|---|
| **completion gate 1c** (`complete_work_item_acceptance_predicate.go`) | **COMMITTED `69479bcf6`**, wired into `verifyBeforeComplete` between gates 1b and 2. **Go ⇒ INERT until a chassis roll** |
| council | **APPROVED round 1**, corr `064841bd-58fc-46a1-a77d-6b0a6309d0ba`, 14 seats, 5 advisory, none high. All four actionable objections acted on in `74829da90` |
| the stamp/strip pair (`storedPredicate` / `predicateForEvaluation`) | live in the emit file, pinned by `TestStampAndStripAreInverses` |
| `loadAcceptancePredicateSubjects` | now takes `*sql.DB`, serving BOTH ends of a predicate's life from one query |
| the claim-timeout lockstep | extended: a **refusing** gate-1c entry must be excluded, a **recording** one must not. Promotion is now a build failure |
| `ACCEPTANCE_PREDICATE_NOT_EVALUABLE` | declared `human-evidence` in `finding_code_registry.json` (caught by the source-side scan, in the same commit) |
| register `WII-033`, index row, `bugs_open/395` §8, `016b` §9 correction | all written |
| `LANDMINES` ×2, `WRONG_CALLS` ×1 | written; `landmines-verify-dispatch.sh` run (5 dispatched) |
| `architecture_review/RFC_055` | filed — the architecture seat's accumulation signal, with a recommended **partial** |
| `SUMMARY_2026-08-25_something_reads_the_tests.md` | written (new file, tenth in the series) |

## What the next session should do

### 1. AFTER THE NEXT ROLL: get the negative control. This is the work.

Everything else is downstream of one row. `bugs_open/395` §6 says a fix is not proven until an item
whose predicate is **satisfied** completes green through this gate, and there is no such row anywhere.

```sql
SELECT s.domain, wi.item_type, wi.completed_at,
       wi.result->'_verification'->'acceptance_predicate'->>'outcome'  AS outcome,
       wi.result->'_verification'->'acceptance_predicate'->>'verdict'  AS verdict_now
  FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
 WHERE wi.result->'_verification' ? 'acceptance_predicate'
 ORDER BY wi.completed_at DESC;
```

⚠ **`outcome='permitted'` is the row.** A census of nothing but `recorded_only` means the gate has
still only ever seen failures — which is indistinguishable from one that refuses everything, and is
exactly why it does not refuse today.

**Manufacture it if it does not arrive on its own**, which is what `HANDOFF_2026-08-25` already
advised and is now more precisely actionable: hand-fix one page's `meta_description` so a live
predicate's condition holds, then re-fire the analyser and let the item close. The satisfying string
for the `index` predicate is in `complete_work_item_acceptance_predicate_test.go` as
`satisfyingIndexMeta`.

⚠ **FIRST verify the gate is running at all.** Go changes are inert until a roll:

```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor 69479bcf6 <that sha>   # exit 0 ⇒ gate 1c is live
```

⚠ **An empty `_verification.acceptance_predicate` column means "never switched on", NOT "nothing
refutes".** Do not read a clean census as a clean estate before you have run the ancestry check.

### 2. Then, and only then, consider promoting to `predicateRefuses`

`PromotionOwes` on the roster entry states the debt; the build enforces half of it. In order:

1. the negative control from §1;
2. add `content_rewrite` to `livespec.ClaimedItemTimeoutExclusions` **AND ship the migration
   amending the live `scheduled_tasks.pre_query`** — the declaration alone changes nothing in
   production, and without it the timeout sweep completes items straight past the gate
   (`bugs_closed/317` reintroduced BY adding a guard). ⚠ **`WII-032` (the `bugs_open/375` lane, same
   day) needs the SAME migration for `required_fields_missing` — coordinate, one migration could
   serve both.**
3. re-visit the `inapplicable` arm: it records and never blocks, which is a stated asymmetry with
   RFC_017 and only defensible while the outcome is record-only.

⚠ **And the 2.3% tail:** `[MEASURED 2026-08-25]` 38 of 1,638 `content_rewrite` completions carry
`handled_by` NULL, written by neither completion action. Harmless today (none ever carried a
predicate); it is what a refusal would silently miss.

### 3. The rest of the v2 batch — write `602` for (a)+(b)+(c), config-only

**Unchanged from the last handoff and still not done.** v2(d) shipped alone because it needed Go;
these do not. Traps in `features_open/030` §10:
- **v2(a)** bounded head-of-hero excerpt per page — ⚠ GROWS the surface; re-run the truncation check
  on webdesign.co.uk afterwards. It also **widens what a predicate can address**, since body-text
  shapes are excluded today precisely because the surface carries no content.
- **v2(b)** attribution in the `why` clauses — partly done by 537; re-read the live prompt first.
- **v2(c)** `primary_model` in the degraded arm's field list — LATENT; must not be the reason to open
  the batch, and do not fix it by letting the model *infer* one.

### 4. `features_open/034` — claims audit over `site_specs` prose

Owner-approved 2026-08-14, still not designed. Gate 1c checks whether a page matches what we said
about it; 034 asks whether what we said was true.

### 5. Coverage, and the estate figure that is NOT ours to quote

`[MEASURED 2026-08-24]` five sites carry `offer_ordering` out of 28 live sites. ⚠ **Do not carry a
site count forward from any document — re-run it.**

## Watch-outs this lane has now paid for (new ones first)

- **⚠ A HANDOFF HANDS YOU A REASON AS WELL AS A CONCLUSION, AND THE REASON IS THE PART TO RE-DERIVE.**
  `bugs_open/395` §4 and the last handoff both said gate 1c belongs beside `noChangeGates` because
  *"`VerifyTarget` carries the SPEC, not the RESULT"*. **That is gate 1b's argument and does not
  transfer** — 1c needs the spec (which `VerifyTarget` carries) and the current page row (which a
  verifier can read). The conclusion held; the reasoning was borrowed from a gate answering a
  different question, and I had already propagated it into `016b` §9. Corrected in three places.
- **⚠ A STORED `acceptance_predicate` CANNOT BE FED TO `EvaluateAcceptancePredicate`.** The emit gate
  stamps `verdict_at_emission`/`evidence_at_emission` AFTER evaluating; the evaluator enforces a
  closed key set. Every live predicate returns `inapplicable` — a legitimate verdict whose message
  names a KEY, so it reads as a fault in the model's output. Use `predicateForEvaluation(stored)`,
  **never** the stored map, and never widen the key set (`bugs_closed/335`). `LANDMINES.md`.
- **⚠ `pattern-check.py` READS YOUR PROSE.** `logged-model-output` matches nine ordinary English
  words over six raw lines from a log sink, string literals and comments included. The comment you
  write to explain the false positive is INSIDE the window and re-fires it. And
  **`python3 scripts/pattern-check.py` with no arguments reads `git diff --cached`** — unstaged means
  it scans zero files and prints a clean result. Confirm the denominator.
- **⚠ ATTACHING A LOOKUP IS NOT ATTACHING THE RIGHT ONE.** Answering the `prior_art_librarian` seat, I
  wrote "grep says the emit gate and this file", ran it, and got **seven**. The claim survived only
  via a writer/reader split the grep cannot make. Put the breakdown next to the command, or the next
  reader runs it and concludes your sentence is false.
- **⚠ `platform/livespec` FAILS AT PLAIN HEAD** (2026-08-25):
  `TestNoNewMigrationFileReadersOutsideTheAllowList` — `work_item_owned_page_door_test.go` reads a
  path under `sql_for_agents` and is not allow-listed. **Verified against unmodified HEAD**, so it is
  not yours; it belongs to the owned-page-door lane. Do not let it stop you reading a
  `verify-head-builds.sh` result — scope the run to `./platform/orchestration/...`.
- **⚠ `verifyBeforeComplete` NOW RETURNS FOUR VALUES**, and `verification != nil` NO LONGER MEANS
  BLOCKED. A completion that proceeds can carry a verdict (a satisfied predicate, or a refuting one
  on a recording type). Stamp the payload, THEN read `mayComplete`.
- **⚠ the three `wont_fix` items are the OWNED-PAGE DOOR, not a regression here** — see
  `CONTRIB_2026-08-25_write_audit_findings_bypasses_the_new_owned_page_door.md` in this directory.

### Carried forward, still true

- **⚠ CONFIRMING THAT THE PROMOTER YOU THOUGHT OF IS OFF IS NOT A SAFETY ARGUMENT.** The cheap check
  is the inverse query — `SELECT name, enabled FROM scheduled_tasks WHERE enabled` — because a
  `WHERE` clause naming your own suspicion cannot discover a second cause.
- **⚠ `pages.in_header` IS NOT THE RENDERED NAV** — 13 rows vs 7 served destinations. Nav is out of
  the predicate vocabulary for this reason. `rendered_header` is not the escape route.
- **⚠ A ROLL KILLS AN IN-FLIGHT COUNCIL**, and a lone casualty is the EXPECTED shape. The pod-age
  comparison is the check; a peer census is not a second opinion.
- **⚠ MIGRATION NUMBER COLLISION: there are TWO 601s.** Resolve by slug, never by number.
- **⚠ `site_work_items` has no `audit_source` COLUMN** — it is `spec->>'audit_source'`, and the column
  form ERRORS rather than returning zero.
- **⚠ `orchestration_states` has no `agent_type`** — it is `owner_agent_type`.
- **⚠ psql prints UTC, your shell prints BST** — always toward alarm. Make the DATABASE subtract.
- **⚠ `run-migrations.sh --record-only` REFUSES a `_HOLD` file.** Record by hand INSERT.

## Residuals, stated plainly

1. **No live negative control.** §1. Until `outcome='permitted'` appears on a real row, the refusal
   arm is proven by units only — **CLM-023's residual in a THIRD place**, and stated as such in the
   roster entry rather than buried.
2. **Nothing is prevented yet.** Record-only, by design.
3. **The blast radius (`395` §5) is still a plan.** `[MEASURED 2026-08-25]` exactly ONE of 1,638
   `content_rewrite` completions all-history carries a predicate. It grows only as the producer runs.
4. **Why the handler produces content failing its own predicate is untouched.** The council's
   `constitution` seat flagged it; gate 1c is detection, not that fix.
5. **The truncation asymmetry, unmeasured:** the model authors predicates against a meta description
   truncated at **160 chars**; the evaluator reads the FULL column. No live instance yet.
6. **Four hand-wired gates on one function** — `architecture_review/RFC_055`, awaiting an owner call.
   Its recommended partial is small: extract only the "can this gate refuse" registry.

## Who owns what nearby

**`bugs_open/375`'s lane is working the SAME seam** — `WII-030`/`031`/`032`, all committed today, and
`WII-032`'s step (1) is the same claim-timeout migration §2 needs. **Coordinate before writing it.**
The **leopardess lane** still holds five of this lane's findings at `needs_human_review` — coordinate
before firing B4 at that site, and note that filing findings there dispatches handlers at work they
are holding. The **owned-page-door lane** owns the `platform/livespec` HEAD failure in §5.
`bugs_open/333` belongs to the 301 lane. The **`bugfix_308_cta_destination_provenance`** lane has
routed the undecidable-CTA question to this agent (`CONTRIB_2026-08-24`, owner ruling in `RFC_047`
§10); its own read is *"after your v2 batch"* and nothing is blocked on us.

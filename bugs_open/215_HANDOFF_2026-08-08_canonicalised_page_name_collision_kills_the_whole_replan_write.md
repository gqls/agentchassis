# 215 — two emitted pages canonicalise to ONE name and the whole replan write dies on the unique index; which plans fail is decided by LLM emission variance

**Filed 2026-08-08** by the brochure_component_library lane, from a live failure
the same morning (fundamentallyai replan, corr
`1cb17b11-fd10-4d52-836c-36e2fa246ff6`):

```
step write_site_plan failed: ... insert site_plan_pages for "tool-llm-cost-calculator":
ERROR: duplicate key value violates unique constraint "idx_site_plan_pages_name" (SQLSTATE 23505)
```

**Verification route, declared per the 2026-07-31 owner ruling:** not through the
090 loop; substituted first-hand evidence from the failing run's own
`collected_data` plus the code path read at HEAD — the raw emission contains
both colliding names (query below), the canonicalisation site and the
unguarded insert are cited by line, and the day-before run on the SAME site
shows the differential (one variant emitted → no collision).

## Mechanism

1. The planner LLM may emit a page under its stem name AND its canonical name
   in one plan. This run: `llm-cost-calculator` (page_type `tool`, 3 sections)
   **and** `tool-llm-cost-calculator` (page_type `tool`, 0 sections — a stub),
   plus the same pattern in waiting: `tools` (2 sections) and `tool-tools`
   (0 sections). Read them from the failed run:
   ```sql
   SELECT p->>'name', p->>'page_type', jsonb_array_length(p->'sections')
   FROM (SELECT jsonb_array_elements(collected_data->'llm_plan'->'result'->'pages') p
         FROM orchestration_states
         WHERE correlation_id='1cb17b11-fd10-4d52-836c-36e2fa246ff6') x
   WHERE p->>'name' LIKE '%llm-cost-calculator%' OR p->>'name' LIKE '%tools%';
   ```
2. `WriteSitePlanAction` canonicalises each page independently
   (`datahelpers.CanonicalisePage`, called at
   `write_site_plan_action.go:277`): a `tool`-typed `llm-cost-calculator`
   becomes `tool-llm-cost-calculator`. **Nothing dedups the page list after
   canonicalisation**, so two entries now carry one name.
3. The per-page insert (`write_site_plan_action.go:379`) hits
   `idx_site_plan_pages_name` on the second one and the action errors. The
   write is transactional — verified in the same incident: no new `site_plans`
   row, previous plan still `is_current`, zero orphan rows — so there is **no
   data damage**, but the entire replan is lost.
4. Whether any given replan fails is therefore decided by **whether the LLM
   happened to emit both spellings that run**. The previous morning's replan
   of the same site emitted only the stem variant and succeeded; this one
   emitted both and died. A retry may pass. That makes this a low-frequency,
   zero-signature reliability hole in every `build-site-planner` run
   fleet-wide — the failure names neither the canonicaliser nor the LLM, and
   an operator's likeliest wrong conclusion is the one the error suggests
   (a stale unique index or concurrent write).

## Why the stub pages exist at all

The prompt's context shows the site's existing pages (which include the
canonical `tool-*` names) while the planning instructions talk about tool
pages by stem. The model, asked to enumerate every page (rule 17's
every-page requirement arrived the same morning — seed 333 — and plausibly
raised the odds of exhaustive enumeration), lists both spellings. Emission
variance, not a prompt defect as such — the write path must be safe against
it regardless.

## Fix candidates, ordered by what closes the door

1. **Dedup after canonicalisation, inside `WriteSitePlanAction`** (the only
   door): group the validated+canonicalised page list by final name; merge
   duplicates keeping the entry with sections (a stub loses to a composed
   entry; two composed entries = keep first, log loudly with both section
   lists). ~20 lines before the insert loop; makes the collision
   unrepresentable regardless of what the LLM emits.
2. **Planner prompt: name pages canonically** (tell it tool pages are always
   `tool-<stem>` and never to emit both). Reduces the odds; closes nothing —
   the write path would still be one emission away from dying.
3. Retry-on-23505 at the orchestration layer — treats the symptom, hides the
   defect, explicitly NOT recommended.

## How to verify a fix

Feed `WriteSitePlanAction` a page list containing a stem + its canonical
variant (unit: the failed run's raw pages array is a ready fixture) — the
write must succeed with ONE `site_plan_pages` row for the name, the composed
sections must win over the stub, and a log line must name the merge. Then
re-run the census: no plan write failure with `SQLSTATE 23505` on
`idx_site_plan_pages_name` in `orchestration_states.error` after the fix
ships (query `error LIKE '%idx_site_plan_pages_name%'`).

## Relations

Found while executing `bugs_open/151` candidate 1 Slice B (RFC_016 §3a option
(a) compliance replan — the failure cost that observation run);
`bugs_open/204`/`214` (the same wire's other positional/naming traps);
`datahelpers.CanonicalisePage` (the canonicaliser itself is correct — the gap
is the absent post-canonicalisation dedup).

---

## CORRECTED + STRENGTHENED 2026-08-09 — the collision is proven, the PAIRING was read from the wrong key, and the same defect has a second, quieter damage mode

**What stands, re-verified at HEAD today:**

- The error is a quoted fact: `insert site_plan_pages for "tool-llm-cost-calculator":
  duplicate key ... idx_site_plan_pages_name`. The insert names `r.Name`, so **two
  rows in `planRows` carried that canonical name** — that much is certain.
- **There is no dedup anywhere on the path**, re-read at HEAD 2026-08-09:
  the canonicalise loop appends unconditionally
  (`write_site_plan_action.go:274-315`) and the insert loop executes one statement
  per row (`:355-381`). `idx_site_plan_pages_name` is `UNIQUE(plan_id, name)`.
  So a post-canonicalisation duplicate ALWAYS aborts the write.

> **CORRECTED: the claim that the colliding pair was the emitted
> `llm-cost-calculator` + a `tool-llm-cost-calculator` stub is an INFERENCE, not a
> measurement — and it was read from the wrong stage.** I took it from
> `llm_plan.result` / `validate_plan`, but `WriteSitePlanAction` reads neither: it
> calls `extractPagesFromPlan`, which reads **`page_plan` then `site_plan`**
> (`site_db_actions.go:749-782`). I never inspected `site_plan` for that run, and
> the row has since expired (~24h; verified gone 2026-08-09), so **which two
> entries collided is now permanently [UNVERIFIABLE]** for this incident. This is
> the same error class as `WRONG_CALLS` 2026-08-08, committed one day after
> writing that entry — see the 08-09 entry there.
> **A reproduction must read `site_plan`, not `validate_plan`.**

**Second damage mode, measured today — the same dual-identity problem that does
NOT crash, and it reached production.** The 2026-08-07 replan of fundamentallyai
(plan `8ee5807b`) wrote page rows for canonical/stem twins of pages that were
already live under the other spelling. Three rows, all created 08-07 08:24:22,
all `planned` + `deployed_at IS NULL` + zero components — i.e. permanent 404s:

| phantom row (archived 08-08) | live twin, serving 200 |
|---|---|
| `tool-llm-cost-calculator` → `/tools/llm-cost-calculator/index.html` | `llm-cost-calculator` → `/tools/llm-cost-calculator.html` |
| `tool-tools` → `/tools/tools/index.html` | `tools` → `/tools.html` |
| `ai-readiness-checker-guide` → `/blog/ai-readiness-checker-guide.html` | `tool-ai-readiness-checker-guide` → `/guides/…` |

Note the direction flips (phantom is the canonical form twice, the stem form
once) — the invariant is **two identities for one page**, not a fixed prefix.

They were found and hand-archived by the fundamentallyai sweep front on
2026-08-08 (`HANDOFF_2026-08-09_sweep_front_continue_here.md` §2b), which also
had to cancel four `needs_human_review` work items pointing at them. Worse,
while they existed they were valid internal-link targets — a `pages` row is
`active` from creation — which is the ammunition behind that front's own
linkability fix (`1c2e25c8f`): a served page linked to
`/platform-log/index.html` for 18 days while it 404'd.

**So the severity is higher than filed, and the cost is already paid twice:**
one lost replan (crash mode) and three phantom 404s plus four dangling work
items (quiet mode), from two consecutive replans of one site.

**Fix candidate 1 covers both modes and needs one addition:** dedup by canonical
name inside `WriteSitePlanAction` closes the crash; the quiet mode also needs the
plan's page identities reconciled against **realised pages under either
spelling** (a plan row whose canonical name differs from a live page's name but
resolves to the same page must not create a second identity). Verify the quiet
mode with the census:

```sql
SELECT s.domain, p.name, p.url, p.created_at
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE p.status NOT IN ('deleted','archived')
  AND p.deployed_at IS NULL AND COALESCE(p.build_status,'')='planned'
ORDER BY 1,2;   -- fleet-wide phantom candidates; HTTP-test before acting
```

---

## 2026-08-09 — CRASH MODE FIXED (committed `14b1cff28`, inert until the chassis rolls). Three corrections to this file, one of them to its own verification step

Fix candidate 1 built as `dedupePlanPageRows` in `write_site_plan_action.go`,
called between canonicalisation and the transaction. Council-Submitted
`8ab18991-ee83-4048-8965-4f7990baa188`. **This bug stays OPEN**: the fix is
inert until a chassis roll, and the quiet mode below is deliberately not fixed.

**Correction 1 — there are TWO crash doors, not one.** This file names only
`idx_site_plan_pages_name`. Read from the live schema 2026-08-09,
`site_plan_sections` carries `idx_site_plan_sections_key` UNIQUE
`(plan_id, page_name, ordering)` — so two composed pages sharing a canonical
name would ALSO collide there. It has never been the observed error only
because the pages insert runs first and aborts the transaction. The same dedup
closes both. (Checked and negative: there is **no** unique index on `url`, so a
URL collision is not a third door.)

**Correction 2 — there are THREE collision families, and this file names the
least likely one.** Read from `datahelpers/page_canonical.go` at HEAD rather
than inferred from the incident:

| family | two spellings that collide | canonical result |
|---|---|---|
| prefix collapse (`:153-184`) | slug `llm-cost-calculator` / `tool-llm-cost-calculator`, role `tool` | `tool-llm-cost-calculator` |
| **homepage collapse** (`:117-127`) | role `index` / slug `home` under content-landing-empty | `index` |
| section-index (`:136-149`) | slug `guides` / `guides-index` under any section-index role | `guides-index` |

`guide-` and `game-` behave exactly as `tool-`. **The homepage family is the
likeliest of the three in ordinary emission** — a planner listing both a
homepage and a "home" page is a far more ordinary slip than emitting a
tool page twice — so the frequency implied by "the LLM happened to emit both
spellings" is understated in the original filing. All three are pinned by
`TestCanonicalisePage_CollapseFamilies`, which exists so that if the
canonicaliser ever stops collapsing, the suite reports the dedup as dead code
rather than leaving it silently inert.

**Correction 3 — THIS FILE'S OWN "how to verify a fix" STEP CANNOT COME OUT
OTHERWISE, and I nearly shipped it as proof.** The step above says: after the
fix, re-run `error LIKE '%idx_site_plan_pages_name%'` over
`orchestration_states` and expect no failures. Measured today:

```
status      count  oldest      newest
COMPLETED    4822  2026-08-08  2026-08-09
FAILED         60  2026-08-08  2026-08-09      -- and 'failed rows older than 24h' = 0
```

4,935 rows, oldest overall `2026-07-13` — so the **table** is not 24h-limited,
but **failures are**. The census returns 0 today, and the 2026-08-08 incident
in this file's own header **is one of the rows it cannot see**. For an event
whose period is longer than the retention window, that query reads 0 before the
fix and 0 after it, whatever the truth. It is not a verification; it is a
formality that would have looked like one.

Replaced by two signals that can actually be non-zero: a
`duplicate_pages_merged` counter on the action's result (additive — grep across
non-test Go and a scan of active `agent_definitions` both return **zero**
consumers of any key of that map), and the merge log lines, which name both raw
spellings so a collision stays diagnosable after the run row expires. The real
proof of the fix is the unit suite plus mutation evidence (pass-through fails
6/7 tests; always-keep-first fails; unguarded backfill fails), not a census.

**Still OPEN — the quiet mode is NOT fixed, deliberately.** A plan row whose
canonical name differs from a live page's name but resolves to the same page
still creates a second identity (the three phantom 404s above). That
reconciliation belongs in the reconciler against realised pages under either
spelling — not in `WriteSitePlanAction`, which sees only one plan's emission and
has no knowledge of realised pages. Putting it here would also be exactly the
"shared seam smuggled inside a bug patch" the 2026-07-28 ruling forbids. The
phantom census in the section above is still the right check for it: **20**
candidates fleet-wide today.

**Also unchanged, and adjacent enough to trip someone:** page-scope imagery
`scope_ref` keys off the **raw** LLM page name, not the canonical one
(`flattenImageryBlock`), so the two name-spaces coexist on this path. That is
`bugs_open/214`'s territory; this fix neither improves nor worsens it.

---

## 2026-08-09 (evening) — CRASH MODE IS NOW LIVE-VERIFIED on chassis **v1.0.1276**

The fix rode a fleet roll. Verified at the artefact on **both replicas**
(`agent-chassis-767d7f5674-5sxdc`, `-sfct5`), not at git and not at the tag:

```
POSITIVE  "duplicate page collapsed after canonicalisation"          -> 1  (both pods)
POSITIVE  "two composed pages canonicalise to one name"              -> 1
POSITIVE  "plan contained pages that canonicalise to a shared name"  -> 1
POSITIVE  "duplicate_pages_merged"                                   -> 1
NEGATIVE  "collapsed after canonicalization"  (US spelling)          -> 0  (both pods)
```

The negative control is the load-bearing half: the same grep, same exec, same
binary, differing only in one letter, returns 0 — so a positive proves the
spelling, not merely that the pipeline can find *something*. (This change
removed no strings, so a removed-string control was not available; a
plausible-but-absent spelling is the substitute.)

**Status: the CRASH mode is fixed and live. The QUIET mode is not, and this bug
stays OPEN for it** — a plan row and a live page still hold two identities for
one page. So a replan of an affected site still generates phantom 404s and still
costs another front cleanup; it just no longer loses the whole plan.

**Not yet observed in production, and worth saying plainly:** no plan write has
been through the new path since the roll, so `duplicate_pages_merged` has never
been non-zero in the wild. The fix is proven by unit tests, mutation evidence
and a pod-grep — *not* by a live merge. The first real signal will be that
counter or the merge log lines; do not read their absence as either success or
failure until a replan has actually run (and per the 2026-08-09 landmine, do
not reach for an error census over `orchestration_states` to decide it).

## 2026-08-10 — OWNER RULING: richer-wins RATIFIED, conditional on durable observability (now shipped)

The outstanding policy question ("how much silent loss is acceptable when two COMPOSED pages
collide") was ruled 2026-08-10: **richer-wins stands** — it discards strictly less than keep-first
in every case, the observed collision shape (composed + stub) takes the lossless branch, and
failing the write would restore the very whole-replan loss this bug is about. **Condition:** the
lossy branch must be durably observable. Its only trace was a chassis Warn, and an active chassis
pod retains **under one second** of log (bugs_open/136 §11), so "has richer-wins ever actually
dropped authored content" was unanswerable.

Shipped with the Slice B resubmission (corr `a06ff850`, same commit as this note):
`dedupePlanPageRows` now returns the lossy merges alongside its existing results, and
`WriteSitePlanAction` persists each as **`PLAN_PAGE_MERGE_LOSSY`** on `agent_error_log`, carrying
both raw names and both FULL section lists — the discarded composition is reconstructable from the
row. Mutation-tested (suppressing the detail fails `TestDedupePlanPageRows_LossyMergeDetailReturned`).

**Ratification terms: richer-wins, durably recorded, revisit if `SELECT count(*) FROM
agent_error_log WHERE error_code='PLAN_PAGE_MERGE_LOSSY'` is ever non-zero.** Ruling context and
the re-look that reframed the condition:
`docs/agent_docs/docs024_key_docs_latest/brochure_component_library/DECISIONS_2026-08-10_owner_rulings_after_relook.md`.

> **THE RICHER-WINS REVISIT TRIGGER TRIPPED, 2026-08-11 10:21:48.** The standing owner
> item ("`PLAN_PAGE_MERGE_LOSSY` count non-zero ⇒ look at richer-wins again") is now live:
> the census replan (corr `e74974b3`) recorded TWO rows — `automation-savings-estimator-guide`
> and `model-approach-selector-guide` each canonicalise-collided with their `tool-`-prefixed
> twin, and all four page rows are active+deployed, so both merges discarded a real page's
> entry (both full section lists are in the error rows' context). Richer-wins did its job
> (the replan survived), but the case it was ratified on ("stub loses to composed entry")
> is not this case: these are composed-vs-composed. Owner decision requested; recorded in
> the lane README 2026-08-11. The duplicate-page family itself (one page under two names,
> both live) is the underlying condition — neither this bug's dedup nor the merge rule
> resolves WHICH name should own the page.

---

## 2026-08-11 — QUIET MODE: fix built. Two coupled halves, because "it belongs in the reconciler" is right about the DECISION and insufficient on its own

Taken up as the remaining scope, per the owner ruling of 2026-08-11 §1 ("the
phantom-mode fix stays in `bugs_open/215`'s remaining scope on its own merits").
Council submission `3cd9fd92-da62-46b9-9799-cb439574eff2`.

### The correction this file needs

This file says the reconciliation "belongs in the reconciler against realised
pages under either spelling — not in `WriteSitePlanAction`". The first half is
right and I have built it there. **The second half is incomplete, and a
reconciler-only fix would have been INERT for exactly the pages this bug is
about.** Both canonicalisation surfaces re-derive every page's identity
unconditionally (`write_site_plan_action.go`, `site_db_actions.go`), and
`CanonicalisePage` **cannot express a legacy identity** — a `tool`-typed page
always comes back `tool-<bare>` at the role's default hub.

> **[MEASURED] 2026-08-11, live DB: 71 live SHIPPED rows fleet-wide are not fixed
> points of `CanonicalisePage`.** For every one of them, a reconciler that
> correctly recognises the twin still hands the writer a page that is re-derived,
> conflicts with nothing on `(site_id, name)`, and is INSERTed as a second row —
> the phantom, re-minted by the very pass that spotted it.

So: the reconciler decides, and the writers must be told to stop overruling it.
That second half is the one genuinely new authority here, and it is the edit to
distrust — it is opt-in, default OFF, and named in the submission as the thing
the council should attack hardest.

### What was built

| piece | where | default |
|---|---|---|
| `PagePathKey` / `PageItemStem` / `PageCanonicalNameForRow` | `datahelpers/page_identity.go` (new) | n/a — extraction |
| layer 1: normalised path key | `reconcilePlanWithRealised` | **ON** |
| layer 2: predicted canonical identity | `reconcilePlanWithRealised` | **ON** |
| layer 3: stem twin, both directions | `reconcilePlanWithRealised` | **OFF**, dark-launched |
| writer honour guard | both canonicalisation surfaces, one shared reader | **OFF** |
| `reconciled_from` imagery alias | `buildCanonicalPageNameMap` | ON (one line) |

All four match routes (including the pre-existing exact-URL Pass B) now go
through **one extracted arm**, so the `bugs_open/050` empty-page routing and the
`bugs_open/151` fact-assignment carry cannot drift between them. Snap, never
drop — dropping is what Pass C2 does, and it discards the plan-time fact
assignments with the entry.

Guards, all unconditional even when a layer is on: refuse a key two realised
pages claim; refuse when the plan already carries the realised spelling; refuse a
never-shipped stem twin; stem requires **exactly one** side prefixed.

### Two things this fix deliberately does NOT do

1. **It does not resolve the both-deployed pairs, and it must not.** When both
   spellings are realised AND both are in the plan, the layers REFUSE. Snapping
   would hand the writer two entries with one name, and richer-wins would then
   resolve the pair by evicting a live page. Which name owns the page is a
   remediation decision — which is precisely the question the 08-11 note above
   raises about the two composed-vs-composed lossy merges. **My fix does not
   answer it; the runbook below scopes it for the owner.**
2. **It does not touch the archived-page rebuild.** See below — separate defect,
   filed separately.

### Measurements, all 2026-08-11, all able to have come out otherwise

- **Would-merge survey** (current plans joined to realised pages, names differing):
  normalised path matches **3** pairs, stem matches **11**; a human read **0** of
  them as genuinely different pages. Confined to fundamentallyai and robot-hands.
- **Both-deployed twin pairs: 7, across 4 domains** — duplicate LIVE content, not
  just phantom 404s. All 14 URLs HTTP-tested 200, against a 404 control of 2697
  bytes (so the 200s are content, not the error page). Component counts differ per
  side (robot-hands 5/3/4 against 1 each), i.e. genuinely different builds.
- **Today's replan (corr `e74974b3`) minted NO new page rows.** Its twins collapsed
  in-plan instead — the two lossy merges in the note above. That is emission luck,
  not protection: the same plan one spelling different mints a phantom.

### Mutation evidence

Pass-through of the whole function fails 5 tests · deleting the path-key layer
fails its own test (and exposes the canonical layer catching the same fixture,
which the assert-on-layer catches) · removing the exactly-one-prefix guard fails
the `tool-pricing`/`guide-pricing` test · dropping the both-in-plan refusal fails
the robot-hands test · not stripping the forged marker fails its test · **removing
the writer guard from one surface leaves every unit test green and fails only
`TestIdentityPolicyReachesBothCanonicalisationSurfaces`**, which is why that test
exists.

### Two defects found while doing this, both recorded rather than folded in

1. **Archived pages are rebuilt and re-deployed by the work-item pipeline.**
   `ai-readiness-checker-guide` and `tool-llm-cost-calculator` were hand-archived
   on 08-08 with `deployed_at IS NULL` and zero components; on **2026-08-11 they
   acquired `deployed_at` stamps (10:34:21 and 11:13:25) and now serve HTTP 200**
   beside their live twins. So the sweep front's hand-archive is **not durable**
   against the refile loop, and remediation cannot be assumed to stick.
   `loadRealisedPages` (`reconcile_site_plan_action.go:458`) selects from `pages`
   with **no status predicate**; where the archived status should have gated the
   BUILD or the DEPLOY is not self-evident from a read, so it went to the
   diagnosis loop rather than into an assertion here — **090 run correlation
   `38099787-c7f9-46d4-b75e-3a1867fcaf41`**. Open work items sit on archived pages
   across **8 domains**, so this is a class, not a fundamentallyai quirk.

   > **CORRECTED, same day, before anyone relied on it: I called this a "distinct
   > defect" as though it were undiscovered. It was not.** The fundamentallyai
   > sweep front had written the mechanism into its own handoff at **12:47**,
   > hours before I wrote this section at 17:28 — it PREDICTED the replan would
   > re-plan all three archived pages, named `ai-readiness-checker-guide` as the
   > one that would auto-build and deploy, and flagged that it would need file
   > retraction this time. The chain (plan still names the page → reconcile emits
   > `needs_page` → build → deploy) is the documented regeneration trap, PLAN-017's
   > landmine, not a new one. **What caught it:** reading that handoff before
   > writing my coordination note into it, which is the step that should have come
   > first. The measurement stands and the 090 run is still worth its cost, but
   > the question it actually asks is narrower and should be read that way:
   > *should the build/deploy path refuse a page whose `status` is `archived`,
   > rather than relying on the plan never naming it?*
2. **The canonical layer must derive its key the way the write path does**
   (`firstNonEmpty(slug, name)`, not name alone). Caught by reading the actual
   `PLAN_PAGE_MERGE_LOSSY` rows rather than inferring from the names in them: the
   entry NAMED `tool-model-approach-selector-guide` canonicalised to the BARE
   name because its slug said so.

### Status

**Part 1 committed (`65c1984d0`)** — the shared keys and the policy helper, inert,
building and testing green against a clean HEAD tree. **Part 2 (the wiring) is
written, tested and NOT yet committed**: all three wiring files currently carry
another lane's uncommitted work in the same hunks, and committing them would
either ship that work under this message or leave HEAD unable to compile. Held
rather than swept. This bug stays **OPEN** until the wiring lands, the chassis
rolls, and a site opts in.

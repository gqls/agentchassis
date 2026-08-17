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

### 2026-08-11 (later) — COUNCIL: REVISE round 1, and the objection that was worth the round

Verdict REVISE on `56e13695`, gated by `editquality`. Resubmitted round 2 on the
same trail id. Two things worth carrying forward whatever the verdict:

**The design changed, and the seats were right.** `path_key` and `canonical_name`
shipped **default ON**, on the argument that they ask the same question the
exact-URL rename already asks unconditionally. The `guardian` and `architecture`
seats independently objected: that changes matching behaviour for every existing
caller fleet-wide on deploy, while the weaker stem layer got a dark launch, and
*"behaviour changed for an existing caller"* is architecture-scope however sound
the argument. **The inconsistency was mine — I demanded measured evidence before
enabling the layer I distrusted and exempted the two I preferred.** All three
layers are now gated (`twin_identity_snap`, `stem_twin_snap`), all default OFF,
all counting while off. A side effect worth noting: the change is now entirely
opt-in/default-OFF with no live consumer, which is RFC_022's carved-out shape, so
the `needs_rfc` the architecture seat raised no longer applies.

**The objection that could have made the whole fix dead code**, from
`editquality`: does the `identity_authority` marker actually SURVIVE from the
reconciler to the two write surfaces? `site_plan_pages` has no column to carry it,
and if the surfaces reload from that table the guard silently never fires — *"the
exact silent-guard-indistinguishable-from-a-dead-one failure mode the plan itself
warns about elsewhere"*. It does survive, and **not** through `site_plan_pages`:
`validate_plan`'s `output_field` is `site_plan`, and both surfaces read that key
out of `collected_data` via `extractPagesFromPlan`, which appends each page map
**whole**. That is now **proven rather than argued** —
`TestReconcile_MarkerSurvivesTheStepBoundary` drives the real extractor over a
validate-shaped payload, and is mutation-checked (stop stamping the marker and it
goes red, naming the surviving keys).

Answered with evidence, no code change: `resolveToolPageIdentity` exists
(`deploy_tool_action.go:664`) and does what was cited; and on the
three-`pages`-upsert-helpers landmine — only ONE of the two surfaces writes
`pages` at all (`SyncPagesToDBAction` → `upsertPage`), `WriteSitePlanAction`
writes the PLAN table, and `upsertPage` is the correct helper by that landmine's
own rule (a role arriving from a plan belongs on the plan-sync path).

---

## 2026-08-11 (evening) — COUNCIL APPROVED round 3, and the quiet-mode fix is LIVE on chassis v1.0.1288 (artefact-verified). It is also INERT, deliberately

**Verdict: APPROVED**, corr `56e13695-17cb-48ec-bc6b-0371fde8b717`, decided
17:06:27Z after two REVISE rounds. The trail: revise (16:38) → revise (16:55) →
approved (17:06).

**Live, verified at the artefact on BOTH replicas** (`agent-chassis-596d84f6b-kmc2t`,
`-tb8gd`, image `v1.0.1288`), not at git and not at the tag:

```
POSITIVE  "twin_identity_snap"                 -> present (both pods)
POSITIVE  "PLAN_PAGE_IDENTITY_TWIN_OBSERVED"   -> present (both pods)
POSITIVE  "stem_twin_snap"                     -> present (both pods)
POSITIVE  "honour_realised_identity"           -> present (both pods)
NEGATIVE  "twin_identity_snapp"  (one letter)  -> absent  (both pods)
```

`twin_identity_snap` is the load-bearing one: that literal entered the tree only
in `038211dd8`, the REVISE response, so the running binary carries the FINAL
gated design and not merely the earlier parts.

> **Two notes on how this was verified, because the obvious routes both fail here.**
> (1) The documented `logs -l app=agent-chassis | grep 'build provenance'` returned
> **1.4MB of council payloads** quoting the phrase — the known false-match trap —
> and the startup line itself had rotated out of both pods. (2) Probing
> `/proc/1/exe` for my COMMIT SHAs returned absent for all of them **and** for the
> fabricated control, i.e. no positive control, so it proved nothing. The literal
> probe with a near-miss negative control is what actually settles it.

**STATUS: the mechanism is fixed and shipped; NO SITE IS USING IT YET, by design.**
All three gates default OFF, so `v1.0.1288` changes nothing for any site until a
structure spec opts in. Baseline taken the moment the build landed:

```
PLAN_PAGE_IDENTITY_TWIN_OBSERVED  0
PLAN_PAGE_STEM_TWIN_OBSERVED      0
PLAN_PAGE_IDENTITY_SNAPPED        0
PLAN_PAGE_MERGE_LOSSY             2   (the pre-existing pair, 08-11 10:21)
```

**Those three zeros are NOT evidence of anything yet** — no replan has run since
the roll. The first replan of any site will move `*_OBSERVED` off zero while the
gates are still off; that is the dark-launch evidence the enable decision rests
on, and it is the first thing to read next session.

**This bug stays OPEN**, for two reasons that are now clearly separable:
1. the prevention is live but unproven in the wild (no replan through it yet) and
   unenabled anywhere;
2. the existing damage — 7 both-deployed twin pairs across 4 domains — is
   untouched and needs an owner decision per pair (`RUNBOOK_2026-08-11_duplicate_page_identity_remediation.md`).

## 2026-08-12 — the dark-launch read, and why the counters cannot be waited on

Re-measured the §4 population as the previous handoff instructed. **Still
0/0/0/0** [MEASURED 2026-08-12 13:02Z, `SELECT now()`]. The reason is the absence
of demand, and it is now nailed down rather than assumed:

> **CORRECTED, same session:** this line first read `~03:50Z`. I had inferred my
> own measurement time from the nearest timestamp in the data (noted.co.uk's
> 03:22:51 plan) instead of asking the clock — wrong by nine hours. The cheap check
> is `SELECT now()` / `date -u`, which is now what the stamp quotes. Nothing else
> in this section moves: the pods still started 2026-08-11T21:53Z, so the uptime
> the zeros cover is ~15h, not ~6h, which makes the absence of a replan a slightly
> stronger observation rather than a weaker one.

- Only **one** plan has been written since the roll — `noted.co.uk`
  (`185149a7`, 2026-08-12 03:22:51). It is that site's **only plan ever**, and its
  5 `pages` rows were all created **0.65s AFTER** the plan row (min `created_at`
  03:22:51.900 vs the plan's 03:22:51.254). Zero pre-existing realised pages, so
  `reconcilePlanWithRealised` had nothing to reconcile against. An initial build,
  not a replan.
- fundamentallyai's `40a66d3a` (2026-08-11 10:21:47) **predates** the roll: both
  chassis pods started **2026-08-11T21:53Z**.

**The code survived a further roll and is still live.** The fleet moved on to
`v1.0.1290` (not the `v1.0.1288` recorded above). All four lane commits
(`65c1984d0`, `7a066dba1`, `b36163fb3`, `038211dd8`) are ancestors of `HEAD`, and
the literal probe was re-run on **both** new replicas with a discriminating
control pair:

```
POSITIVE  PLAN_PAGE_MERGE_LOSSY             -> PRESENT (both pods)   pre-dates the lane; proves the probe works
POSITIVE  PLAN_PAGE_IDENTITY_TWIN_OBSERVED  -> PRESENT (both pods)
POSITIVE  PLAN_PAGE_STEM_TWIN_REFUSED       -> PRESENT (both pods)
POSITIVE  stem_twin_snap                    -> PRESENT (both pods)
NEGATIVE  PLAN_PAGE_IDENTITY_TWIN_OBSERVEQ  -> absent  (both pods)   one-letter near-miss
NEGATIVE  stem_twin_snup                    -> absent  (both pods)   one-letter near-miss
```

Gates re-checked with `data ? 'key'` (never `->>'key' = 'true'`) across all **6**
current `structure` specs fleet-wide: **zero sites carry any of the three keys.**
Damage population re-measured with the runbook's own query: **still 7 pairs, 4
domains**, component counts unchanged (robot-hands 5/1, 3/1, 4/1 — matching the
08-11 figures).

### The finding: an OBSERVED row is not a free observation

The previous handoff's ordering — "read the population, THEN enable" — treats the
dark-launch counter as a passive instrument. Reading the code, it is not. With the
gate off, `observeOrSnap` records the observation and **returns without snapping**
(`v3_site_actions.go:5852-5868`), so the plan carries both identities and the
name-keyed upsert proceeds. The remedy string says so in its own words: *"each of
these is a second page identity about to be written."*

So the counter the enable decision rests on populates **only** by letting the
defect occur — with one important exception. Where the plan entry's twin is
**already realised**, an OBSERVED row is pure detection of existing damage and no
new row is minted. Where it is not, one is. **The counter does not distinguish
these two cases; the row's context fields do** — join its `plan_name` back
against `pages` to tell a re-detection from a fresh twin. That join is mandatory
before any OBSERVED count is read as a rate.

> Recorded because I nearly filed the stronger claim that the dark launch is
> *inherently* self-harming. It is not: the harmless case is real, and it is the
> case the known population is actually in. The `eligible` arm
> (`v3_site_actions.go:5834-5846`) is what makes the difference, and I had to read
> it rather than reason from the counter's name.

### What each known pair would actually do on its next replan

Traced through the real layer order (path_key → canonical_name → stem) against
`PageItemStem`'s actual prefix set — `tool-`, `guide-`, `game-`, prefix-only, so
a name *ending* `-guide` is bare (`page_identity.go:74-82`). Which side the
current plan names [MEASURED 2026-08-12]:

| domain | pair | in current plan | predicted signal |
|---|---|---|---|
| robot-hands.com (×3) | bare + `tool-` | **both sides** | `PLAN_PAGE_STEM_TWIN_REFUSED`, ~2 per pair (one per direction). `eligible` false ⇒ path_key/canonical skip **silently**; only the stem layer records |
| fundamentallyai.com (×2) | bare + `tool-` | **bare only** | `PLAN_PAGE_STEM_TWIN_OBSERVED` — stem layer, `planIsBare` ⇒ `byStem` hits the unplanned `tool-` page. **Harmless: both twins already exist, so no new row** |
| ai-agent-orchestration.com | bare + `tool-` | **neither side** | unpredictable — a fresh plan may name either or a third spelling |
| finetuning.uk | bare + `tool-` | **neither side** | as above |

**This validates the previous handoff's recommended pilot, for a reason it did not
state.** Enabling `honour_realised_identity` + `twin_identity_snap` on
fundamentallyai while leaving `stem_twin_snap` OFF yields ~2
`PLAN_PAGE_STEM_TWIN_OBSERVED` rows on its next replan that are **pure
detection** — the stem-layer evidence, at zero cost in new damage, because both
sides of both pairs are already realised and deployed. The two enabled layers
meanwhile guard against new twins. That is the cheap experiment; waiting for the
counter to fill on its own is not, because nothing schedules a replan.

**Also worth knowing before enabling on fundamentallyai:** were
`stem_twin_snap` ON there, the snap would rewrite each bare plan entry onto the
**`tool-` realised page it matched** (`snapPlanPageOntoRealised` returns
`normaliseRealisedToPlanPage(rp)`), i.e. it would move future builds onto the
`tool-` side. Both fundamentallyai pairs are 3 components against 3, so that is a
**survivor choice made by machine** — exactly the per-pair owner call O2 reserves.
Another reason `stem_twin_snap` stays off on that site until O2 is decided.

### The decomposed-site exclusion, actually resolved to domains

Both the handoff and the runbook say "do NOT enable on the five decomposed sites"
without naming them, which makes the rule unusable at the moment you need it. Re-ran
`bugs_open/204`'s own census [MEASURED 2026-08-12 ~13:10Z]:

| domain | slot names | unresolvable |
|---|---|---|
| loancalculator.co.uk | 63 | 63 |
| loanandmortgagecalculator.co.uk | 43 | 21 |
| gaswholesalers.com | 125 | 11 |
| **finetuning.uk** | 162 | **10** |
| leopardessconsulting.co.uk | 172 | 6 |
| oufe.com | 24 | 2 |

**It is SIX sites now, not five** — 204's figure is stale, and this is recorded here
rather than only in 204 because the exclusion list is consumed from *this* lane.

Two consequences for the enable decision:

- **fundamentallyai.com is NOT on the list**, so the recommended pilot is clear of
  the `bugs_open/204` hazard. This was worth checking rather than assuming: the
  pilot recommendation and the exclusion rule were written in the same handoff and
  never reconciled against each other.
- **finetuning.uk IS on the list, and it is also one of the four twin domains.**
  Nothing in the handoff, the runbook or the remediation ordering says so. Its twin
  pair (`ai-readiness-quiz` / `tool-ai-readiness-quiz`) must therefore **not** get
  gates enabled, and its remediation is additionally constrained by 204. Neither
  document flags this overlap; it is now flagged.

### 2026-08-12 14:13Z — O1 DECIDED AND EXECUTED: the pilot is live on fundamentallyai

**Owner decision 2026-08-12:** enable the two safer gates on fundamentallyai, leave
`stem_twin_snap` off, and **wait for a natural rebuild** rather than trigger one.

Seeded and verified. `docs024_key_docs_latest/brochure_component_library/SEED_2026-08-12_fundamentallyai_identity_gates.sql`:

```
domain               fundamentallyai.com
spec id              c4c6b829-8e70-4048-a8c2-a050112ff72d   (created_by brochure_215_quiet_mode_thread)
honour_realised_identity  present, true
twin_identity_snap        present, true
stem_twin_snap            ABSENT  (asserted absent, not merely unset — see below)
url_shape                 ABSENT  (so URL shaping is unchanged)
current structure specs fleet-wide  6 -> 7; exactly ONE carries any gate
```

**This was an INSERT, not the carry-forward every sibling `SEED_*.sql` performs —
fundamentallyai had no structure spec row at all.** Checked rather than assumed that
creating one is inert for everything else: the only other reader of the aspect is
`siteUsesFlatURLs`, whose contract is "absent spec, absent key … all mean false"
(`site_url_shape.go:29-32`), so a row carrying neither `url_shape` nor the adoption
keys is indistinguishable from the previous no-row state for every reader but mine.
The site was framework-built, never adopted, which is why there was nothing to carry.

The seed's verify block asserts the **no-op** as well as the change: it aborts if
`stem_twin_snap` exists *at all*, even as `false`. Absent and false behave identically
today, but an explicit `false` reads to the next operator as "someone considered and
disabled this", which is a different fact from "never set" — and O2 is still open.

**Nothing has happened yet, by design.** The gates bite only when fundamentallyai's
page list is next rebuilt, and no rebuild is scheduled. **So a zero in the counters
tomorrow still means "no replan yet", not "no twins"** — the demand control
(`SELECT max(created_at) FROM site_plans`, and whether that plan's site had pages
predating it) remains mandatory before reading any zero, and the classifying join in
`LANDMINES.md` remains mandatory before reading any non-zero.

Expected first signal, for whoever reads it: **~2 `PLAN_PAGE_STEM_TWIN_OBSERVED` rows**
(the stem layer is the one that fires on this site's pairs, and it is off), plus
whatever the two enabled layers prevent. Those two OBSERVED rows are the harmless
kind — both sides of both pairs are already realised, so observing them mints nothing.

### Unchanged loose end

The 090 diagnosis `38099787-c7f9-46d4-b75e-3a1867fcaf41` (archived pages rebuilt
and re-deployed) is **still verdict-less**: 3 orchestration rows all `COMPLETED`
2026-08-11 13:33–13:34, and **zero** `doc_notes` rows mention the correlation.
Nobody has read a root cause. Runbook finding 3 ("assume any archive can be
undone by the next replan-triggered build") therefore still stands, and it gates
step 5 of the remediation.

### 2026-08-12 (evening) — the counters are still 0/0/0/0, and the "unchanged loose end" above was WRONG

**Counters, re-read with both controls.** Still `0` on all four codes. The demand control
explains it and the instrument control proves the query can see anything at all:
`agent_error_log` took **3,503 rows in the last 24h** (newest 18:37Z), so the zeros are real
absence, not a dead table. The only plan since the roll is still noted.co.uk's first build
(0 `pages` predating its plan row — it cannot exercise the reconciler). **No replan has run
through the new path yet. Nothing to classify, nothing to conclude.**

> **CORRECTED — the "Unchanged loose end" section immediately above is false, and I wrote it.**
> The 090 run `38099787…` was **not** verdict-less. It completed on 08-11 with **five
> `diagnosis_artifacts` bundles** and a correct root cause. My check looked in `doc_notes`,
> where **no diagnosis run has ever written anything** — that query returns `0` for a healthy
> run too, so it could not have come out otherwise. I also searched by the wrong one of the
> item's **two** correlations. Both errors point the same way, which is why `0` felt confirmed.
> Landmine filed (*A 090 diagnosis writes its findings to `diagnosis_artifacts`, NEVER to
> `doc_notes`*); incident in `WRONG_CALLS.md`.

**The diagnosis is now read, verified first-hand and filed as `bugs_open/266`.** It is not
restated here. What matters for *this* file is that it **corrects item 1 of "Two defects found
while doing this" above** — including that item's own correction:

- That item talked itself down from "distinct defect" to "PLAN-017's documented regeneration
  trap, not a new one". **The regeneration trap accounts for one of the two pages.** For
  `tool-llm-cost-calculator`, `reconcile_site_plan` correctly withheld the build
  (`owned_page_review` / `needs_human_review`, still uncompleted today) and
  **`image-build-handler` rebuilt and deployed it anyway, sixteen minutes later**, by an
  unrelated path. Two further producers (`page-rerender`, `section-editor`) have re-deployed
  these pages since.
- **The damage is live and recurring, not historical.** The `deployed_at` stamps this file
  records (08-11 10:34 / 11:13) are already superseded: 08-11 19:05 and **08-12 14:25**, the
  latter four hours before this note. Both pages are still `status='archived'` and both serve
  HTTP 200 (verified against a fabricated-URL 404 control).
- **Consequence for this lane's O2:** runbook finding 3 stands and gets stronger — an archive
  is not durable against **four** producers, so remediating a twin pair by archiving one side
  will be undone. `266` is the blocker to fix; `215`'s own remediation should not assume
  archiving holds until it is.

---

## 2026-08-14 (evening) — O2 remediation: pair 5 executed to step 5, one dispatch owed

Second of the seven pairs to be worked, and the first to clear step 5 cleanly. Full account and
the owed command: `docs024_key_docs_latest/brochure_component_library/HANDOFF_2026-08-12_215_quiet_mode_continue_here.md` §15.

- **robot-hands.com `gripper-payload-calculator`** (`48d52965…`) — owner ruling 2026-08-13, keep
  `tool-gripper-payload-calculator`. Steps 3 (plan surgery: 1 + 3 rows), 4 (9 work items
  cancelled) and 5 (archived) landed in one asserted transaction; the survivor was asserted
  `active` and in-plan inside it. **Step 6 (retraction) not run — the harness permission
  classifier refuses the Kafka publish. Not a platform refusal.**
- **The inbound census was re-run before mutating and carries its own positive control**: same
  query, same site, same run — pair 5 zero, pair 7 zero, **pair 6 four editorial + one nav**.
  Reproduces §14's read-only table exactly.
- **Two procedural corrections went into the runbook** (its 2026-08-14 evening amendment):
  "open" work items means `workItemClosedStatuses`, not `workItemTerminalStatuses` — three of
  the nine were `unresolved`, which RFC_010 rules OPEN; and step 4 is not durable on its own,
  because the fleet sweep re-queues a rerender per **active** page (31 on this site today), while
  an **archived** page is excluded (11 archived on-site, 10 of them untouched by the wave).
- **This pair will not supply `266`'s behavioural proof.** Archiving removed the page from the
  rerender population and its two queued items were cancelled, so no producer is now aimed at it.
  `ARCHIVED_PAGE_%` remains 0 with a live instrument (1,003 `agent_error_log` rows in 24h).

### Same evening — pair 5 COMPLETE, all 8 steps

Retraction dispatched 16:59:26Z (corr `5a574b41…`), COMPLETED in 6s, `delete_file` removed
`robot-hands.com/gripper-payload-calculator.html`. **No refusal — the read-only inbound census
predicted a clean pass and was right**, which validates it in both directions (it reproduced
pair 1's real refusal and now predicted this pass). Artefact: loser **404 at 2,886 b, identical
to the fabricated-URL control**; survivor 200/34,157 b and two collateral pages unchanged to the
byte. **Part 2 of `098`'s acceptance is owed — it must still 404 after the ~20:0x news refresh.**
`deployed_at` is untouched by retraction, so this page is now a deliberate false positive for the
blind `archived AND deployed_at IS NOT NULL` detector. First of the seven pairs finished.

---

## CONTRIBUTION 2026-08-17 — first live consumer of `honour_realised_identity`, and the twins were minted anyway

From the `loanandmortgagecalculator.co.uk` D6 planner lane. Reporting a measurement, not a
diagnosis.

**We are (as far as we can see) this flag's first live consumer.** Before firing we measured
the population the flag exists for, using the real `datahelpers.CanonicalisePage` through the
descriptor `write_site_plan_action.go:487` actually builds (`Role` = stored `page_type`,
`Slug` = `firstNonEmpty(slug,name)` → the stored NAME for a realised page, `ParentSection` =
`parentSectionFromURL(url)`, `FlatURLs` = false), with a positive and a negative control in the
harness so an inert run could not report "nothing moved":

```
site ed633ada-f8af-424b-b4d4-8af79160dbcd, 45 active pages:
  7 fixed points, 38 moved (name 17, url 38, type 0)
```

All 17 name moves are calculator pages, e.g. `mortgages-stamp-duty` →
`tool-mortgages-stamp-duty` at `/mortgages/mortgages-stamp-duty/index.html` (the doubled
segment is the stored name being used as the slug).

**Then we seeded `honour_realised_identity='true'` and fired one canary — and 19 phantom
pages were INSERTED anyway**, 17 of them exactly the predicted `tool-<name>` twins, at
`/tools/<name>/index.html`. Verified before the fire that the key was live in the current
structure row (`6ca809d6`, value the string `'true'`, which `(…)::boolean` accepts). Chassis
`v1.0.1305`, corr `6fe6ee93-67b9-4831-bf17-2ca473e1d30c`, COMPLETED 12:07:05Z. All 19 rows
had zero `page_components` and zero references in the 8 tables that FK to `pages`; deleted
after measurement, and the live site never served them (phantom paths 404).

**Candidate links, none established by us:**
1. **The reconciler may never have paired the plan pages with the realised rows.**
   `twin_identity_snap` and `stem_twin_snap` were **absent** on this site — a deliberate
   "one canary, one question" choice that now looks like the likely cause, since
   `stem_twin_snap` is described as matching a bare plan page against a prefixed realised
   one, which is exactly this pairing. Honouring a realised identity cannot help a page the
   reconciler never marked `identity_authority='realised'`.
2. The site-spec reader may not be in the running binary. ⚠ **We could not settle this and
   our probe was uninformative, not negative**: 7- and 9-char sha prefixes for recent
   commits all came back absent from `/proc/1/exe` — *including the negative control* — and
   the startup provenance line has scrolled (pods ~14 h old), with no commit env var set.

Filed for diagnosis rather than asserted: 090 run correlation
`33d4d7bc-62f8-4886-a8e2-7c39f0c0a302`. If (1) is the answer, the useful outcome for this
file is that the identity flag has a **precondition** — it is inert without a pairing layer —
and that is worth stating where the flag is documented, because we read
`site_identity_policy.go` closely, measured the population it asks for, and still got this
wrong.

Full account, incl. the repair and its assertions:
`docs024_key_docs_latest/loanandmortgagecalculator_couk/NOTES_…md` entry **2026-08-17 (d)**.

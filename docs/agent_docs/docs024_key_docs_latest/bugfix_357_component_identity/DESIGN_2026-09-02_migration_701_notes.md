# DESIGN_2026-09-02_migration_701_notes.md — design notes for the 701 migration pair (bugs_open/357 phase 3, Option B)

Drafted 2026-09-02 by the drafting agent, for the parent session and the council round.
Files (scratchpad, NOT the repo):

- `701_retype_357_population_by_adoption_HOLD.sql`
- `701_retype_357_population_by_adoption_HOLD_ROLLBACK.sql`

Both parse clean under libpg_query (pglast 8.4): 20 + 10 statements, and all 10 `DO`
block bodies parse as plpgsql. NOT executed anywhere — this session was read-only at
the DB by instruction, so there has been no rehearsal against a real server. The
applier should expect first-execution wrinkles of the kind a parse cannot catch
(type-resolution corners, trigger interactions) and apply the pilot scope first for
that reason too.

## 1. Two live findings that CONTRADICT the brief's measured facts

Both re-measured 2026-09-02 (today) while drafting; everything else in the brief
checked out exactly (see §12).

### 1a. gamesdesign/tool-ttk-calculator has ZERO site_plan_sections rows — not one

The brief: "Each page has EXACTLY ONE site_plan_sections row with
component_name='hero' at ordering 0 in its site's single current plan." True for 21
of 22. For `tool-ttk-calculator` the site's single plan (c96b501c…, `is_current`,
created 2026-06-05, 44 rows) carries NO rows for that page_name at all — checked by
tuple-join, by equality, and by `LIKE '%ttk%'` across ALL plans of the site. The
page row exists and its `pages.sections` reads `["hero","generic-text-block"]`, so
the page predates or bypassed the plan; sync_pages upserts FROM plan rows, so a page
absent from the plan is simply never touched by it — which is also why its
`pages.sections` repoint is durable on its own.

Handling: the census pins `has_plan_row=false` for this one row; the plan-leg
UPDATE skips it; Guard 2 asserts the current plan STILL has zero rows for it (a
plan row appearing before apply = changed premise = abort); the backup stores
`pre_plan_row_id NULL`; verify expects 21 plan repoints on `scope=all`
(`plans_expected` is computed from the census, so the arithmetic follows the flag).

Worth knowing while I measured it: my own first query batch appeared to show 2 plan
rows for ttk; that was my misread of a GROUP BY listing (ttk was absent from it,
not present with 2). The zero-rows result held across three independent query
shapes. I flag this so nobody "corrects" the census back on the strength of a
similar misread.

### 1b. "exactly one function collision" is true at the library predicate, but there is a SECOND row on the function

The brief pinned one collision: `tool-equity-release_pre_037`
(a5236dec-c38e-46b5-8f05-4f7b43fa2f3f) claims function `tool-equity-release` under
`idx_cc_tool_function_unique`'s predicate. Confirmed. But the same site ALSO
already holds `tool-equity-release_pre_037-mortgagecalculator-co-uk`
(befacff0-4a5d-4fb6-98c4-4de5834bc299, `forked_from=a5236dec`, active,
created_from='manual') — a deploy_tool_to_site-shaped fork, with **0 placements**
(no page_components, no site_components, no plan row names it).

Consequences, stated for the council rather than solved here:

- It does not block anything: forks are outside the unique index, and our new name
  (`tool-equity-release-mortgagecalculator-co-uk`) does not collide with the fork's
  name (`…_pre_037-mortgagecalculator-co-uk`) — the two naming conventions differ,
  which is RFC_036 §9.2's own observation.
- After apply, mortgagecalculator carries TWO site copies of the equity-release
  library tool: the unplaced manual fork and our placed adopted fork. That is
  exactly the tail RFC_036 §11 addendum 2 tracks ("the two fork producers are not
  recognised as the same copy"). Nothing breaks (the page layer would fail loudly
  on `pages_site_id_name_key` if the deploy path ever tried to place its fork on
  this page), but a reviewer grepping forks of a5236dec will see both.
- Guard 1 asserts the pinned parent still exists as an active library claim with
  `forked_from NULL`, and aborts on ANY unpinned library claim over the other 21
  functions.

## 2. RFC_036 §9.3 — verified, and it prescribes exactly what the brief said

§9.3 verbatim mechanism: look up a library tool claiming the function
(`component_level='tool' AND forked_from IS NULL AND is_active`, no site filter);
if one exists, set the new component's `forked_from` to its id; nothing else
changes. That is precisely `forked_from='a5236dec-…'` for equity-release. No
discrepancy with the brief. (§9.3 is BUILT and LIVE per §11 addenda — but that is
the Go writer; a migration INSERT does not pass through it and must carry the fork
itself, which this one does.)

## 3. The vetcomparison.uk naming decision

**Chosen: function `tool-vet-comparison`, name `tool-vet-comparison-vetcomparison-uk`.**

- The mechanical CLC-020 derivation (function = page name = `index`) would claim
  the fleet-wide tool function `index` under `idx_cc_tool_function_unique` — the
  one string every site's homepage shares. Today NOTHING claims `index` (measured:
  zero content_components rows carry that function at any level), so the mechanical
  choice would not abort — it would succeed and then poison the name: any future
  tool adoption or native build on any site's index page would be silently diverted
  into a fork of a vet-comparison widget. A terrible library identity is worse than
  a naming exception.
- The cost 693 would have paid — its repoint joins `cc.function = p.name`, which a
  renamed function breaks — is NOT paid here, because migration 701 never joins on
  function: every step keys on the pinned `pc_id` census, and the naming is DATA in
  that census. The joins stay honest for all 22 rows including this one.
- The trade-off honestly stated: the page name (`index`) and the component function
  (`tool-vet-comparison`) now disagree for this one page, so any future tooling
  that ASSUMES function==page-name for adopted tools has a counterexample. Nothing
  live assumes it (resolution is by component_id first, then by slot/plan-element
  NAME, which is the cc.name and matches); the census row and the doc_note record
  the exception.
- `investor-index` kept its mechanical derivation (`investor-index`) — distinctive
  fleet-wide, so the `index` argument does not apply, and the brief only opened the
  vetcomparison case for deliberation.

## 4. The alignment: plan element = slot_name = cc.name (deviation from BOTH 578 and 693)

- 578 preserved `slot_name='hero'` because it did not move the plan: a lone slot
  rename would break Layer 2's slot-equality match and arm the re-append landmine
  (tool + fresh hero band). Its safety then rested on the UNTESTED identity-carry —
  the very dependency (precondition 4) the lane proved untestable-in-place
  (HANDOFF_2026-08-26b Findings 1–2).
- 693 set `slot=p.name` but its plans were empty, so nothing constrained it.
- 701 moves plan element AND slot TOGETHER, in one transaction, to one string: the
  new cc.name. Stored `component_id` resolves first at rebuild regardless
  (rerender_page_sections_action.go `resolveComponent`; loadComponentSchemasByID
  reads content_components by id); the plan element is the writer-path key; Layer 2
  matches on slot. With all three equal — and the string globally unique by
  `content_components_name_key`, unlike `'hero'` — every name-keyed path resolves
  to the same component the id path does, and the splice match is trivial rather
  than carried. This is the property that makes the regeneration arm a no-op BY
  CONSTRUCTION (template IS the bytes), which is the whole reason Option B won.

## 5. Pilot mechanism (the "simplest honest mechanism" asked for)

One migration file, three scopes, carried as a psql variable and injected into the
DO blocks via a transaction-local GUC (`set_config('m701.scope', :'scope', true)` —
psql does not interpolate `:variables` inside dollar-quoted bodies, and a
transaction-local GUC dies at COMMIT/ABORT so nothing leaks):

- `-v scope=pilot` → mortgagecalculator/tool-simple only (pc cb406ec9…). **This is
  also the DEFAULT when the variable is omitted** — the smallest honest action,
  matching the lane's stated plan. Defaulting to `all` would let a naive run skip
  the staging the owner asked for.
- `-v scope=remainder` → the other 21; Guard 2 REFUSES unless the pilot row is
  verified already repaired on all three legs.
- `-v scope=all` → all 22 at once, for an owner decision to skip the pilot.
  Running `all` after the pilot ran aborts with a readable message (the pilot row
  is no longer censused-exact) steering to `remainder`.

Guard 2 is scope-aware but always checks ALL 22 pinned rows plus the fleet-wide
predicate, so drift ANYWHERE in the population aborts every scope. A deliberate
side effect: run the SQL through anything that skips the psql preamble and
`current_setting('m701.scope')` raises — the file refuses to run outside its own
harness (and the `_HOLD` suffix keeps the runner away regardless).

578 solved scoping differently (predicate-selects-all, prints before writing);
the 701 way was chosen because the pilot/remainder split has to be REPEATABLE
across two hand runs with the second run able to PROVE the first happened.

## 6. Why census-pinned selection, not 578's predicate-driven selection

Two of the measurements that license adoption — zero `{{` bindings, and every body
PASSING `toolTemplateValid` through the real production function with both-way
controls — were taken against these 22 bodies specifically. A row minted after the
census has neither measurement, and `toolTemplateValid` cannot be re-run from SQL.
So the population predicate is re-run at apply time as a GUARD (exact-set match,
abort on growth or shrinkage), never as a SELECTOR. Growth additionally means the
phase 2 birth-fix regressed, which is worth knowing loudly. The `{{` check alone
IS re-runnable in SQL and is kept as Guard 3 (693's shape) belt-and-braces.

## 7. Rerender filing: generic pages only — a deliberate deviation from "copy 693 exactly"

The filing SHAPE is 693's verbatim (item_key `page_rerender_<name>_<siteid>_assemble`,
status `triaged`, handler `page-rerender`, priority 80, page_id first-class,
created_by/source naming this lane+migration, spec with
domain/page_id/filename=ltrim(url,'/')/page_name/reason, stacked-rebuild abort
guard). The TARGETING deviates: only `rebuild_policy='generic'` targets get one —
16 of 22 (`all`), 15 (`remainder`), 1 (`pilot`). The six gamesdesign owned rows are
refused at assemble by the owned-page guard (578's own §6 analysis: the pipeline
refuses these pages, which is also why a repaired owned row STAYS repaired — their
repair needs no rebuild and cannot be undone by one). Filing theirs would mint six
guaranteed failures into the immune system's failure sweeps. Burst: 16 at once is
within the live producer's observed batch size (a 'rerender-pages' batch of 20+ on
2026-09-01, cited in 693's council round) at the same priority 80. Measured today:
the only live page_rerender rows on the three sites are 12 `deferred` rows from
2026-08-03 on OTHER mortgagecalculator pages — no key overlap; the guard would
catch one anyway.

Note the pilot pass condition in the HOLD header is Option B's own, NOT
Finding 2's (`sections_saved > 0` belonged to the 578 shape, where the plan still
said 'hero' and the carry had to be exercised; under 701 the plan element names the
adopted component, so the rebuild's write path renders the adopted template —
byte-identical — and the things to verify are row count still 1, md5 unchanged,
component still the adopted one, served page still the tool).

## 8. component_versions: none minted — verified, not assumed

693 minted none; the brief asked me to confirm render does not REQUIRE a version
row. Checked at the code: `loadComponentSchemasByID`
(plan_sections_action.go:2063) reads content_components only;
`resolveComponent` (rerender_page_sections_action.go:~361) resolves from that map;
grep for `component_versions` across rerender/plan/save action files: zero hits —
`page_components.component_version_id` is only carried as a provenance stamp
(rerender_page_sections_action.go:164-168, copied through if present). So no
version row is needed and none is created; the repaired rows keep
`component_version_id NULL` exactly like 693's. (578 needed a version row because
its shape bound content through the adopted-fragment `{{.body}}` template — a
different design, not a precedent for this one.)

## 9. Backup design

One table, `page_components_backup_357b_20260902` (the brief's name):
per touched row — the WHOLE page_components row as `to_jsonb(pc.*)` (578's
back-up-everything practice), the load-bearing columns extracted for exact
restores (pre component_id/slot/position/md5/bytes), the `pages.sections`
pre-state, the plan row's id + `to_jsonb(sps.*)` pre-state (NULL for ttk), the new
name/function it was repointed to (so the ROLLBACK is self-contained), and the
scope + timestamp of the applying run. `ON CONFLICT (pc_id) DO NOTHING` on the
insert is safe ONLY because Guard 2 has just proven the live pre-state equals the
pinned census — a re-apply after a rollback stores byte-identical facts. The
rollback KEEPS the table.

## 10. Rollback deviations from 693's (each argued in its header too)

1. **Strict per-row state classification** (post-701 → restore; already pre-701 →
   skip; anything else → abort naming the row) instead of a blanket predicate
   UPDATE. Three legs moved here (row, plan, sections) and they can drift
   independently; a blanket restore would silently half-restore a drifted page.
2. **Wider unreferenced check before DELETE**: page_components AND site_components
   AND `forked_from` (someone may have forked FROM an adopted row — §9.3 is live,
   so a future native build of e.g. `tool-stamp-duty` anywhere in the fleet would
   do exactly that). A referenced survivor aborts the WHOLE rollback — it is
   all-or-nothing by design; partial rollback is a hand-edit, not a mode.
3. **A rollback doc_notes row is filed**, so the decision trail does not end on
   "adoption happened" after it was undone. 693 did not do this; extension, cheap,
   and it keeps doc_notes honest.
4. **The deployed_at warning is inverted for this case**: lendzy's pages could not
   deploy at all, so its warning was about a stamp appearing. Ours ALREADY serve
   correctly on both sides (the template is byte-identical to the artefact), so
   the warning is about REBUILDS: if a 701-filed rerender ran, the artefact was
   assembled FROM the new identity, and rolling the record back both makes the
   record lie about provenance and re-arms 357's §3 disaster (the next rebuild
   re-mints a 2KB hero band over the working tool). The header carries the exact
   query to run first, and says plainly there is no artefact-level reason to hurry
   a rollback.

## 11. A consequence the council should weigh: 21 new fleet-wide library claims

The 21 non-fork rows are, by the platform's own definition
(deploy_tool_action.go:11-12: library tool = `component_level='tool' AND
forked_from IS NULL`), LIBRARY entries claiming their functions fleet-wide under
`idx_cc_tool_function_unique` — including portfolio-relevant names like
`tool-stamp-duty`, `tool-repayment`, `tool-affordability`, `tool-portfolio`,
`tool-simple`. RFC_036 §10 warns the ~140-domain finance wave shares calculator
names by construction. What actually happens post-§9.3 (live since v1.0.1316): a
future site's native build of `tool-stamp-duty` is recorded as a FORK of
mortgagecalculator's adopted body — its own generated HTML, only the `forked_from`
bookkeeping is odd ("fork" implying a derivation that did not happen). §11's
measured finding stands: nothing filters tool rows on `forked_from` except the
fork mechanism itself, so this is bookkeeping, not behaviour. 693 accepted the
identical trade for its three names. Alternatives (marking ours as forks of
nothing, or inactive) are ruled out — `forked_from` must point at a real row, and
the resolve path requires `is_active`. Flagged so the round can accept it
knowingly rather than discover it.

## 12. What was re-verified live today (all read-only), and what is taken on trust

Confirmed by direct measurement 2026-09-02:

- The population predicate (component_id=23f95f00-… AND hero literal prefix absent
  from rendered_html) returns EXACTLY the pinned 22 pc_ids; every md5 and byte
  count matches the brief's census; all slot 'hero', position 1; 16 generic +
  6 owned as stated.
- Page ids, site ids, urls, `pages.sections` arrays (each contains exactly one
  "hero"); `tool-simple` planned=1; vetcomparison index is
  `["hero","info-card-grid","latest-news","call-to-action"]`.
- 21 of 22 have exactly one current-plan hero row at ordering 0, version NULL
  (§1a for the 22nd); one current plan per site (unique partial index
  idx_site_plans_current, and counted).
- Function claims: exactly one library claim across the 22 target functions
  (a5236dec, pinned); zero claims on `tool-vet-comparison` AND on `index`;
  the befacff0 fork exists with 0 placements (§1b).
- All 22 proposed names are free against `content_components_name_key` (which is
  TOTAL — no predicate — hence Guard 4).
- `trg_cc_refuse_null_section_type` (refuse_selector_invisible_section) EXEMPTS
  `component_level='tool'` rows — read from pg_proc, so 693's
  section_type-NULL insert shape is legal here too.
- `chk_created_from_valid` permits 'adopted'; `chk_function_kebab_case` passes all
  22 functions (incl. `game-p2p-networking`, `tool-vet-comparison`).
- `idx_swi_dedup`'s terminal-status list matches the guard's literal list (read
  from `\d site_work_items`); the 12 live (deferred) page_rerender rows on these
  sites have no key overlap with ours.
- sync_pages/reconcile: `reconcile_site_plan_action.go:599` confirms
  `sections = EXCLUDED.sections` (pages.sections is the derived copy);
  upsert shape in site_db_actions.go:1266-1268.
- The resolve path needs no component_versions row (§8).

Taken on trust from the parent's brief (dated 2026-09-02, not re-runnable from
SQL): the toolTemplateValid pass on all 22 bodies with both-way controls, and the
drift-reconciler premise (compares `built_from_plan_version` to the current
`site_plans.id`; no new plan row is minted here and that column is untouched, so
no drift storm — consistent with what I read, but I did not re-trace the
reconciler's arm myself).

## 13. Reviewer questions I expect, with answers

- **"Why does the UPDATE not archive the old bytes?"** No archive trigger fires
  because none needs to: `trg_page_component_artefact_archive_upd` keys on
  rendered_html changes and `…content_archive_upd` on content_data changes;
  this migration touches neither. The digest pair
  (rendered_html / rendered_html_digest) stays in lockstep, so the divergence
  guards correctly see nothing (578's §"CANNOT TRIP THE DIVERGENCE GUARDS"
  analysis carries over verbatim — we move even less than it did).
- **"What about the stale content_data on the repointed rows?"** Left untouched,
  deliberately: the adopted template binds nothing, so render cannot consume it;
  nulling it would fire the content-archive trigger and change more state than
  the repair needs. It is inert residue of the hero mislabel; noted here so it is
  not rediscovered as a mystery.
- **"Can the UPDATE hit uq_page_components_no_byte_identical_duplicate?"** No: the
  new (page_id, slot_name, …, component_id) tuples are unique per page by
  construction — each page has exactly one row in the census and the new slot
  string exists nowhere else on the page (Guard 2 asserts pages.sections carries
  zero of the new name pre-apply; slots and sections are written by the same
  pipeline).
- **"What if another session edits the shared hero template mid-window?"** (The
  webdesign colour-churn landmine makes this real.) Guard 2 recomputes the prefix
  at apply time; a reshaped template either empties the prefix (explicit abort) or
  shifts the predicate so the exact-set match fails (abort naming rows). No arm of
  the guard trusts the census's own predicate result.
- **"Why one doc_note per site per run?"** The trail then records what actually
  happened and when (a pilot on one date, a remainder on another), rather than one
  row asserting a repair that was half-applied at the time it was written.
- **"pages UPDATE fires trg_invalidate_nav_on_page_change."** Yes — nav cache
  invalidation, by design idempotent and cheap; the repointed pages keep their
  nav_label/in_header untouched so the rebuilt cache is identical.
- **"Is 578 dead?"** For this population, yes by owner decision (Option B); the
  701 header says so and forbids applying both. 578's file is untouched (the brief
  forbade touching it) and remains the record of the alternative.

Drafting correction worth recording (caught in self-review, fixed before
delivery): Guard 1's first cut checked library claims across ALL census
functions, which would have made the migration REFUSE ITS OWN remainder run —
after the pilot, the pilot's adopted component is itself an active library claim
on `tool-simple` with `forked_from NULL`. Guard 1 now considers TARGET rows only
(safe because the 22 functions are pairwise distinct, so a non-target's claim can
never collide with a target's INSERT). The cheap check that caught it: walking
each guard through the scope=remainder state table by hand. A council reviewer
should re-walk the same table.

## 14. Open items for the parent session (not blockers)

1. The 578 file's own header/preconditions now describe a superseded plan for this
   population — consider a dated correction note IN the lane docs (not in 578
   itself without owner say-so) pointing at 701.
2. The concept-register/LANDMINES follow-through once applied: the ttk zero-plan
   finding (§1a) and the two-site-copies state on equity-release (§1b) are both
   the kind of fact the next session will trip over.
3. Council submission should name BOTH the migration and the rollback, quote §11
   (the 21 new library claims) in the risks block, and cite the pinned census as
   grounded_in evidence.
4. After the pilot: verify at the ARTEFACT with the invented-URL control
   (parked-domain landmine), and count the page's rows — a count of 2 is the
   carry-forward landmine and stops the remainder run.

---

## ROUND 2 (2026-09-02) — the REVISE answered, objection by objection

Round 1 verdict: REVISE, gated by debug_historian (HIGH). Every objection either answered by
measurement or fixed in the files; nothing defended.

1. **debug_historian HIGH — the 357-keyed landmine ("adopted row + needs_rebuild crashes"):**
   engaged in the HOLD header, mechanism not hope: the cv1 crash NEEDS an empty render
   (`{{.body}}` → 0 sections → no page_html → 408 recursion); 701's templates are
   binding-free (Guard 3 aborts otherwise) so a rebuild renders them TO THEIR LITERAL BYTES —
   the no-content path is unreachable. needs_rebuild writers enumerated (4 live ones) — the
   engagement is real. Belt-and-braces: 408's fix is committed+approved and rolls soon.
   701's own filings are page_rerender (path measured free of 408), never needs_rebuild.
2. **debug_historian medium — served-page verification method:** specified in the header
   (browser-like Accept; wait out the measured 11–97s publish lag; assert tool MARKERS not
   byte-equality; invented-URL control).
3. **editquality medium — name uniqueness "asserted not verified":** it is enforced —
   `content_components_name_key` UNIQUE CONSTRAINT btree(name), read from `\d
   content_components` 2026-09-02 (also carried in the round-1 header text). Now in
   grounded_in with the source.
4. **editquality low — DESIGN notes counted as an edit:** accepted; r2 plan lists the two SQL
   files only.
5. **bug_historian medium — root cause not cited:** a fair catch on the RATIONALE, not the
   design. The birth mechanism IS fixed and proven: 357 **phase 2** ("stop the mislabelling
   at birth") live in production since 2026-08-25 12:24Z, two organic adoptions stamped
   correctly, survived multiple rolls; `population_stamped = 0` throughout. The related
   refusal defect is filed separately (`bugs_open/406`). Guard 2's growth-abort doubles as
   the regression tripwire: a 23rd row ABORTS the repair and names phase 2 as regressed.
6. **tooling_provenance medium — birth-path guards bypassed by raw INSERT:** the guard set is
   covered by measurement rather than by the code path: `toolTemplateValid` (the structural
   half) probed 22/22 through the REAL function with both-way controls;
   `TemplateNeedsInstanceID` is `{{-?\s*\.InstanceID`-shaped (component_instance_scope.go:203)
   and cannot match a body with zero `{{` — false for all 22 by the same losslessness
   measurement Guard 3 re-asserts. These bodies are rendered singleton INSTANCES already in
   production, not parameterised templates awaiting instantiation. Same trade 693 took,
   approved.
7. **tooling_provenance medium/missing — doc_plans tool fences:** MEASURED, they EXIST: 10
   rows over 8 stripped-form subject_keys (bridging-loan, equity-release×2, fee-analyser,
   overpayment, rate-forecaster, repayment, simple, stamp-duty — all mortgagecalculator).
   None mentions component/slot/hero in its body (queried 2026-09-02); the fences key on the
   tool NAME, which 701 does not change. Tier-2 acceptance unaffected, measured not asserted.
8. **guardian medium — do discovery checks treat a library row as "already built" for other
   sites:** NO — `check_missing_tools.go` counts tools by JOIN through `page_components` on
   THE SITE (placement-scoped, read at :85-92), and `tool_eligibility.go` likewise joins
   through placements. Disclosed side effect in the right direction: these three sites'
   deployed-tool counts CORRECT upward post-701 (mcalc +11, gamesdesign +10, vet +1) — today's
   mislabelling undercounts them, so the missing-tools suggester may currently be proposing
   tools these sites already have.
9. **guardian low — "no key overlap" verified:** query run 2026-09-02, ZERO non-terminal
   items match the 701 filing pattern on the three sites. Also: the vetcomparison lane
   confirmed CLEAR (zero dispatchable items on the index page — 40 open, none satisfying the
   dispatcher's predicates) and no sequencing need.
10. **guardian low — rollback vs pending work items:** fixed in the rollback (§4b): ABORTS if
    any non-terminal rerender/rebuild item from another source targets an affected page
    (cancelling another lane's items is not ours to do), with the ordering note that pc
    restore precedes component delete so even a race resolves against the restored row.

---

## ROUND 3 (2026-09-02) — round 2's fresh objections answered

Round 2 was a REAL second review (round 1's answered objections all turned to approvals —
editquality, tooling_provenance, debug_historian now approve). New gate: prior_art_librarian
(HIGH). Every round-2 objection:

1. **prior_art HIGH — ConvertTemplateToInstanceScope / ScopeToolBirthTemplate exist:** read at
   source; they are the bugs_open/283 programme — multi-instance id-scoping of SHARED
   templates by `{{.InstanceID}}-` rewriting. Declined for a mechanism reason now in the HOLD
   header: applying them REWRITES SERVED BYTES (violates byte-for-byte adoption) and the
   defect class they close (same-page multi-instance id collisions) cannot exist for
   single-placement adopted bodies. Disclosed: a future second placement re-opens the class —
   the pre-existing property of every adopted/owned tool. 693 minted its components the same
   way; its own prior-art gate was a DIFFERENT question (inert repair — no rerender half),
   which this file has carried since round 1.
2. **prior_art medium — §9.3 liveness is a deployed-binary claim:** probed, not prescribed:
   agent-chassis v1.0.1354, "new component forks from it" PRESENT + positive/negative
   controls both correct. The re-run command is in the header for future appliers.
3. **bug_historian medium (095/039 wrong-slot shape):** the verify already asserts the
   three-way equality per row (HOLD :670 `cc.name = c.new_name AND pc.slot_name =
   c.new_name`; plan-leg count requires `component_name = new_name`) — cited rather than
   added. ttk (no plan leg): the pages.sections leg still aligns and resolution is id-first.
4. **bug_historian medium (binding-free forever):** accepted and disclosed in the header —
   `template_closed` measured to be a quality field, not a lock; no standing invariant
   borrowed or minted (a new detector is its own track, per this council's own
   seam-in-a-bug-patch rule); worst rebuild outcome since v1.0.1354 (408 fix aboard, probed)
   is a visible empty render, not a crash.
5. **bug_historian low (rollback ttk classification):** already explicit —
   `IF r.pre_plan_row_id IS NOT NULL THEN` guards the plan-leg check; the NULL case skips
   plan classification by construction.
6. **reuse_agent medium (hand-rolled backup vs the history triggers):** the triggers EXIST
   (`trg_page_component_artefact_archive_upd/_del`, verified in pg_trigger) and are now cited
   in the header as the automatic net; the table is kept for the two legs the trigger does
   not cover (pages.sections, plan row), the keyed new_name restore mapping, and STY-056
   (history rows can be pruned).
7. **reuse_agent low (693's prior rounds):** queried — three rounds, r1 bug_historian gate,
   r2 prior_art gate (the inert-repair question), r3 APPROVED + 3 advisories. Cited with the
   honest caveat that its prior-art gate content differs from ours.
8. **guardian medium/missing (who owns the function namespace):** no register names an owner;
   the header now states the 21 claims are an OWNER-LEVEL ACCEPTANCE and hand-apply is the
   sign-off moment, per CLC-020 naming.
9. **guardian low (701 number collision):** pre-apply `ls | grep '^701'` line added — this
   file was already renumbered from 700 for exactly this.
10. **guardian lows (key overlap, rollback pending items):** unchanged from round 2's
    answers — overlap query zero, rollback §4b aborts on other lanes' pending items.

**Milestone recorded in passing:** the §9.3 probe run also confirmed **v1.0.1354 carries the
bugs_open/408 fix** (`paths_tried` PRESENT against the webdesign-tool-rebuilds lane's proven
discriminating baseline) — 408 is fixed AND live at the binary; its §6 end-to-end check is
now cheap and owed separately.

---

## APPROVED — round 3, corr `df6c1b41`, 2 advisories (none high), both dispositioned

1. **bug_historian (medium) — the 22 newly enter check_tool_health → tool-improver scope,
   and case 012 (improver truncates component) is the documented damage shape.** ACCEPTED AS
   INTENDED with the guards named: entering tool-health scope is what being real library
   tools MEANS (693's doc_note says "intentionally" and ours carries the same); 012 is
   CLOSED and its class is exactly what `toolTemplateValid` was calibrated against — a
   truncated rewrite FAILS the guard and the re-render loop stays on the carry-stored-HTML
   path rather than deploying broken markup. The pilot's verification includes the served
   page, which would catch an improver-shaped regression on the one page before the
   remainder runs.
2. **bug_historian (low) — case 044's empty-schema-deferral heuristic vs our
   `input_schema` NULL adopted rows:** NAMED AS A PILOT WATCH ITEM — if a re-plan defers the
   adopted section by that heuristic, the pilot's save shape shows it (section absent from
   the writer output) before the remainder runs. Resolution is id-first regardless.
3. **guardian (medium) — the 21 fleet-wide claims:** standing acceptance, unchanged — the
   owner-acceptance line in the header is the decision record.
4. **guardian (low) — gamesdesign has no owning lane:** header now says the owner's
   hand-apply EXPLICITLY covers gamesdesign's 10 rows (nobody else can speak for that site).

**Verdict read in full; commits may carry `Council-Reviewed: df6c1b41-b600-41d1-8f7e-3e96fe422b31`.**

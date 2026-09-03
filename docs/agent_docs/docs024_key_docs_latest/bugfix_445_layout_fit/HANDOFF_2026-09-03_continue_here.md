> **⚠ SUPERSEDED 2026-09-03 (evening) by `HANDOFF_2026-09-03b_continue_here.md` — in the same
> directory. Do NOT act on §1 or §3 of this file.** Two of §1's three liveness rows have
> flipped: the Phase 1 Go fix IS now live (`v1.0.1359`), and the second council verdict §1
> tells you to read **never existed** (its run was killed by the 12:06Z roll and swept FAILED;
> resubmitted as `adfa4d03-67a8-419f-bc22-d0ef125f94ee`). §3's pre-registered prediction 1 has
> been **RETIRED by evidence** — coverage rose to 100% on the first real classification and
> the tags described the wrong site, so the threshold cannot be re-derived on post-16:54Z
> data. Kept unedited below as the record of what this lane believed at midday.**

# HANDOFF — bugs_open/445, the layout library gap — 2026-09-03 — continue here

**Read this, then `README_where_we_are.md` (plain prose), then `NOTES_layout_fit.md` (the
missteps). Verify §1 before believing anything about liveness — every session before you that
skipped that step published something false.**

Session that wrote this: "bugs_open/445". Owner ruling 2026-09-03 (relayed via the
`designblog.co.uk` lane): *"a thread has taken bug 445"* — **this lane owns BOTH the detector and
the archetype.** Both have now shipped; what remains is the fleet sweep and the reusable guard.

---

## 0. What 445 turned out to be (one paragraph, so you don't re-derive it)

445 said the layout library lacks an archetype for "content hub with embedded tools" — true. The
mechanism underneath: **the estate could not SEE a library gap of that shape, by construction.**
The growth signal fired only when a layout's TOTAL score was zero library-wide, and the
category/description/scheme bonuses are added independently of tag matching, so a layout matching
NONE of a site's tags still scored above zero. Four live sites were recorded as `tags 0.00` with
`layout_source: library_match`; **2 gap items exist across 63,007 work items ever written, both
the no-tags arm**. Migration 103 had specified the fix (a 0-1 `layout_match_score`, a 0.5
threshold) in April 2026 and none of it was built. Separately, a peer lane found the classifier
had been rendered a literal `null` where the layout tag list should be (dropped at the
`input_fields` boundary), while being told "coin a tag; an unmatched one will trigger a review" —
so 87% of emitted terms matched nothing and four attractor strings decided every layout. Two
broken links in one loop; theirs made the mess, ours is why nobody saw it. Full evidence:
`bugs_open/445` §8.

## 1. FIRST ACTION — establish liveness. Three things, each one command.

| what | check | expected on 2026-09-03 at handoff |
|---|---|---|
| **Phase 1 Go fix (`76db94fc7`) rolled?** | `kubectl -n ai-persona-system logs -l app=agent-chassis --tail=400 \| grep -m1 'build provenance'` then `git merge-base --is-ancestor 76db94fc7 <stamped sha>`. If the line has scrolled, probe `/proc/1/exe` for the literal `weak_tag_fit` **with both controls** (a literal that must be present, one that must be absent — never `strings`, never a bare log grep). | **MEASURED 2026-09-03: NOT ROLLED.** Chassis `v1.0.1358` pods started 12:06:47Z; `76db94fc7` was committed after that. `/proc/1/exe` probe with both controls (`enforceListingItemSources` present ×2, `layout_match_score` **0**, `weak_tag_fit` **0**, absent control 0) — first by `portfolio_positioning`, then confirmed by this lane. **The fit evidence and the widened signal are in HEAD and NOT in any running binary.** The next fleet roll carries them; nobody should single-service roll for it (releases are whole-fleet, owner runs `make release`). **The cheapest instrument, settled by two independent measurements:** `v1.0.1358` stamps itself `build provenance git_commit d0252fd4dab2a3a583d1cc8eb8e1b26e9c422d85`, and `git merge-base --is-ancestor 76db94fc7 d0252fd4d` → **not an ancestor** (control `587666be8` → ancestor). When a new chassis rolls, read ITS stamp and re-run that one line — no symbol probe needed. |
| **Migration 735 live?** (prompt honesty) | `SELECT position('will trigger a library-growth review' in default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}') FROM agent_definitions WHERE type='domain-research-classifier' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;` | **0** (applied + verified) |
| **Migration 736 live?** (the archetype) | `SELECT is_active, scheme, cardinality(industry_tags) FROM layouts WHERE name='content-hub-tools';` | **t \| light \| 9** (applied + verified) |

Then the two council verdicts, **neither read at handoff**:
```sql
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' IN
 ('34d57f60-7013-4fa4-8106-e8d8e5e29887',   -- Phase 1, commit 76db94fc7
  '39942a14-7d1d-49ff-a2a0-2706098f76f0');  -- Phase 5, migration 736
SELECT body FROM doc_notes WHERE categories ? 'council-gate' AND (body LIKE '%34d57f60%' OR body LIKE '%39942a14%') ORDER BY created_at DESC;
```
Both commits carry `Council-Submitted:`; **write nothing with `Council-Reviewed:` until you have
READ an approved verdict.** A REVISE means the code is already on the shared branch — act on it.
(Migration 735's commit `41dcae9e7` has NO trailer; it was committed before submitting — a
known misstep class, `WRONG_CALLS.md` 2026-09-03. Submit it if you want it credited.)

## 2. What shipped, where

| phase | artefact | state |
|---|---|---|
| 0 | migration **735** — classifier prompt: promise removed, layout names dropped from tag examples, FORM-over-industry sentence added | **applied**, verified at the live row; 734's `input_fields` wiring re-verified intact |
| 1 | commit **`76db94fc7`** — `layoutFit`, `TagCoverage`, `LibraryGap()` with a weak-fit arm at `lmMinTagCoverage = 0.50`, persisted as `lineage.layout_match_score` + `lineage.layout_fit`; `is_scheme_mismatch` and `source` now recorded | **committed; MISSED v1.0.1358 (measured) — inert until the next roll** (§1) |
| 5 | migration **736** — layout `content-hub-tools` | **applied**, 0 sites on it |
| 4 | reachability guard | **partial**: inline DO block in 736 only |
| 2, 3 | `internal/cronchecks`, `cmd/layout-fit-check` | **not started** |

Other commits: `d7aa08c04` (445 §8), `679033bc1` (WRONG_CALLS ×3), lane docs, register
DES-086/087 + DES-037 correction, `41dcae9e7` (735), and the 736 commit.

## 3. Pre-registered predictions — check these, they are how the work gets disconfirmed

1. **The 0.50 threshold.** Coverage bands at design time: 0% / 7–10% / 15–38% / 62–87%; the cut
   sits in the widest empty band. Migration 734 (11:39:14Z) rendered the taxonomy for the first
   time and 735 steers toward form words, so **coverage should rise fleet-wide from the next
   classification**. If new compositions land inside 38–62%, the cut was a 33-site artefact:
   re-derive from the fuller distribution. `lineage.layout_fit.threshold` records the cut per row
   (post-roll) so the comparison stays honest.
2. **copyonline.co.uk is the canary** (portfolio_positioning, remake №5). Their prediction:
   `tool-portal-light`; `magazine-grid` if editorial words lead; and now a third arm,
   `content-hub-tools`, if the classifier emits `content-hub`/`editorial-guides`/… alongside
   `editorial`. They have promised `layout_name`, `lineage.layout_source`, `layout_candidates`, the
   `reasoning` string AND `classification.industry_tags`. It is the first classification ever with
   a real tag list rendered, so it tests 734, 735 and 736 at once.
3. **If the sweep (Phase 3) is built and its numbers do not reproduce** 7 on magazine-grid, 8 on
   tool-portal-light, 9 unreachable, 4 zero-fit (as of 2026-09-03), the census differs — find out
   why before trusting either.

## 4. What is NEXT, in order, and why that order

1. **Read the two verdicts** (§1). Act on any REVISE first — the code is live on the branch.
2. **Prove the roll** (§1 row 1). Then the first post-roll `resolved_composition` row must carry
   `lineage.layout_match_score` — that is the proof the fit evidence is real, not the roll.
3. **Phase 2 — `internal/cronchecks`** (owner decision: build this *before* the sweep, answering
   the open `RFC_024` rather than adding the tenth un-harnessed cron check). ~120 lines: `DB()`,
   `Note()`, exit-code constants (0 clean / 1 findings / **2 refused, never a pass**),
   `SystemSiteID`, the `idx_swi_dedup` status predicate single-sourced with a test that parses
   migration 157, and a **schedule-collision ratchet test** (four collisions exist today: 06:50,
   07:05, 07:25, 07:40). Adopt in the new check only; migrate `cmd/verifier-remit-check` later in
   its own measured commit; touch none of the other 17.
4. **Phase 3 — `cmd/layout-fit-check`**, on the harness. Template `cmd/verifier-remit-check/`
   (the only fleet-wide-finding check). Keys on `sites → style_collections → css_themes →
   layouts`, NOT on `resolved_composition` — `SelectStyleCollectionAction`
   (`v3_site_actions.go:67`) writes no lineage at all, and theme-kit sites never run the matcher.
   **Unit of a finding = cluster `(layout, exact matched-term set)`, not site**; digest the sorted
   term set into the `item_key`; shelf at `system.internal`, `pipeline='maintenance'`,
   `status='deferred'`, **empty `handler_agent`** (migration 690's trigger); never `DO UPDATE`.
   Needs the scorer in Go → **extract `platform/orchestration/actions/layoutmatch`** first
   (24 references in 2 files; keep `resolveLayoutByTags`'s SQL and `Reason` byte-identical).
   Birth-commit registries: `liveItemTypes` + `itemTypesWithoutVerifiers`, `RELEASE_IMAGES` +
   `AGENT_DEPLOY_SERVICES`, `council-scope.sh` + `098`'s `SCOPE_PATHS`. Pick a free minute.
5. **Phase 4 proper** — `assert_layout_reachable(p_layout, p_min_sites)` as a migration-guard
   function + a `scripts/pattern-check.py` rule requiring any `INSERT INTO layouts` migration to
   call it. Both council scope.
6. **The two theme-kit facts** to carry: kit-chosen sites record NO fit evidence (the short-circuit
   returns before the matcher — correct, by design; the sweep is what will score them), and
   `soft-editorial` is now the editorial category's reachable-only-by-kit layout.

## 5. Traps this lane hit — read before you touch anything (details in RUNBOOK / NOTES)

- **Any all-time count of work items must union `site_work_items_archive`**; `site_specs` does
  NOT archive (versions in place under `is_current`). I published a wrong figure to four lanes;
  a peer "independently verified" it by making the same omission.
- **`090` on a code-only symptom needs `SEED_SCOPE`** or it fails after 6 minutes and burns its
  only attempt. The script warns; read to the end.
- **The tree may not build because of someone else's dirty file.** Use
  `scripts/verify-head-builds.sh --with <yours> … --test`; never stash. A pre-existing failure
  on clean HEAD: `TestStylesheetGutted_TokenSetMatchesCanonicalCSSTokens` (`discovery_checks`) —
  not ours.
- **The same-scheme bonus alone makes any same-scheme layout "eligible"** — the scheme-gap arm is
  unreachable while any same-scheme layout exists (that is why `soft-editorial` is the permanent
  0.50 runner-up). A test fixture assuming otherwise fails.
- **A mutation can miss a test by never reaching it** (mutation (ii) killed one test, not two:
  the zero-overlap case never evaluates the mutated expression). Record what happened, not what
  you predicted.
- **The council submission schema:** `.plan` is an object `{summary, edits[≤8], grounded_in[]}`;
  `operation ∈ modify|add|remove|config_change`. **Submit first, commit naming the correlation.**
- **Prompt text is not the prompt.** Read `llm_call_log.prompt_rendered`, anchored on a LONG
  phrase — `'Current library tags'` alone matches the JSON schema example earlier in the same
  prompt and returns the wrong region.
- **A subagent report told me 689 was unapplied** — stale register entry. Verify at the table
  (`to_regclass`), not at the register.

## 6. Peers and what is owed

| lane | state |
|---|---|
| `portfolio_positioning` | Found the `null` tag list (their 734, applied 11:39:14Z); their 734 council round drew a **REVISE** (editquality: is `positioning_register` referenced in `prompt_template`, not just `input_fields`?) — told them. They owe the copyonline five-field read. Their seventeen-remake CONTRIB is the forward-looking population. |
| `theme kits` | Waiting on nothing now; told them the archetype is live and about `soft-editorial`'s kit-only reachability. Their seed set has a self-declared defect (2 of 4 kits redundant with the matcher). |
| `designblog.co.uk` | Churn in flight (migration 726, GTM chrome wave). **Nothing may queue work against designblog without telling them first.** Their header pin is HELD (alternative headers unpopulatable AND semantically wrong for a blog). |
| `site design planner` | Filed 445; confirmed it free; owns the composition resolvers; touched `resolve_composition_layout_action.go` last at `bd8e45aba`. |

## 7. Files

`docs/agent_docs/docs024_key_docs_latest/bugfix_445_layout_fit/` — `PLAN_2026-09-03_layout_fit.md`,
`RUNBOOK_layout_fit.md` (9 commands with gotchas), `NOTES_layout_fit.md` (append-only, missteps),
`README_where_we_are.md` (owner's log), two `COUNCIL_SUBMISSION_*.json`, this file.
Migrations: `sql_for_agents/735_*`, `736_*` (+ `_ROLLBACK` each; 736's DEACTIVATES, never deletes).
Scratch (session-local, will not survive): `scratchpad/score.py` (the validated scorer),
`simulate2.py` (candidate tag sets), `sitetags.txt`, `layouts_live19.txt`, `prompt_after.txt`.
**The scorer must become Go (`layoutmatch`) before it can be relied on by anyone else.**

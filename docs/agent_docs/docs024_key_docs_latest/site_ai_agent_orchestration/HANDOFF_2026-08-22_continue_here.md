# HANDOFF — ai-agent-orchestration.com. START HERE. Written 2026-08-22 ~12:40Z.

**Supersedes `HANDOFF_2026-08-18_continue_here.md`** for current state. Read that one second, and
treat its §4 (images) and §5 (carousels) as **superseded plans, not just stale figures** — both
described work that would not have produced the result. Why, below.

> ## ✅ NOTHING IS BLOCKED. Contrast, carousels and images are DONE and verified live. Only `pricing` remains.
>
> | ask | state |
> |---|---|
> | **contrast** | **32 → 8 firm failures.** `index`, `about`, `services` at **0**. All 8 survivors are on `pricing`, unreachable by re-render |
> | **carousels** | ✅ **LIVE** on `index` + `enterprise-reference-deployment` (migration `559`). Opt-in, default OFF; other two sites verified untouched |
> | **images** | ✅ **10/10 card images live at HTTP 200** (migrations/items 2026-08-22). 0 broken images in the audit |
> | **the "196 agents" claim** | ✅ **Fixed at source** (migration `557`) — but NOT by writing 196; see §1, the number is the wrong lever |
> | `pricing` rebuild | **STILL BLOCKED, one claim away.** Two of its three blockers are cleared; §3 |
> | `bugs_open/364` | Clock-time false positive in the claims layer. **Fixed + tested + council-submitted; INERT until the next fleet roll** |
>
> **Everything below was measured on 2026-08-22.** Counts carry their date per the owner ruling of
> that morning — re-run them, do not quote them.

---

## 1. The claim fix — why "196" was the wrong instruction to follow literally

The owner said *"update the claim to 196 agents at source"*. **Writing 196 would have reproduced
the bug with a different number.** Do not undo this reasoning:

- The agent counts are **live SQL**, re-run on every evidence refresh: 175/174 (07-26) → 196
  (08-19) → **199** (08-22). Any literal in a spec is wrong within days.
- **"170+" was never false.** Both facts carry `tolerance:"gte"`; `numberSupported` accepts any
  value ≤ the registered one (`claims.go:1007`). 170 ≤ 199.
- **The rejection was the CONTEXT gate** (`claims.go:990-1001`): a fact is skipped entirely unless
  one of its `context_terms` appears near the number. "a registry of 170+ agents" matched none of
  `agent definition` / `specialised ai agent` / `agents in the registry` / `ai agents`.

**Root cause: `evidence_base.writer_block` instructed the writer to produce a phrase its own facts
could not validate.** Fixed by `557_aiao_evidence_base_stops_mandating_a_phrase_its_own_facts_cannot_validate.sql`
— no literal counts left in the block (all figures delegated to the facts list), plus 5 phrase-based
`context_terms`. **Fact VALUES untouched**; they belong to `refresh_evidence_base_action`.

⚠ **Do NOT "simplify" the added context terms to the bare word `agents`.** It would make every
"N agents" sentence eligible for a `gte` fact of 199. Proven the same day: the index rebuild was
correctly refused for *"…Threatens a **40**-Agent Pipeline"*, a client-scale claim this site does
not measure. The bare word would have certified it.

### ⚠ The open follow-up: `writer_block_managed` is NOT set, and setting it today loses controls

`refresh_evidence_base_action` regenerates `writer_block` from each fact's `writer_line` with
`{value}` interpolated — but only where `writer_block_managed: true` (`:474`). This site never opted
in, which is why the block froze on 2026-07-27 while the facts refreshed.

**Do not just set the flag.** `composeWriterBlock` (`:996`) builds the block from `writer_line`s and
`allowed_entities` and **nothing else**, so it would silently delete: both NEVER-write bans, the
whole NOT-TRACKED/NEVER-STATE list, and the two "DO NOT state a figure" cautions (orchestrations —
rolling window; work items — a reaped ledger that FALLS). The bans survive as `banned_claims`
regexes so enforcement holds; **prevention** does not, and the two cautions are enforced nowhere.
**The prerequisite is `composeWriterBlock` learning to carry negative guidance.** Until then this
block stays hand-written and `557` keeps it honest.

## 2. Carousels — SHIPPED, and the register was never the delivery mechanism

⚠ **`HANDOFF_2026-08-18` §5 said the work was "APPROVE + BIND, not design". That plan produces
nothing.** [MEASURED 2026-08-22] the experience register is a **specification and verification**
system: only **3** Go files touch it — `write_experience_pattern` (records a contract),
`bind_site_experience` (records which page), and `verify_site_experience`, whose header says it
*"run[s] a bound fork's criteria against the deployed page"*. **Nothing renders from
`site_experiences`.** Binding would have created a rule demanding a carousel on a page that had
none. The trigger script says the rest: *"nothing in this lane can write `status='approved'`.
Applying a verdict is a separate action that does not exist yet, deliberately."*
Register unchanged: **11** patterns, **0** approved, council **0** runs ever (2026-08-22).

**Shipped as `559_case_studies_grid_optional_scroll_snap_carousel.sql`**, implementing the
`arrow-and-swipe-card-carousel` contract (which is genuinely good — follow it, don't reinvent):
native scroll-snap so it works with **no JS**; JS adds arrows only; reduced-motion honoured; init
idempotent against a double include and a re-init. **Auto-advance deliberately NOT built** — the
contract makes it conditional and it is the clause that drags in IntersectionObserver, hover/focus
pause and re-derive-after-swipe. Nothing rotates, so none of those failure modes exists.

⚠ **Controls hide on OVERFLOW, not on card count.** The component ships a category filter that hides
cards with `display:none`; a count taken at init stays wrong after filtering to one card, leaving
inert arrows. Visibility derives from `scrollWidth > clientWidth`, re-checked on scroll, resize and
a `MutationObserver` on the cards' `style`.

**Opt-in, default OFF** (owner ruling 2026-08-02). Component is on **4** pages / **3** sites
(2026-08-22). Verified: both aiao pages carry it; `finetuning.uk` and `leopardessconsulting.co.uk`
return **zero** carousel markers. They may opt in whenever they choose — that is their call.

## 3. `pricing` — the ONLY contrast work left, and it is now one claim away

8 firm failures, all here. 5/5 components `content_data IS NULL`, last rendered 2026-04-13, so no
re-render can reach it; it closes via a framework rebuild or not at all.

**Three blockers found; two cleared:**

| blocker | state |
|---|---|
| section shrink floor (CTA 483→213 chars) | ✅ **not a blocker** — the generator is non-deterministic and the 08-17 draw was bad. A re-run produced **489** visible chars, *longer* than the live copy. Lowering `section_shrink_floor` would have been wrong |
| `unregistered_number "170"` | ✅ **cleared by `557`** |
| `unregistered_number "2"` from "2am" | ⚠ **`bugs_open/364`, fixed in Go, INERT until the next fleet roll** |

⚠ **The live copy the shrink floor was defending is itself defective** — it serves
`[LLM Provider Cost Comparison Calculator](/tools/…)` as **literal unrendered markdown**. Contributed
to `copy_quality_two_stage`.

**Next attempt should wait for the roll** (check the build stamp, don't infer). If it still refuses,
read the issues — `agent_error_log`, `error_code='CONTENT_VALIDATION_BLOCKER_DETAIL'`, which stores
the structured list precisely so pod logs are not needed (retention here is ~4 minutes).

## 4. Live traps confirmed or found on 2026-08-22

- ⚠ **Generating N images for one site SERIALISES behind N page rebuilds.** Each landed asset makes
  `image-build-handler` file a `needs_page`, and one `claimed` row of **any** type holds the
  per-site mutex (`029`). My run stalled 5-of-9 for ~9 min behind a rerender the image pipeline
  filed itself. It drains; **do not poke it** — a takeover resets the 4-hour reaper clock.
- ⚠ **Those `needs_page` rebuilds fail at `validate_content` and write nothing**, so the image
  pipeline's own propagation does not work here. Use a **page-scoped `template_changed`
  `page_rerender`** (the MERGE path, RUNBOOK R8). That is what actually shipped the images.
- ⚠ **A full page REBUILD would drop `carousel_enabled` and the ten image URLs** — they live in
  `content_data`, which a rebuild regenerates (a rerender merges). This nearly fired on 08-22 and
  was stopped only because an unrelated claims error refused the write. **If the carousel or the
  images vanish after a rebuild, re-set the keys; do not re-debug the CSS.** There is no durable
  per-site presentation flag on this seam — `site_experiences` would be the right home and nothing
  reads it.
- ⚠ **The served asset path is DERIVED, not stored**: `DeployedWebPath` → `AssetKeyFilename`
  (underscores → dashes) + the extension from `ImagePurposes[purpose]`. `content_hero` ⇒ **`.jpg`**.
  The old card URLs ended `.png` and could never have resolved whatever was generated.
- ⚠ **`image-url-404-handler` / `image-source-unsatisfiable-handler` TRIAGE, they do not generate.**
  Real generation is `needs_imagery` → `image-build-handler`. Unchanged from the 08-18 handoff and
  still the trap it was.
- ⚠ **The 17 parked `contrast_failure` items cannot drain yet** — the site's render audit has not
  run here since 2026-08-10. Their presence is not evidence anything failed. `bugs_open/296` owns
  this; this site is a clean natural experiment for it now that index/about measure 0 and pricing 8.

## 5. Next actions, cheapest first

1. **Wait for the fleet roll, then retry `pricing`.** `364`'s fix must be live first — verify at the
   build stamp, not by inference. One page-scoped `needs_page` with
   `spec.reason='content_data_backfill'`; it writes nothing if it refuses.
2. **Tell `finetuning.uk` and `leopardessconsulting.co.uk` the carousel exists** and is one
   `content_data` key away. Their call, not ours (owner ruling 2026-07-29 §3: shared-seam consumers
   must be told, not merely measured).
3. **`composeWriterBlock` must carry negative guidance** before any site sets
   `writer_block_managed`. Small, and it unlocks self-maintaining evidence blocks fleet-wide.
4. **The literal-markdown defect** in the pricing CTA — with `copy_quality_two_stage`.
5. Optional: bind the carousel in the experience register now that a real carousel exists, so
   `verify_site_experience` can hold it to its contract. **Value is verification, not delivery** —
   and nothing can mark the pattern `approved`, so expect a `proposed` fork.

## 6. What shipped, with its rollback

| migration | what | rollback |
|---|---|---|
| `469` | departments-grid + leadership-team consume site tokens (24 contrast failures) | byte-exact from `migration_backups` |
| `557` | evidence_base stops mandating an unvalidatable phrase | byte-exact restore |
| `559` | opt-in scroll-snap carousel on case-studies-grid | byte-exact + clears the flag |
| `560` | binds 10 card slots to 9 stable `/assets/images/*.jpg` | byte-exact |
| `bugs_open/364` | `am`/`pm` added to the number-scan exclusions (Go) | revert the commit |

All applied by hand with `psql -f` and recorded via `--record-only` — **never `--apply`**, which
would sweep other lanes' pending files (17 of them at the time). Each was rehearsed with
`COMMIT`→`ROLLBACK` first; that rehearsal caught a real defect in **two** of the four.

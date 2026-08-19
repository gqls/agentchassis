# HANDOFF 2026-08-19 — meta descriptions (`bugs_open/320`): LIVE and PROVEN, with three things left

**Read this first if you are picking this up cold.** Everything below was verified at the
artefact; where a claim is unproven it says so.

---

## 1. The one-paragraph version

`pages.meta_description` is the sentence Google prints under a page title, and on this
estate it is also each blog card's excerpt. **407 of 731 live pages had none.** Two live
mechanisms caused it: the planner was never asked for one, and an unguarded upsert clause
blanked the ones that existed. Both are fixed and live. A backfiller now exists, is
proven, and has filled 26 pages so far. **355 pages remain empty because only two sites
have been run.**

## 2. State, and how each was proven

| thing | state | proof |
|---|---|---|
| Chassis | **`v1.0.1315`**, revision `590ca3a20cca…` | pod `imageID` **==** local `RepoDigests` (a real build, not a same-tag rebuild); all four commits ancestors; **control: `HEAD` correctly NOT an ancestor** |
| M2 guard, **four** write paths | **LIVE** | `grep -rn 'meta_description = EXCLUDED' --include='*.go' . \| grep -v COALESCE` → 0 live sites |
| M1 — planner asked (mig `485`) | **LIVE** since 08-19 | config, live on apply; chain verified: `write_site_plan_action.go:535` reads the exact key `485` adds |
| `save_page_meta_description` (SEO-004) | **LIVE** | binary probe PRESENT, positive + negative controls both behaved |
| Copy gates (owner's condition) | **LIVE, in the action** | 3 tests, mutations RUN: delete the gate call → 3 fail; make the unreadable branch pass → 1 fails |
| `meta-description-backfiller` (mig `488` + `493`) | **LIVE and PROVEN** | 26 pages written; idempotence proven at row level |
| Council | **APPROVED** round 2 | corr `46734ae9-91c5-47d6-9a8a-4cd1fa213d21`; round 1 REVISE found a real defect |

**Fleet: 407/731 empty → 381/736.**

## 3. WHAT TO DO NEXT, in order

### (a) Measure whether `bugs_open/309` is actually unblocked — do NOT assume it
All five pages blocking `309` now have descriptions. **That is not the same as `309` being
fixed**, and the arithmetic says so:

`309` §9 measured the old blog-listing slot at **2,478** chars of visible text, with a
**50% shrink floor = 1,239**, and projected the rebuilt slot at ~**1,818** *assuming ~157
chars per description*. `[MEASURED 2026-08-19]` **the descriptions came in at a mean of
102 chars (range 65-177)**, well under the 120-155 the prompt asks for. So the projection
is materially lower and the guard may still refuse.

**The only way to know is to dispatch the rerender** (envelope and the `check_rerender_mode`
gotcha are in `bugs_open/309` §9 — the reason MUST be one it recognises, `template_changed`
is what was used). Expected on success: 8 cards, 2 anchors each, archived guide absent.
**Verify at the SERVED page**, never the stored HTML.

If it still refuses, the honest options are to improve the prompt's length adherence (the
cheap one) — **not** to lower `section_shrink_floor`, which is step config and therefore
fleet-wide.

### (b) Run the remaining sites
```bash
./scripts/backfill-meta-descriptions.sh <domain>      # one site, safe to re-run
```
Reads the LIVE agent config and sends it inline, so it cannot drift from the seeded row.
`overwrite_existing` defaults false, so it fills blanks and never replaces copy — proven
at row level (a repeat run touched only the still-blank page). Sites with the most empty
pages: `webdesign.co.uk` (78), `loancalculator.co.uk` (43), `finetuning.uk` (32),
`ai-agent-orchestration.com` (27), `loanandmortgagecalculator.co.uk` (26).

⚠ **Check the artefact, not the status.** The first canary reported `COMPLETED` and wrote
nothing.

### (c) Consider the description length
Mean 102 against a 120-155 ask. Harmless for SEO, load-bearing for (a). If (a) fails, this
is the first lever.

## 4. Traps this lane paid for — do not re-pay

- **A `COMPLETED` orchestration that wrote nothing.** `output_format: "array"` returns a
  BARE ARRAY, so a gate reading `X.count > 0` resolves to nothing and silently routes to
  else. `.count` exists only under `output_format: "object"`
  (`database_actions.go:129-145`). **This is `bugs_open/313`, and it arrived here because I
  copied `internal-linker` — the agent 313 was filed against. Copying a live agent copies
  its bugs.** ≥8 other live conditions share the shape; that sweep belongs to `313`.
- **`check_voice_tells` cannot see this column.** It scans `page_components.rendered_html`;
  `pages.meta_description` is invisible to it and to every `rendered_html` census. Wiring
  it would have produced a confident pass over text it never examined. The reusable
  text-level entry points are `VoiceGate.ScanVoice([]string, longForm)` and
  `checkBannedClaims([]string, …)`.
- **`content_sample` from raw markup is mostly CSS.** A model handed a stylesheet will
  still write you a fluent, wrong sentence, and no copy gate catches that.
- **`--record-only` REFUSES an uppercase-suffixed sidecar.** A file that is `_HOLD`,
  hand-applied, and left named `_HOLD` ends up applied with **no ledger row**. Rename it
  back the moment the hold is satisfied.
- **`display_name` and `category` on `agent_definitions` are NOT NULL with no default.**
- **A wrong claim I propagated into four places:** the new guard does NOT "match
  `nav_label`". `nav_label` is `COALESCE(NULLIF(pages.…,''), EXCLUDED.…)` — **existing
  wins**; `meta_description` is the mirror image — **incoming wins unless blank**. Both
  deliberate, deliberately different. **Do not unify them.**
- **When two of your own measurements disagree, neither is evidence** until you can say
  which is wrong and why (I got 1 vs 314 visible chars on the same unchanged page and
  nearly filed the wrong one).

## 5. Where everything lives

- `bugs_open/320` — the case; **§11 is the owner ruling**, §12 the live results.
- `bugs_open/309` — blocked-on-this; its §10 carries my correction of its own explanation.
- `bugs_open/313` — the `.count`/`output_format` class this lane tripped over.
- This directory: `PLAN`, `NOTES` (append-only, newest last — the missteps are the point),
  `RUNBOOK`, `README_where_we_are` (owner's prose log), `SUMMARY_2026-08-19`, the two
  council submissions.
- Migrations: `485` (planner), `488` (the agent), `493` (the canary fixes) — all applied
  and ledger-recorded, each with a ROLLBACK sidecar.
- Code: `save_page_meta_description_action.go` (+ two test files),
  `site_db_actions.go`/`apply_adoption_plan_action.go`/`adopt_verbatim.go`/
  `cmd/webdesignport/import.go` (the four guards).
- Register **SEO-004** (`register/seo.md`) + its index row.
- `scripts/backfill-meta-descriptions.sh`.

## 6. The owner's standing instructions on this lane

1. **Backfill authorised FLEET-WIDE, review pass WAIVED** (`320` §11, 2026-08-19).
2. **Condition of that waiver:** the summaries go through the copy guidance and checks so
   they don't sound like AI. Guidance is in the prompt; checks are in the action, before
   the write, where a workflow author cannot forget them.
3. **The framework writes the content, not a session** (2026-08-06). If a generator does
   not exist, that is the finding to report — it is why this lane exists.
4. **`309`: wait for the writer/replan** — no regeneration of the five articles, no
   shrink-floor change.

# 220 — `unbuilt_internal_link` dispatch rebuilds the CONTAINING page, not the target; marks itself `complete`; re-detects next run — a convergence-free green loop

**Filed 2026-08-08** by the `bugs_open/206` lane while answering the owner's question "would
the improvement loop have picked these problems up?". Diagnosed first-hand rather than via the
090 loop — declaring the substitute per the 2026-07-31 ruling: every link in the causal chain
below is quoted from the live system (the live `agent_definitions` config, the work item's own
`spec`/`page_id`/`result` columns, the `pages` table), not inferred; and a live re-demonstration
was fired the same day (correlation `867d6054-3f8f-4b11-9352-b29cecd9aaaa`, one-shot
improvement-loop over vetcomparison.uk — outcome: **REPRODUCED, see the dated section
at the bottom**).

## One-line mechanism

The phantom-links check deliberately files an `unbuilt_internal_link` work item against the
**target** page (the never-deployed one) via the item's `page_id` column — but
`build-dispatch-loop`'s `call_handler` step maps `page_name` from `current_item.spec.page_name`,
which for this item type is the page **containing** the link, and never reads `page_id`. The
handler then rebuilds/redeploys the container, returns success, and `mark_complete` runs
unconditionally. The wrong result looks exactly like the right one.

## Evidence (all read live 2026-08-08)

Item `8abc9f8d-c555-4378-bed4-dac8042025b2` (site vetcomparison.uk, minted by discovery
2026-08-02, one of three that day):

- `item_key` = `unbuilt_internal_link:page_component:index:info-card-grid:/directory/index.html`
- `page_id` column = `11326da0-df30-4378-a601-00c2812a891e` → **pages.name = 'directory-index'**
  (the target — correctly filed by the check).
- `spec.page_name` = `"index"` (the container), `spec.target_page_id` = the directory-index id,
  and `spec.fix` says verbatim: *"Do NOT rebuild the linking page — its href is correct."*
- `status` = `complete`, `attempt_count` = 0, `error` NULL, `claimed_by` = build-dispatch-loop,
  completed 2026-08-02 10:58:44Z.
- `result` (truncated): `deploy_result … "files": ["/index.html", "/tools/assets/latest-news.js"]
  … "success": true, "commit_message": "Rerender: "` — **the homepage was deployed, not the
  target.** Sibling item `b99dea13` (about → /directory/index.html) likewise deployed
  `/about.html`.
- `pages` row for directory-index today: `build_status='planned'`, `deployed_at IS NULL`,
  `sections = []` — the link is still a live 404.

The dispatching config, from the live `build-dispatch-loop` `agent_definitions` row
(`default_config->workflow->steps->process_item`, sub-workflow `call_handler.input_mapping`):

```json
"page_name?": "current_item.spec.page_name",
```

No key in the mapping reads `current_item.page_id`. The check's own routing intent
(`check_phantom_internal_links.go`, the `unbuilt_internal_link` arm: `actionablePageID =
f.TargetPageID` … "For an unbuilt link … the item is filed against the TARGET — rebuilding the
container would re-emit the same correct href and resolve nothing") is defeated by the generic
mapping. A doc comment / spec-prose instruction is not an enforcement mechanism — the dispatcher
never reads either.

## Why it reads green forever (three stacked absences)

1. **Wrong page**: the mapping above — the container rebuilds, the target stays unbuilt.
2. **No verification**: `call_handler.next_step = mark_complete`; `complete_work_item` runs on
   any non-error handler result. Nothing re-checks the actionable page's `deployed_at`.
3. **No convergence brake**: `complete` is terminal for the dedup index, so the next discovery
   pass re-mints the same `item_key` and repeats 1–2. Each cycle spends a full page build +
   deploy on the wrong page and reports success. (Distinct from `bugs_open/083`
   detected-findings-never-reach-a-handler: that gap is items never promoted because the sweep
   is off; this one fires precisely when the loop DOES run.)

Note even fixing (1) alone would not have built this page before 2026-08-08: bare
`page-build-handler` has no layout-filling step (`load_spec_sections` → `plan_sections` →
`mark_no_ready_sections` no-op for a plan-less page) — that capability gap was `bugs_open/206`,
fixed by `directory-build-handler`/`ensure_page_section_layout` (live v1.0.1264 + migrations
325/326). With 206's fix live, correcting THIS bug's routing would still dispatch
`page-build-handler`, not `directory-build-handler` — the check's `HandlerAgent` is hardcoded
and predates `load_work_item_actions.go`'s `availableBuilders` map gaining `entity-directory`.

## Fix candidates (ordered by what closes the door)

1. **Make the dispatcher honour the item's own `page_id`**: map e.g.
   `"page_id?": "current_item.page_id"` and have `page-build-handler`'s `load_page_record`
   prefer it over `page_name` — the column exists precisely to name the actionable page, and
   every check that files cross-page items benefits (not just this one). Config +
   possibly one Go read; blast radius = every `call_handler` dispatch, so measure which live
   item types carry a `page_id` differing from `spec.page_name`'s page before shipping.
2. **Check-side workaround**: set `spec.page_name` to the TARGET's name in the
   `unbuilt_internal_link` arm (the check already resolves the target row). One-line Go change,
   closes only this item type, leaves the mapping trap armed for the next cross-page check.
3. **Verification**: `mark_complete` (or a step before it) asserts the actionable page moved —
   for this item type, `deployed_at IS NOT NULL` on `page_id`. Turns silent wrong-fixes into
   loud failures fleet-wide; larger design question (what "moved" means per item type).
4. Route `unbuilt_internal_link` through `availableBuilders` rather than hardcoding
   `page-build-handler`, so an entity-directory target reaches `directory-build-handler`.

(2) is the cheap immediate stop; (1)+(3) are the class fixes; per this repo's own rule a
one-off deletion/patch is not a class fix — the pair is what makes the bad state
unrepresentable.

## How to verify a fix

Re-run the one-shot improvement loop over vetcomparison.uk (or any site with a live link to a
never-deployed page). Success = the minted item's dispatch builds the page named by the item's
`page_id` (target), `curl -sI` on the target URL returns 200 with fresh `last-modified`, and the
item completes with a result naming the target's file — not the container's.

## Related

- `bugs_open/206` — the capability gap this defect masked; its lane found this while measuring
  the improvement loop's actual behaviour. Fix live 2026-08-08.
- `bugs_open/083` (detected-findings slug) — the delivery gap upstream (sweep off, items park at
  `detected`); this bug is its complement when the loop runs.
- `bugs_closed/049` — created the `unbuilt_internal_link` class; closed on detection + nav-drop
  halves. The end-to-end remediation path was never exercised at close time (its own §
  "phantom_internal_link … fixed zero times, ever" table shows the type had never fired).
- Memory/lesson: "a `complete` work item is not a repaired artefact"; "a doc comment is not an
  enforcement mechanism".

## REPRODUCED 2026-08-08 15:19Z — same session, platform fully carrying 206's fix

The one-shot improvement-loop run (corr `867d6054`) re-minted 7 `unbuilt_internal_link`
items (dedup did not block — the 08-02 items are terminal, absence 3 above). The first
two dispatched before this lane intervened:

| item | finding (container → target) | what the dispatch deployed | target after |
|---|---|---|---|
| `3f066b90` | guide-independent-strategy → /directory/index.html | `/guides/independent-strategy/index.html` | still `deployed_at NULL` |
| `4ba1d4dd` | how-it-works → /entities/practice.html | `/how-it-works.html` | still `deployed_at NULL` |

Both `complete`, no error. This is with `directory-build-handler` live in
`agent_definitions` (migrations 325/326/336 applied) — confirming the check's
hardcoded `HandlerAgent: "page-build-handler"` + the dispatcher's `spec.page_name`
mapping bypass the new capability entirely, exactly as the "Note" paragraph above
predicts. The remaining 5 re-minted items were **cancelled by the 206 lane** (error
text on each points here) because the correct remediation for their targets — the
`needs_page` rows routed at `directory-build-handler` — was already queued; without
that intervention each would have spent a further multi-minute LLM rebuild on the
wrong page. A companion `empty_internal_href` item (`5cc5c24b`, tool page) dispatched
and no-op'd to `needs_human_review` ("no sections ready to build") — the parallel
failure shape for an item type whose container IS the right page but has no plan
sections.

## CONTRIB 2026-08-08 (116 lane) — fix candidate 1's blast radius, MEASURED: `unbuilt_internal_link` is the only type affected

Candidate 1 asks for a measurement before shipping — *"measure which live item types
carry a `page_id` differing from `spec.page_name`'s page"*. Run 2026-08-08 ~17:10Z by
the `bugs_open/116` lane, which needed the same number to decide whether the first
supervised improvement-loop run (owner ruling D3, 116:254) could safely be fired at
leopardessconsulting's 61 parked findings. Contributed rather than forked — this bug
is the 206 lane's.

**Method.** Left-join each work item to (a) the page its `page_id` names and (b) the
page `spec->>'page_name'` resolves to *within the same site*, then count agreement.
The dispatcher reads only (b); the check files against (a).

**Result over all live items (`status IN ('detected','triaged','approved')`), 24 types, 168 items: DISAGREE = 0 for every type.** The types split three ways, and only the
third is even capable of exhibiting this defect:

| shape | types | can it disagree? |
|---|---|---|
| `page_id IS NULL` (dispatcher's `page_name` is the only page signal) | 16 types incl. `needs_internal_links` 18, `undeployed_asset` 20, `needs_rerender` 11 | no — nothing to disagree with |
| `page_id` present and **equal** to the named page | `page_rerender` 66, `phantom_internal_link` 12, `empty_internal_href` 4, `empty_section` 3, `literal_markdown` 1, `improve_tool` 1 | no — filed against the container, which IS the actionable page |
| `page_id` present and **different** | **`unbuilt_internal_link` only** | **yes — this bug** |

**The zero is disconfirmable, and was disconfirmed on purpose before being trusted.**
Positive control: re-run unfiltered by status, the same query flags **13/13**
`unbuilt_internal_link` rows as DISAGREE — including both rows this file names
(`8abc9f8d` index→directory-index, `b99dea13` about→directory-index) plus 11 more
across the fleet (`10b4fc72`, `49f0e189`, `6d1cb353`, `5686a320`, `bdf76041`,
`4ba1d4dd`, `2ef4eb51`, `3f066b90`, `240e4020`, `46f732a4`, `b20fa738`). All 13 are
`complete` or `cancelled`, which is why they fall outside the live filter — **the live
zero is a real zero, not the empty-set kind.**

**What this means for the fix.** Candidate 1 (`"page_id?": "current_item.page_id"` +
`load_page_record` preferring it) has **no live behaviour change outside
`unbuilt_internal_link`** — every other live type either supplies no `page_id` or
supplies one that already agrees. That is the narrow blast radius candidate 1 hoped
for, now measured rather than assumed. It does **not** speak to types that are
currently absent from the queue; a future cross-page check would newly depend on the
mapping, which is the argument for pairing it with candidate 3 rather than shipping
alone. Re-run before shipping — the census moves daily.

**Caveat this lane could not close:** the measurement covers items that already exist.
It says nothing about what a *discovery pass* mints mid-run, and this file's own
REPRODUCED section shows a run re-minting 7. The mint condition is a live link to a
never-deployed page.

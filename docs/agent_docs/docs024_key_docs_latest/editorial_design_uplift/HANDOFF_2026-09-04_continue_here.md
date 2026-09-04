# HANDOFF 2026-09-04 — editorial_design_uplift, continue here

**Supersedes `HANDOFF_2026-09-03_continue_here.md` for STATE.** That file stays as the detailed
record — its §2 (the transaction that did not exist), §4a (the binary probe pair) and §7 (the
finetuning chain) are worth reading once, and its §0 environment figures are now stale. **Ordered by
what you need, not by when it happened.**

**Branch:** `087_towards_multiple_domains`. Evidence: `NOTES_editorial_design_uplift.md`, 2026-09-03
and 09-04 entries. Owner-facing prose: `README_where_we_are.md`.

---

## 1. Where the lane is, in one paragraph

**035 P1 direction 2 is DONE, council-approved, and LIVE in the fleet.** A child edit now recomposes
the ancestors that embed it, guarded and stamped, with six mutation-proved tests. **The remaining work
is P1's read path** — `walkComponentHierarchy` still has no production caller, so composition is
built, reachable and unusable. The imagery half of the lane is parked exactly where 09-02 left it and
should not be restarted at the component layer. Two other lanes are mid-flight on work this lane
contributed to but does not own.

---

## 2. Environment — re-measured 2026-09-04 ~15:50Z, not carried forward

| fact | value |
|---|---|
| chassis running | **`v1.0.1360`** (pods up 2026-09-03T22:06Z) |
| **roll in flight at write time** | **`v1.0.1361`**, cut at **`06c0b18f2`** — 393 commits ahead of 1360, 18 touching Go |
| migration dry-run (owed after the 1360 roll, **DONE**) | `Pending (185)` · **38** `LIKELY ALREADY APPLIED` · 41 inconclusive |
| `page_components` carrying a `parent_instance_id` | **0 of 3,475** |
| active components declaring a `slots` block | **0 of 505** |

⚠ **A further dry-run is owed once `v1.0.1361` lands.** It takes >5 minutes and prints nothing until
it finishes; run it unpiped in background. The already-applied column is `bugs_open/426`'s figure and
is climbing (34 → 36 → 37 → 38 across 09-03/04).

⚠ **RETIRED, do not carry it forward:** the 09-03 handoff's §5 warned that HEAD was RED in
`discovery_checks`. **Re-checked 2026-09-04 and it is GREEN** — the 458 lane fixed it. Scope your
`go test` however you like.

---

## 3. What is DONE and live — with the evidence, so nobody re-derives it

**035 P1 direction 2** (`recomposeAncestors`): wired, guarded with the tombstone and lock predicates,
stamped as `action:recompose_ancestors`, fail-closed on an unreportable row count, six mutations
recorded in the test file header and all killed. Council `cab931b1-8b45-461e-8a37-0dbdfa6aa928`
**APPROVED**, orchestration reached `complete_approved: COMPLETED`, all six advisories actioned or
adjudicated in `3ba94508c`'s message.

**Live in the fleet, confirmed at the artefact across three images:**

| symbol | v1.0.1358 | v1.0.1359 | v1.0.1360 |
|---|---|---|---|
| `PlanSectionsAction` | PRESENT | PRESENT | PRESENT |
| `zzzInventedControl_NotInAnyBinary` | absent | absent | absent |
| `recomposeAncestors` | **absent** | **PRESENT** | **PRESENT** |

The 1358→1359 pair is the discriminating one: one bit moved, both controls held. **Live means
REACHABLE, not exercised** — 0 of 3,475 rows are parented, so the read path is what would exercise it.

**Also done:** `deriveRenderMode`'s third value, the membership helpers, direction 1 (refuse to render
a composition parent alone), the flat-pass extraction. **`check_render_mode` routing is REFUTED, not
deferred** (`5542a76d6`) — nothing reads `render_mode`, so P1's routing story cannot work as 035 wrote
it.

---

## 4. What to do next, in order

1. **THE READ PATH — the remaining core of P1.** `walkComponentHierarchy` has no production caller, so
   a row that opted in today would still render flat. Its own council round.
2. **Hazard 6.9's filter MUST land INSIDE that change.** `loadStoredSections` selects
   `COALESCE(parent_instance_id::text,'')` but its WHERE is only `page_id = $1 AND <not removed>`, so
   every row returns flat. The moment the walk renders children in a nested pass, children are in BOTH
   lists, every later section's `NextOccurrence` shifts, and per-section figures attach to the WRONG
   sections — rendering, deploying and looking correct. Read 035 §6.9 first; it also carries the
   `MergeLockedPageSlots` inverted-polarity trap for any plan-vs-live guard.
3. **TWO COUNCIL CONDITIONS BIND THAT SUBMISSION**, both named by seats as follow-ups:
   - **`stale_ancestor_slots` must be wired into something that FILES.** It is computed and consumed
     by nothing — `bug_historian` called that this platform's most repeated failure shape
     (`bugs_open/083`, `/071`). Harmless while inert; the incident the moment composed pages are real.
   - **Probe the deployed binary after the roll that carries it**, controls on opposite sides. §3's
     table is the worked example.
4. **Then the register entry and the live canary** — P1's actual acceptance.
5. **`news-listing`** — same defect `article-body` had, still unwritten, still behind the 09-02
   handoff's §3 question.
6. **Do NOT restart imagery work at the component layer.** The live question is the planner's page
   composition and it belongs to `bugs_open/114`.

---

## 5. Open commitments to other lanes — what this lane owes, and what it has refused

| to | what | state |
|---|---|---|
| `finetuning_uk_service` | carousel component constraint spec | **DELIVERED** — `SPEC_2026-09-04_carousel_component_constraints.md` (`c2cc6fb55`) |
| `finetuning_uk_service` | checking the creator's output against that spec's §6 | **offered, not yet asked for** |
| `bugs_open/114` | the producer that drops a hero's `content_data` key | **CONFIRMED by that lane** — heroes ARE in the DELETE set (`AgentWritableSQLFor` gates on the lock alone; 122 of 123 violating rows satisfy it). `save_page_sections`' page-wide DELETE + reinsert is the mechanism behind the 664 decay |
| `bugs_open/114` | a COMPLETE census of wholesale `page_components.content_data` writers | **DELIVERED** — `bugfix_114_imagery_wiring/CONTRIB_2026-09-04_…_wholesale_page_components_writer_census.md`. **~10 writers, three OUTSIDE `platform/orchestration/actions`** (the admin API, `cmd/webdesignport`, `cmd/content-data-recover`), so enforcement in the action layer would not see them |
| `framework_prompts_positive_voice` | VIZ constraints as template guidance | standing offer, not taken up |

**REFUSED, deliberately, and the next session should keep refusing it:** choosing what new carousels
should *be* — their character, how many, which suits which grid. That is taste for a marketing
homepage; the design-critique report is the input and this lane's scope is the editorial family.

---

## 6. CLOSED — do not reopen these

- **The infographic question: "untested, not broken".** An infographic needs a current `site_plan`
  AND a registered fact. **21** sites fleet-wide have both; **0** of them planned imagery since
  migration 718, and **0** of the 7 that did are capable. **Disjoint sets — so 718 has never run
  anywhere it could be answered, and no prompt edit is indicated by any of it.** If anyone forces the
  test, the canary is **`robot-hands.com`** (18 numeric facts including `series`, already runs the
  fact-resolved chart components, and it is this lane's own site so no other lane's approved copy is
  exposed). `agritec.uk` is the volume choice at 96 numeric facts but has no `series`.
- **The finetuning.uk imagery half.** The owner chose fleet-wide, so this lane writes nothing there.
- **`article-body` gaining its own image field.** That is migration **686** — applied and rolled back
  because 292 of 301 pages carrying it also carry a hero reading the same key. Do not let it attach
  itself to anything as a rider.

---

## 7. ⚠ WHAT THIS LANE GOT WRONG — because the next session inherits the habit, not just the conclusion

**Five measurements in one session were true, dated, `[MEASURED]` — and answered a NEIGHBOURING
question.** Two reached other lanes as assertions; one designed an experiment that then manufactured
agreement between two lanes. All are in `WRONG_CALLS.md` (2026-09-03 ×1, 2026-09-04 ×2) and one is a
footprinted `LANDMINES.md` entry.

| the claim | what it actually measured |
|---|---|
| `\| tail -12` on a psql list → "the site has twelve aspects, no evidence base" | 26 aspects; `ORDER BY 1` sorted `evidence_base` to the top and **tail ate it first** |
| "does the page have an image-capable section" → 351 of 360 | matched the **hero**, which is chrome |
| a verbatim prompt quote, from my own handoff | migration 718 had replaced that exact sentence **the day I read it** |
| "2 of 7 sites hold an `evidence_base`" | **aspect rows, not facts** — both held an EMPTY array, and this designed the test |
| "finetuning.uk is the only site where the question is askable" | it has 10 facts and **no `site_plans` row**, so it cannot hold section imagery at all |

**The pattern, named because it is consistent: a capability is a CONJUNCTION, and I kept measuring
whichever conjunct was easiest to query.** The checks that would have caught all five: **count before
you list** (a count cannot be truncated by a pipe); **run a control that must come out the other way**;
**re-read the live row, never your own handoff, when a quotation is about to be load-bearing**; and
**before accepting a test, name the result that is impossible in each arm.**

⚠ **One correction from the 114 lane, taken:** IMG-077's `unwired` count cannot see a page that
renders a *different* hero. I reported "4 unwired + 6 no_image_slot" on finetuning.uk without curling
what those pages actually serve. **Do not read an `unwired` count as "pages showing nothing".**

---

## 8. Identifiers

- **This session's code:** `1007be27d` (wiring), `3ba94508c` (council fixes). Council corr
  `cab931b1-8b45-461e-8a37-0dbdfa6aa928` — APPROVED, trailer `Council-Reviewed:`.
- **New writer stamp** `action:recompose_ancestors`; test file
  `platform/orchestration/actions/component_hierarchy_recompose_test.go` (M1–M6 in its header).
- **Papers written for other lanes:** `finetuning_uk_service/CONTRIB_2026-09-03_…` (`a85bcedea`);
  `bugfix_114_imagery_wiring/CONTRIB_2026-09-03_…_664_has_decayed_9_to_3_…` (`c816aa28a`);
  `framework_prompts_positive_voice/CONTRIB_2026-09-03b_…` (`4fb9b526f`) and
  `…CONTRIB_2026-09-04_…_my_section2_was_ALREADY_WRONG_…` (`c44f2b613`, + addenda `8b9aeb439`,
  `9689ba21e`); `SPEC_2026-09-04_carousel_component_constraints.md` (`c2cc6fb55`).
- **Still standing from 09-02 §8:** boxingonline site `d2aa5206-73bc-4707-a69c-2702c1eb9152` at
  `boxingonline.ugg2.com`; `article-body` `5835b2e1-50d7-4f20-8a9c-8da4d270ae3d` md5
  `002cbcd9cada6a37bf4a5158fd1e5f22`; planner definition `f263eaa1-61e1-446e-9410-648e12b7875b`.

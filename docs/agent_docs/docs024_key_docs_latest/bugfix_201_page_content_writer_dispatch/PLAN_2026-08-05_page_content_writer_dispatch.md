# PLAN — `bugs_open/201`, the direct `page-content-writer` dispatch

**Opened 2026-08-05.** Picked up from 201's own HANDOFF, which had done the diagnosis and left
an ordered next step. This lane did **not** re-derive the failure — §3 of that file asks that
it not be, and the request was reasonable.

## The decision, and the three shapes rejected

201 offered three candidates. A fourth appeared while measuring. **Candidate 2 was chosen:
re-point the three checks' `HandlerAgent` at `page-build-handler`.**

| candidate | what it was | why not |
|---|---|---|
| **1 — synthesise a `section_plan` at the dispatch seam** | teach `build-dispatch-loop` to build the rich plan object `087` describes and pass it in | **Wrong layer, and 201 said so before I looked.** It makes every *caller* responsible for reconstructing a structure the *handler* already knows how to build. `page-build-handler` gets its sections from `site_specs.site_plan` in one existing step; candidate 1 would reimplement that at the seam, per caller, in a shape (`{name, status, function, component:{…}, llm_fields}`) that is not one line of `input_mapping`. |
| **2 — re-route to `page-build-handler`** | ✅ **chosen** | Structurally immune to the cause (never reads the caller's spec for sections), empirically proven on already-built pages (32 completes), and **this exact migration has already been made twice** — `check_empty_sections.go` and `save_sections_claims_guard.go` both carry "(was `page-content-writer`)". These three are the un-migrated tail. |
| **3 — diagnose symptom 2 first** | the silent no-op | 201 §2 forbids it explicitly, and is right: a test against the current broken routing reaches `complete` while writing nothing, so fixing the trust-check first would make the *broken* route look repaired. |
| **4 — route to `section-editor`** (found by me, not in 201) | a real, live agent for scoped single-slot edits: `apply_section_edit`, contract `domain + edit_type + (page_component_id \| page_name+slot_name) + edit params` | **Genuinely attractive and genuinely wrong here.** A `literal_markdown` item carries `findings[].slot_name` and `field`, so the *target* maps cleanly. But its `fix` is an **instruction to a writer** ("Rewrite the affected fields WITHOUT markdown syntax… re-word so the words carry it"), not a replacement string — and `section-editor` has **no LLM step** (`ensure_site_record → spawn_deployer → load_edit_context → apply_edit → deploy_page → …`). It applies an edit somebody else composed. Choosing it would have meant inventing a compose-then-edit agent: a new shared mechanism, i.e. architecture scope, to fix a routing bug. |

## The trade-off accepted, stated plainly

`page-build-handler` repairs by rebuilding the page's sections through the writer. For a
one-field markdown defect that is **heavier than the ideal repair** — candidate 4's scalpel is
the better shape *in principle*. It is not available without building it, and the current
behaviour is a hard failure that repairs nothing at all. **A working blunt instrument beats a
sharp one that does not exist**, and `content_rewrite` — 19 completes — is essentially this
same case already running in production.

If someone later builds the compose-then-edit path, these three checks are the natural first
consumers, and this file is the argument for it.

## What this lane is NOT doing

- **Not repointing `site-work-orchestrator`'s `load_work_items` filter.** See NOTES; it invites
  double-dispatch. Recorded in the bug as a do-not-tidy.
- **Not touching `184`'s detection half** (201 §3) — the check is correct and finding real defects.
- **Not fixing symptom 2** — ordered second by 201 §2.
- **Not migrating the 14 existing DB rows.** The code change affects newly-filed items only.
  A re-arm for verification must set `handler_agent` too; that is a verification step, not a
  data migration, and it belongs with the post-roll check.

## Sequence

1. ✅ Re-point the three checks; correct the stale routing prose in the headers.
2. ✅ Prove the guard fails on a bad value (mutation), not merely passes on a good one.
3. ✅ Archive-of-HEAD build, so the shared branch cannot break on another session's WIP.
4. ✅ Council submitted (`71523705-07d1-4067-9c5d-af371ba84b89`), committed `37afbb847`.
5. ⬜ **Read the verdict.** Act on REVISE/REJECTED — the code is already on the branch.
6. ⬜ **Verify after the next roll, at the artefact** — one item, `content_data` must change.
7. ⬜ Then, and only then, symptom 2.

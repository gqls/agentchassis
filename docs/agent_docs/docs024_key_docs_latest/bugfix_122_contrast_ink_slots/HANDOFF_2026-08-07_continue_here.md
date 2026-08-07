# HANDOFF 2026-08-07 — bug 122, contrast / ink slots. START HERE.

Supersedes `HANDOFF_2026-08-06b_continue_here.md`. **That file's "next action 1" is now
DONE** and its §"What NOT to do" list still stands in full — it is not repeated here, so
read it there before touching code. Everything else below is new or changed.

## The one-paragraph state

The **engine half is live**: v1.0.1262 carries VIZ-014 and it is pod-proven on both
replicas (§1). It was rolled by another lane, not by us. The **config half — the part
that changes a visitor's page — is still NOT WRITTEN**: two migrations, fully specified
in the approved submission, and **the number `324` has since been taken by another
session**. Separately, picking up `bugs_open/212` reframed it twice over: its fix ranking
is refuted by arithmetic, its "unenforced contract" premise is wrong, and the real
blocker turned out to be a work-item contract defect now filed as **`bugs_open/213`**.

## 1. DONE — the engine is live, pod-proven [2026-08-07]

Both replicas of `agent-chassis`, image `docker.io/aqls/agent-chassis:v1.0.1262`:

| symbol | count | role |
|---|---|---|
| `buildLegibleInkDefaults` | 4 | the new emitter |
| `legibleInkFor` | 3 | " |
| `worstRatioAgainst` | 2 | " |
| `fillDarkSchemeSpecialisedSlots` | 4 | positive control |
| `zzzInventedControlXyz` | 0 | negative control |

Nothing downstream of a build records which commit it came from, so the symbols are the
only evidence — which is why both controls are in the table and why every replica was
checked. Do not re-derive this from the tag.

## 2. NEXT — migrations, unchanged in substance, one number changed

Both are still specified in `SUBMISSION_2026-08-06b_ink_slots_round2.json` (edit 7 and
the render-audit cadence). The tables of replacements and the five layouts are in
`HANDOFF_2026-08-06b_continue_here.md` §2–§3 — **use those, they are the approved plan.**

Changed since that handoff:

- **`324` IS TAKEN** — `docs/agent_docs/sql_for_agents/324_asset_deployer_passes_asset_id.sql`
  exists (untracked, another session). The old handoff's own warning — *"a number is not
  yours because you named a file"* — fired inside 24 hours. **Re-check for a free number
  at the moment you write the file**, and re-check again before `--apply`.
- Everything else in the non-negotiables still holds verbatim: on the ledger, needle
  gate with `position()` not `LIKE`, `\copy` backup, inline rollback, **every check
  `DO`/`RAISE` and induced once** (a verify block of bare `SELECT`s cannot stop the
  `COMMIT`), and **propagation enqueued as a separate step** — that last one is still the
  open `editquality` medium objection.

## 3. `bugs_open/212` — reframed. Read its §8 before acting on §1–§7

I filed 212 on 2026-08-06 and corrected it on 08-07. Three things changed:

- **The fix ranking is refuted.** On the motivating case the renderer's own
  contrast-checked value measures **1.71:1** against the component literal's **1.72:1**,
  and the muted slot **regresses 1.72 → 1.46**. §5's candidates 2 and 3 are a no-op and a
  slight worsening. `--color-primary-text` (`#0f0f0f`, already in gamesdesign's served
  CSS) gives **8.65:1** — that is candidate 5, and it is new to the file.
- **The class repair already exists.** `fix_forced_text_colours_action.go`'s
  `classifySectionPainting` derives what a template paints from its own CSS and
  `rewriteSectionDeclarationsInHTML` repoints `--section-*` to the on-colour family.
  `system-stats` matches its `paintPaletteBand` regex. **Nothing needs building** — the
  question is why it never ran, and §4 below answers that.
- **`buildSectionDefaults` emits nothing unless something is dark**
  (`color_util.go:185-187`) and its surface variant covers a hardcoded five-class list
  `.system-stats-section` is not in. Served blocks: gamesdesign 1, vonc 1, **idea.uk 0**.
  212's trap 4 was a guess; it is now confirmed at the source.

**Still open in 212, and still wanting a human:** the renderer knows exactly two grounds.
A component that paints its own has no correct token to ask for. Repointing components
one class at a time works; whether the renderer should learn about component-painted
grounds is an architecture question, not a bug fix.

## 4. NEW — `bugs_open/213`, and it is the actual blocker

gamesdesign's defect was **detected, described correctly, given a mechanical
`acceptance_test`, routed to the live fixer, and stamped `complete` in 3m17s with nothing
written** (proof: `content_components.updated_at` is 10.5 hours *earlier* than the item's
`created_at`). Not RFC_017's fail-open — nothing errored. Two producers file under one
`item_type` and the verifier implements only one of their predicates.

**7 of 9 `complete` items on that route are from the audit producer, all carrying an
`acceptance_test` nothing reads; every item that ever failed to close is from the other
producer, 6 of 6.** Do not read that as seven confirmed false-completes — read §4 of the
bug for what it does and does not establish. gamesdesign is the one confirmed at the
artefact.

**Where this bites this lane:** it is why 212 sat unrepaired while the queue reported the
site clean, and it will do the same to any repair we enqueue. Worth reading before
writing the propagation step in §2.

## 5. `090` — three UNVERIFIABLE in a row, and how to read one

Run 3, `b6ab22d6-e49c-4b55-a9d9-dd026532a595` (the renderer's grounds): **UNVERIFIABLE,
iteration-capped** — three `bundle` artifacts, no verdict artifact, no
`metadata->>'decision'`. **Not the stale-index cause this time**: `symbols_unreadable`
was 0 on iterations 2 and 3. Its final hypothesis independently reached §3's conclusion
and named `warnUnusablePrimary`'s blind spot; corroboration, not proof.

Run 4, **`84c3da66-06c0-41a5-94dc-21fbf71260f0`** (the 213 mechanism), was **still
`diagnosing` at handoff** — three `bundle` artifacts, no `decision`, item `updated_at`
frozen at 08:23:59Z. Same shape as run 3, but **it had not terminated, so it is not
recorded as UNVERIFIABLE** — that would be a claim about behaviour rather than the
behaviour. **Check it, and record the verdict in `NOTES_contrast_ink_slots.md` and in
`bugs_open/213` §8 either way, including if it is REFUTED.**

Its `symbols_unreadable` fell **3 → 1 → 0** across the three iterations, and that is not
the loop improving: **migration 332 repointed the code index to the live working branch
mid-run** (`code_symbols`: 5,754 symbols at `087_towards_multiple_domains`, indexed
08:31Z, verified independently — the single-`ref` row proves the old `086_experience_loop`
pin is gone). So this run straddles the fix. **A 090 filed before ~08:30Z on 2026-08-07
was reading a stale index and one filed after was not** — worth knowing before concluding
anything from an UNVERIFIABLE dated today, and it retires the standing "confirm from code,
not the index" workaround.

There is no `verdict` artifact kind and the outcome is not in `doc_notes`,
`site_work_items.spec` or `orchestration_states` — the query and how to read it are now
in `RUNBOOK_contrast_ink_slots.md` § "Reading a 090 verdict".

## 6. Verification target, unchanged

`BASELINE_2026-08-06_render_audit.txt` — 15 sites, 109 failures, complete. Grade **per
selector at the named ratio**, never by total; a falling count is content-dependent.
Expected close from the §2 migrations: 12 failures, plus `.stats-eyebrow` on vonc.
212's class is a **further ~24** and is not in that 12.

## 7. Advisory objections still open on the APPROVED verdict

Unchanged from `HANDOFF_2026-08-06b` §5 — six of them, `editquality` medium (the
propagation enqueue) being the one to act on. Approval means you may proceed; none were
withdrawn.

`Council-Reviewed: c4d9c841-3658-4742-85b5-961e062ecad2` still applies to further commits
in this lane.

## 8. Still no SUMMARY, still deliberately

The pages are unchanged, so the five headings would read "we measured, planned and built"
for the third time. The inflection to write one at is the first page measuring clean.

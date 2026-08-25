# 383 lane — CONTINUE HERE (2026-08-25). The fix is LIVE and PROVEN at the artefact. The lane is ONE observation away from closing.

**Read with:** `bugs_open/383_HANDOFF_2026-08-24_…md` (the bug; §13 is the post-roll verification)
· lane dir `docs/agent_docs/docs024_key_docs_latest/bugfix_283_component_instance_scope/`
· `SUMMARY_2026-08-24_occurrence_derivation.md` (the read-out) · `NOTES_…md` sessions 13–14.

---

## 1. What this bug is, in one paragraph

A component that appears twice on a page must give each copy a different HTML element id, or a
script aimed at one hits the other. The id carries a token, `InstanceToken(function, occurrence)`,
where **occurrence** = how many same-function sections precede this one in position order. Two
render paths see only ONE section at a time — `RenderComponentAction` (every build and every
`content_rewrite`) and the section editor — and both passed a **constant 0**, so every copy got
the same id. That is why repaired pages kept breaking: repairing worked, it did not hold.

## 2. STATUS: fixed, shipped, and verified at the artefact

- **Fix**: `364e80b7f` — build path counts its own loop's already-rendered items; editor counts
  stored predecessors position-exactly; constant 0 is the universal fallback. **No migration, no
  config key**, so it went live at the roll with no activation step.
- **Council `3fd0d026`: APPROVED round 1**, 4 advisories (all dispositioned, §9 of the bug file).
- **Retirement**: `9ba3293e7` deleted `BindSingleSectionInstanceToken`. One binder now.
- **LIVE on chassis `v1.0.1337`**, both replicas, from `4c996e1b5`. Proven by
  `git merge-base --is-ancestor`, **with a control** (today's HEAD is *not* an ancestor).
- **Working, measured at the artefact 2026-08-25**: `apis.uk/index.html` — six
  `illustrated-text-block` sections — had its rows rewritten at **11:27:27, after the 09:27
  roll, by the build path**, and they carry **six DISTINCT tokens**. The old code stamped
  occurrence 0 on all six. *That is the defect gone at its source, on the path that caused it.*
- **All three repaired pages serve distinct ids.** The §12 stored-vs-served divergence on
  `pricing-transparency.html` **closed itself** — a delivery lag, not a defect.

## 3. THE ONE THING LEFT — and it is an observation, not work

**Nothing is broken and nothing is queued for you to build.** One number needs explaining.

`apis.uk/index.html`'s tokens are `c-illustrated-text-block-2 … -7` = occurrences **1..6**. The
canonical walk (position 1 `hero`, positions 2–7 `illustrated-text-block`) must assign **0..5** —
bare, `-2` … `-6`. The build-path count is consistently **one high**.

**Every token is distinct, so there is no collision and nothing a visitor can see.** This is byte
drift in the errs-safe direction the design documents.

**A canonical rerender is already filed** (`page_rerender`, `reason: template_changed`,
`created_by='bugs_open/383'`, priority 60). Read its outcome at the artefact:

```bash
curl -s https://apis.uk/index.html | grep -o 'id="c-illustrated-text-block[^"]*"' | sort | uniq -c
```

- **bare + `-2`…`-6`** → the build's ready list carried one extra same-function item that produced
  no saved row (PLAN §A5 blind spot 2). It self-corrects, as designed. **→ CLOSE THE LANE.**
- **still `-2`…`-7`** → the canonical walk itself produces this, so the disagreement is NOT
  ready-list drift and lives in the derivation. **→ run `090` before touching code**, and re-read
  `PlacementFromLoopStep` against a LIVE orchestration's `loop_item_index` (0-based, verified
  2026-08-24 on a 5-item loop) rather than reasoning from the handler.

⚠ **Do not record blind spot 2 as the cause without that observation.** The obvious story —
"a 7th section was dropped" — is **REFUTED**: `pages.sections` plans exactly 6 and 6 rows exist.
The build's orchestration has been reaped, so the ready list cannot be read back.

## 4. Can we close the lane?

**Yes, on that one observation** — the closing bar is `fixed AND live`, and it is both. If the
rerender re-canonicalises, move `383` to `bugs_closed/` (⚠ `git mv` + a pathspec commit ships a
COPY — name **both** paths on the commit and verify with
`git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ | grep 383`).

**Two things do NOT block closure and should not be dragged into it:**

- **The generic-fallback question is the OWNER's, not this lane's.** The council's
  `bug_historian` seat objected that patching the two known callers leaves the constant-0 default
  able to bite a future caller, and that *"a LANDMINES entry is documentation, not a guard"*. It
  is argued both ways in §7 and §9. A `LANDMINES.md` entry is filed and verifier-dispatched
  (`8cb37bfb`), footprinted on `render_component` / `process_sections_loop`. The peer lane notes
  it is the same shape as **RFC_050** (whether the render seam may refuse) and suggests folding
  both into one change if RFC_050 lands on arming — that seems right, and it is a decision, not a
  defect.
- **`bugs_closed/283`** (literal ids) is a *different* defect and stays closed. Resolve by SLUG.

## 5. Traps this lane paid for — read before touching anything here

- **A corpus count stays GREEN while this bug happens.** The same token twice is still two `c-`
  ids. Count **distinct** tokens per page, or fetch the page.
- **`complete` is not proof.** On 2026-08-24 all three repairs were `complete` with correct
  stored bytes and one page still served the duplicate.
- **A work item's `spec.reason` is load-bearing.** `page-rerender` routes on an **allow-list of
  five**; anything else (including an invented one, including none) goes to assemble-only, which
  re-ships stored bytes and **completes successfully having repaired nothing**. Use
  `template_changed` — the only one of the five with no Go branch keyed on it.
- **Ask the DB for an interval, never eyeball its timestamp against your shell.** The DB is UTC,
  this box is BST. I nearly escalated a healthy dispatch trigger as dead on a one-hour illusion.
- **When you retire a symbol, the prose is the load-bearing edit, not the regex.** In one session
  the retired binder was still named as live advice in `pattern-check.py`'s finding text, in
  CLC-014's body, and in my own CLC-031 entry. Grep `.go` **and** `.py` **and** the register.
- **An append to a shared append-only doc is not finished until it is committed** — two of mine
  left under other lanes' commit messages within minutes.

## 6. Where everything lives

| thing | where |
|---|---|
| the bug, evidence, verification, council dispositions | `bugs_open/383_HANDOFF_2026-08-24_…md` (§13 = post-roll) |
| the code | `component_instance_occurrence.go`, `datahelpers/loop_keys.go`, call sites in `v3_site_actions.go` / `section_editor_actions.go` |
| the cross-package contract pin | `platform/orchestration/loop_item_contract_parity_test.go` |
| what is callable, for another workstream | concept register **CLC-031** (+ CLC-014 corrected) |
| the plan, with its superseded Half A marked | `PLAN_2026-08-24_occurrence_derivation_and_empty_id_detector.md` |
| the read-out for a non-specialist | `SUMMARY_2026-08-24_occurrence_derivation.md` |
| the repair SQL, with its reason-routing warning | `SQL_2026-08-24_repair_duplicated_instance_tokens.sql` |
| my own wrong calls | `WRONG_CALLS.md` entries 9 and 10 |

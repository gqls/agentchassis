# 383 lane — CONTINUE HERE (2026-08-25). The fix is SHIPPED and WELL-TESTED. It is **NOT yet proven at the artefact**, and that is the one thing left.

> **⚠ CORRECTED 2026-08-25, later the same day.** An earlier version of this file said the fix was
> "LIVE and PROVEN at the artefact" and that the lane was one observation from closing. **The proof
> was false** and is retracted in bug-file §14 / `WRONG_CALLS` 11: I read `updated_at::time(0)`,
> which discards the DATE, and cited rows written **2026-08-24 11:27** as having been written
> after the **2026-08-25 09:27** roll. They predate it by ~22 hours. Everything below is the
> corrected position.

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
- ~~Working, measured at the artefact~~ **RETRACTED — see the banner above.** `apis.uk`'s rows
  predate the roll and were almost certainly **hand-written** by that lane (its own register
  entry, CLC-030, says it wrote `rendered_html` directly), which is also why they carry six
  distinct tokens the old code could not have produced.
- **All three repaired pages serve distinct ids** — but they were repaired by the **canonical
  whole-page walk, which was never broken.** That proves the *repair*, not the *fix*. The §12
  stored-vs-served divergence on `pricing-transparency.html` closed itself — a delivery lag,
  not a defect.

## 3. THE ONE THING LEFT — observe a PER-SECTION render on a multi-instance page

**Nothing is broken and nothing is queued for you to build.** What is missing is an observation.

The fix changes two paths: `RenderComponentAction` (every build and `content_rewrite`) and the
section editor. **Neither has run on a multi-instance page since the roll.** Until one does and
is counted at the served page, "the fix works in production" is untested — the unit and
cross-package tests are mutation-proven and the binary is right, but that is a different claim.

Any ONE of these settles it:

- **`apis.uk/index.html`** holds `build_status='needs_rebuild'` and is supposed to rebuild itself,
  but has not since **2026-08-24 11:37**. Worth asking why before relying on it.
- **A `content_rewrite` on any of the 30 multi-instance pages.** This is also the **stickiness**
  test — the one that actually matters, because it is the operation that undid the repairs on
  2026-08-23 — and it is still unrun.
- **A section edit** on the second instance of a multi-instance page (exercises the DB branch).

Then, at the served page:

```bash
curl -s <page> | grep -o 'id="c-[^"]*"' | sort | uniq -c   # every token must appear ONCE
```

⚠ **Count DISTINCT tokens.** A count of `c-`-prefixed ids reads green while this bug is happening.

⚠ **A canonical `page_rerender` does NOT test this.** That path was never broken. Only a
per-section render exercises what was fixed.

A canonical rerender of `apis.uk/index.html` is filed (`created_by='bugs_open/383'`,
`reason: template_changed`) — still `triaged`, `attempt_count=2`, unclaimed as of 09:38 UTC. It
now answers a much smaller question than the earlier version of this file claimed: whether the
canonical walk re-canonicalises hand-written tokens. **It is not the test that matters.**

## 4. Can we close the lane?

**Not yet.** The bar is `fixed AND live`. It is fixed and it is shipped, but "live" on this
estate means the behaviour is observed at the artefact, and the observation in §3 has not been
made. **Do not close it on the strength of the three repaired pages** — those were repaired by a
path that was never broken. Once a per-section render has been observed producing distinct
tokens, move `383` to `bugs_closed/` (⚠ `git mv` + a pathspec commit ships a
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
